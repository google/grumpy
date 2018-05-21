"""
The :mod:`lexer` module concerns itself with tokenizing Python source.
"""

from __future__ import absolute_import, division, print_function, unicode_literals
from . import source, diagnostic
import re
import unicodedata
import sys

if sys.version_info[0] == 3:
    unichr = chr
    byte = lambda x: bytes([x])
else:
    byte = chr

class Token:
    """
    The :class:`Token` encapsulates a single lexer token and its location
    in the source code.

    :ivar loc: (:class:`pythonparser.source.Range`) token location
    :ivar kind: (string) token kind
    :ivar value: token value; None or a kind-specific class
    """
    def __init__(self, loc, kind, value=None):
        self.loc, self.kind, self.value = loc, kind, value

    def __repr__(self):
        return "Token(%s, \"%s\", %s)" % (repr(self.loc), self.kind, repr(self.value))

class Lexer:
    """
    The :class:`Lexer` class extracts tokens and comments from
    a :class:`pythonparser.source.Buffer`.

    :class:`Lexer` is an iterable.

    :ivar version: (tuple of (*major*, *minor*))
        the version of Python, determining the grammar used
    :ivar source_buffer: (:class:`pythonparser.source.Buffer`)
        the source buffer
    :ivar diagnostic_engine: (:class:`pythonparser.diagnostic.Engine`)
        the diagnostic engine
    :ivar offset: (integer) character offset into ``source_buffer``
        indicating where the next token will be recognized
    :ivar interactive: (boolean) whether a completely empty line
        should generate a NEWLINE token, for use in REPLs
    """

    _reserved_2_6 = frozenset([
        "!=", "%", "%=", "&", "&=", "(", ")", "*", "**", "**=", "*=", "+", "+=",
        ",", "-", "-=", ".", "/", "//", "//=", "/=", ":", ";", "<", "<<", "<<=",
        "<=", "<>", "=", "==", ">", ">=", ">>", ">>=", "@", "[", "]", "^", "^=", "`",
        "and", "as", "assert", "break", "class", "continue", "def", "del", "elif",
        "else", "except", "exec", "finally", "for", "from", "global", "if", "import",
        "in", "is", "lambda", "not", "or", "pass", "print", "raise", "return", "try",
        "while", "with", "yield", "{", "|", "|=", "}", "~"
    ])

    _reserved_3_0 = _reserved_2_6 \
        - set(["<>", "`", "exec", "print"]) \
        | set(["->", "...", "False", "None", "nonlocal", "True"])

    _reserved_3_1 = _reserved_3_0 \
        | set(["<>"])

    _reserved_3_5 = _reserved_3_1 \
        | set(["@", "@="])

    _reserved = {
        (2, 6): _reserved_2_6,
        (2, 7): _reserved_2_6,
        (3, 0): _reserved_3_0,
        (3, 1): _reserved_3_1,
        (3, 2): _reserved_3_1,
        (3, 3): _reserved_3_1,
        (3, 4): _reserved_3_1,
        (3, 5): _reserved_3_5,
    }
    """
    A map from a tuple (*major*, *minor*) corresponding to Python version to
    :class:`frozenset`\s of keywords.
    """

    _string_prefixes_3_1 = frozenset(["", "r", "b", "br"])
    _string_prefixes_3_3 = frozenset(["", "r", "u", "b", "br", "rb"])

    # holy mother of god why
    _string_prefixes = {
        (2, 6): frozenset(["", "r", "u", "ur"]),
        (2, 7): frozenset(["", "r", "u", "ur", "b", "br"]),
        (3, 0): frozenset(["", "r", "b"]),
        (3, 1): _string_prefixes_3_1,
        (3, 2): _string_prefixes_3_1,
        (3, 3): _string_prefixes_3_3,
        (3, 4): _string_prefixes_3_3,
        (3, 5): _string_prefixes_3_3,
    }
    """
    A map from a tuple (*major*, *minor*) corresponding to Python version to
    :class:`frozenset`\s of string prefixes.
    """

    def __init__(self, source_buffer, version, diagnostic_engine, interactive=False):
        self.source_buffer = source_buffer
        self.version = version
        self.diagnostic_engine = diagnostic_engine
        self.interactive = interactive
        self.print_function = False
        self.unicode_literals = self.version >= (3, 0)

        self.offset = 0
        self.new_line = True
        self.indent = [(0, source.Range(source_buffer, 0, 0), "")]
        self.comments = []
        self.queue = []
        self.parentheses = []
        self.curly_braces = []
        self.square_braces = []

        try:
            reserved = self._reserved[version]
        except KeyError:
            raise NotImplementedError("pythonparser.lexer.Lexer cannot lex Python %s" % str(version))

        # Sort for the regexp to obey longest-match rule.
        re_reserved  = sorted(reserved, reverse=True, key=len)
        re_keywords  = "|".join([kw for kw in re_reserved if kw.isalnum()])
        re_operators = "|".join([re.escape(op) for op in re_reserved if not op.isalnum()])

        # Python 3.0 uses ID_Start, >3.0 uses XID_Start
        if self.version == (3, 0):
            id_xid = ""
        else:
            id_xid = "X"

        # To speed things up on CPython, we use the re module to generate a DFA
        # from our token set and execute it in C. Every result yielded by
        # iterating this regular expression has exactly one non-empty group
        # that would correspond to a e.g. lex scanner branch.
        # The only thing left to Python code is then to select one from this
        # small set of groups, which is much faster than dissecting the strings.
        #
        # A lexer has to obey longest-match rule, but a regular expression does not.
        # Therefore, the cases in it are carefully sorted so that the longest
        # ones come up first. The exception is the identifier case, which would
        # otherwise grab all keywords; it is made to work by making it impossible
        # for the keyword case to match a word prefix, and ordering it before
        # the identifier case.
        self._lex_token_re = re.compile(r"""
        [ \t\f]* # initial whitespace
        ( # 1
            (\\)? # ?2 line continuation
            ([\n]|[\r][\n]|[\r]) # 3 newline
        |   (\#.*) # 4 comment
        |   ( # 5 floating point or complex literal
                (?: [0-9]* \.  [0-9]+
                |   [0-9]+ \.?
                ) [eE] [+-]? [0-9]+
            |   [0-9]* \. [0-9]+
            |   [0-9]+ \.
            ) ([jJ])? # ?6 complex suffix
        |   ([0-9]+) [jJ] # 7 complex literal
        |   (?: # integer literal
                ( [1-9]   [0-9]* )       # 8 dec
            |     0[oO] ( [0-7]+ )       # 9 oct
            |     0[xX] ( [0-9A-Fa-f]+ ) # 10 hex
            |     0[bB] ( [01]+ )        # 11 bin
            |   ( [0-9]   [0-9]* )       # 12 bare oct
            )
            ([Ll])?                      # 13 long option
        |   ([BbUu]?[Rr]?) # ?14 string literal options
            (?: # string literal start
                # 15, 16, 17 long string
                (""\"|''') ((?: \\?[\n] | \\. | . )*?) (\15)
                # 18, 19, 20 short string
            |   ("   |'  ) ((?: \\ [\n] | \\. | . )*?) (\18)
                # 21 unterminated
            |   (""\"|'''|"|')
            )
        |   ((?:{keywords})\b|{operators}) # 22 keywords and operators
        |   ([A-Za-z_][A-Za-z0-9_]*\b) # 23 identifier
        |   (\p{{{id_xid}ID_Start}}\p{{{id_xid}ID_Continue}}*) # 24 Unicode identifier
        |   ($) # 25 end-of-file
        )
        """.format(keywords=re_keywords, operators=re_operators,
                   id_xid=id_xid), re.VERBOSE|re.UNICODE)

    # These are identical for all lexer instances.
    _lex_escape_pattern = r"""
    \\(?:
        ([\n\\'"abfnrtv]) # 1 single-char
    |   ([0-7]{1,3})      # 2 oct
    |   x([0-9A-Fa-f]{2}) # 3 hex
    )
    """
    _lex_escape_re = re.compile(_lex_escape_pattern.encode(), re.VERBOSE)

    _lex_escape_unicode_re = re.compile(_lex_escape_pattern + r"""
    | \\(?:
        u([0-9A-Fa-f]{4}) # 4 unicode-16
    |   U([0-9A-Fa-f]{8}) # 5 unicode-32
    |   N\{(.+?)\}        # 6 unicode-name
    )
    """, re.VERBOSE)

    def next(self, eof_token=False):
        """
        Returns token at ``offset`` as a :class:`Token` and advances ``offset``
        to point past the end of the token, where the token has:

        - *range* which is a :class:`pythonparser.source.Range` that includes
          the token but not surrounding whitespace,
        - *kind* which is a string containing one of Python keywords or operators,
          ``newline``, ``float``, ``int``, ``complex``, ``strbegin``,
          ``strdata``, ``strend``, ``ident``, ``indent``, ``dedent`` or ``eof``
          (if ``eof_token`` is True).
        - *value* which is the flags as lowercase string if *kind* is ``strbegin``,
          the string contents if *kind* is ``strdata``,
          the numeric value if *kind* is ``float``, ``int`` or ``complex``,
          the identifier if *kind* is ``ident`` and ``None`` in any other case.

        :param eof_token: if true, will return a token with kind ``eof``
            when the input is exhausted; if false, will raise ``StopIteration``.
        """
        if len(self.queue) == 0:
            self._refill(eof_token)

        return self.queue.pop(0)

    def peek(self, eof_token=False):
        """Same as :meth:`next`, except the token is not dequeued."""
        if len(self.queue) == 0:
            self._refill(eof_token)

        return self.queue[-1]

    # We need separate next and _refill because lexing can sometimes
    # generate several tokens, e.g. INDENT
    def _refill(self, eof_token):
        if self.offset == len(self.source_buffer.source):
            range = source.Range(self.source_buffer, self.offset, self.offset)

            if not self.new_line:
                self.new_line = True
                self.queue.append(Token(range, "newline"))
                return

            for i in self.indent[1:]:
                self.indent.pop(-1)
                self.queue.append(Token(range, "dedent"))

            if eof_token:
                self.queue.append(Token(range, "eof"))
            elif len(self.queue) == 0:
                raise StopIteration

            return

        match = self._lex_token_re.match(self.source_buffer.source, self.offset)
        if match is None:
            diag = diagnostic.Diagnostic(
                "fatal", "unexpected {character}",
                {"character": repr(self.source_buffer.source[self.offset]).lstrip("u")},
                source.Range(self.source_buffer, self.offset, self.offset + 1))
            self.diagnostic_engine.process(diag)

        # Should we emit indent/dedent?
        if self.new_line and \
                match.group(3) is None and \
                match.group(4) is None: # not a blank line
            whitespace = match.string[match.start(0):match.start(1)]
            level = len(whitespace.expandtabs())
            range = source.Range(self.source_buffer, match.start(1), match.start(1))
            if level > self.indent[-1][0]:
                self.indent.append((level, range, whitespace))
                self.queue.append(Token(range, "indent"))
            elif level < self.indent[-1][0]:
                exact = False
                while level <= self.indent[-1][0]:
                    if level == self.indent[-1][0] or self.indent[-1][0] == 0:
                        exact = True
                        break
                    self.indent.pop(-1)
                    self.queue.append(Token(range, "dedent"))
                if not exact:
                    note = diagnostic.Diagnostic(
                        "note", "expected to match level here", {},
                        self.indent[-1][1])
                    error = diagnostic.Diagnostic(
                        "fatal", "inconsistent indentation", {},
                        range, notes=[note])
                    self.diagnostic_engine.process(error)
            elif whitespace != self.indent[-1][2] and self.version >= (3, 0):
                error = diagnostic.Diagnostic(
                    "error", "inconsistent use of tabs and spaces in indentation", {},
                    range)
                self.diagnostic_engine.process(error)

        # Prepare for next token.
        self.offset = match.end(0)

        tok_range = source.Range(self.source_buffer, *match.span(1))
        if match.group(3) is not None: # newline
            if len(self.parentheses) + len(self.square_braces) + len(self.curly_braces) > 0:
                # 2.1.6 Implicit line joining
                return self._refill(eof_token)
            if match.group(2) is not None:
                # 2.1.5. Explicit line joining
                return self._refill(eof_token)
            if self.new_line and not \
                    (self.interactive and match.group(0) == match.group(3)): # REPL terminator
                # 2.1.7. Blank lines
                return self._refill(eof_token)

            self.new_line = True
            self.queue.append(Token(tok_range, "newline"))
            return

        if match.group(4) is not None: # comment
            self.comments.append(source.Comment(tok_range, match.group(4)))
            return self._refill(eof_token)

        # Lexing non-whitespace now.
        self.new_line = False

        if sys.version_info > (3,) or not match.group(13):
            int_type = int
        else:
            int_type = long

        if match.group(5) is not None: # floating point or complex literal
            if match.group(6) is None:
                self.queue.append(Token(tok_range, "float", float(match.group(5))))
            else:
                self.queue.append(Token(tok_range, "complex", float(match.group(5)) * 1j))

        elif match.group(7) is not None: # complex literal
            self.queue.append(Token(tok_range, "complex", int(match.group(7)) * 1j))

        elif match.group(8) is not None: # integer literal, dec
            literal = match.group(8)
            self._check_long_literal(tok_range, match.group(1))
            self.queue.append(Token(tok_range, "int", int_type(literal)))

        elif match.group(9) is not None: # integer literal, oct
            literal = match.group(9)
            self._check_long_literal(tok_range, match.group(1))
            self.queue.append(Token(tok_range, "int", int_type(literal, 8)))

        elif match.group(10) is not None: # integer literal, hex
            literal = match.group(10)
            self._check_long_literal(tok_range, match.group(1))
            self.queue.append(Token(tok_range, "int", int_type(literal, 16)))

        elif match.group(11) is not None: # integer literal, bin
            literal = match.group(11)
            self._check_long_literal(tok_range, match.group(1))
            self.queue.append(Token(tok_range, "int", int_type(literal, 2)))

        elif match.group(12) is not None: # integer literal, bare oct
            literal = match.group(12)
            if len(literal) > 1 and self.version >= (3, 0):
                error = diagnostic.Diagnostic(
                    "error", "in Python 3, decimal literals must not start with a zero", {},
                    source.Range(self.source_buffer, tok_range.begin_pos, tok_range.begin_pos + 1))
                self.diagnostic_engine.process(error)
            self.queue.append(Token(tok_range, "int", int(literal, 8)))

        elif match.group(15) is not None: # long string literal
            self._string_literal(
                options=match.group(14), begin_span=(match.start(14), match.end(15)),
                data=match.group(16), data_span=match.span(16),
                end_span=match.span(17))

        elif match.group(18) is not None: # short string literal
            self._string_literal(
                options=match.group(14), begin_span=(match.start(14), match.end(18)),
                data=match.group(19), data_span=match.span(19),
                end_span=match.span(20))

        elif match.group(21) is not None: # unterminated string
            error = diagnostic.Diagnostic(
                "fatal", "unterminated string", {},
                tok_range)
            self.diagnostic_engine.process(error)

        elif match.group(22) is not None: # keywords and operators
            kwop = match.group(22)
            self._match_pair_delim(tok_range, kwop)
            if kwop == "print" and self.print_function:
                self.queue.append(Token(tok_range, "ident", "print"))
            else:
                self.queue.append(Token(tok_range, kwop))

        elif match.group(23) is not None: # identifier
            self.queue.append(Token(tok_range, "ident", match.group(23)))

        elif match.group(24) is not None: # Unicode identifier
            if self.version < (3, 0):
                error = diagnostic.Diagnostic(
                    "error", "in Python 2, Unicode identifiers are not allowed", {},
                    tok_range)
                self.diagnostic_engine.process(error)
            self.queue.append(Token(tok_range, "ident", match.group(24)))

        elif match.group(25) is not None: # end-of-file
            # Reuse the EOF logic
            return self._refill(eof_token)

        else:
            assert False

    def _string_literal(self, options, begin_span, data, data_span, end_span):
        options = options.lower()
        begin_range = source.Range(self.source_buffer, *begin_span)
        data_range = source.Range(self.source_buffer, *data_span)

        if options not in self._string_prefixes[self.version]:
            error = diagnostic.Diagnostic(
                "error", "string prefix '{prefix}' is not available in Python {major}.{minor}",
                {"prefix": options, "major": self.version[0], "minor": self.version[1]},
                begin_range)
            self.diagnostic_engine.process(error)

        self.queue.append(Token(begin_range, "strbegin", options))
        self.queue.append(Token(data_range,
                          "strdata", self._replace_escape(data_range, options, data)))
        self.queue.append(Token(source.Range(self.source_buffer, *end_span),
                          "strend"))

    def _replace_escape(self, range, mode, value):
        is_raw     = ("r" in mode)
        is_unicode = "u" in mode or ("b" not in mode and self.unicode_literals)

        if not is_unicode:
            value = value.encode(self.source_buffer.encoding)
            if is_raw:
                return value
            return self._replace_escape_bytes(value)

        if is_raw:
            return value

        return self._replace_escape_unicode(range, value)

    def _replace_escape_unicode(self, range, value):
        chunks = []
        offset = 0
        while offset < len(value):
            match = self._lex_escape_unicode_re.search(value, offset)
            if match is None:
                # Append the remaining of the string
                chunks.append(value[offset:])
                break

            # Append the part of string before match
            chunks.append(value[offset:match.start()])
            offset = match.end()

            # Process the escape
            if match.group(1) is not None: # single-char
                chr = match.group(1)
                if chr == "\n":
                    pass
                elif chr == "\\" or chr == "'" or chr == "\"":
                    chunks.append(chr)
                elif chr == "a":
                    chunks.append("\a")
                elif chr == "b":
                    chunks.append("\b")
                elif chr == "f":
                    chunks.append("\f")
                elif chr == "n":
                    chunks.append("\n")
                elif chr == "r":
                    chunks.append("\r")
                elif chr == "t":
                    chunks.append("\t")
                elif chr == "v":
                    chunks.append("\v")
            elif match.group(2) is not None: # oct
                chunks.append(unichr(int(match.group(2), 8)))
            elif match.group(3) is not None: # hex
                chunks.append(unichr(int(match.group(3), 16)))
            elif match.group(4) is not None: # unicode-16
                chunks.append(unichr(int(match.group(4), 16)))
            elif match.group(5) is not None: # unicode-32
                try:
                    chunks.append(unichr(int(match.group(5), 16)))
                except ValueError:
                    error = diagnostic.Diagnostic(
                        "error", "unicode character out of range", {},
                        source.Range(self.source_buffer,
                                     range.begin_pos + match.start(0),
                                     range.begin_pos + match.end(0)))
                    self.diagnostic_engine.process(error)
            elif match.group(6) is not None: # unicode-name
                try:
                    chunks.append(unicodedata.lookup(match.group(6)))
                except KeyError:
                    error = diagnostic.Diagnostic(
                        "error", "unknown unicode character name", {},
                        source.Range(self.source_buffer,
                                     range.begin_pos + match.start(0),
                                     range.begin_pos + match.end(0)))
                    self.diagnostic_engine.process(error)

        return "".join(chunks)

    def _replace_escape_bytes(self, value):
        chunks = []
        offset = 0
        while offset < len(value):
            match = self._lex_escape_re.search(value, offset)
            if match is None:
                # Append the remaining of the string
                chunks.append(value[offset:])
                break

            # Append the part of string before match
            chunks.append(value[offset:match.start()])
            offset = match.end()

            # Process the escape
            if match.group(1) is not None: # single-char
                chr = match.group(1)
                if chr == b"\n":
                    pass
                elif chr == b"\\" or chr == b"'" or chr == b"\"":
                    chunks.append(chr)
                elif chr == b"a":
                    chunks.append(b"\a")
                elif chr == b"b":
                    chunks.append(b"\b")
                elif chr == b"f":
                    chunks.append(b"\f")
                elif chr == b"n":
                    chunks.append(b"\n")
                elif chr == b"r":
                    chunks.append(b"\r")
                elif chr == b"t":
                    chunks.append(b"\t")
                elif chr == b"v":
                    chunks.append(b"\v")
            elif match.group(2) is not None: # oct
                chunks.append(byte(int(match.group(2), 8)))
            elif match.group(3) is not None: # hex
                chunks.append(byte(int(match.group(3), 16)))

        return b"".join(chunks)

    def _check_long_literal(self, range, literal):
        if literal[-1] in "lL" and self.version >= (3, 0):
            error = diagnostic.Diagnostic(
                "error", "in Python 3, long integer literals were removed", {},
                source.Range(self.source_buffer, range.end_pos - 1, range.end_pos))
            self.diagnostic_engine.process(error)

    def _match_pair_delim(self, range, kwop):
        if kwop == "(":
            self.parentheses.append(range)
        elif kwop == "[":
            self.square_braces.append(range)
        elif kwop == "{":
            self.curly_braces.append(range)
        elif kwop == ")":
            self._check_innermost_pair_delim(range, "(")
            self.parentheses.pop()
        elif kwop == "]":
            self._check_innermost_pair_delim(range, "[")
            self.square_braces.pop()
        elif kwop == "}":
            self._check_innermost_pair_delim(range, "{")
            self.curly_braces.pop()

    def _check_innermost_pair_delim(self, range, expected):
        ranges = []
        if len(self.parentheses) > 0:
            ranges.append(("(", self.parentheses[-1]))
        if len(self.square_braces) > 0:
            ranges.append(("[", self.square_braces[-1]))
        if len(self.curly_braces) > 0:
            ranges.append(("{", self.curly_braces[-1]))

        ranges.sort(key=lambda k: k[1].begin_pos)
        if any(ranges):
            compl_kind, compl_range = ranges[-1]
            if compl_kind != expected:
                note = diagnostic.Diagnostic(
                    "note", "'{delimiter}' opened here",
                    {"delimiter": compl_kind},
                    compl_range)
                error = diagnostic.Diagnostic(
                    "fatal", "mismatched '{delimiter}'",
                    {"delimiter": range.source()},
                    range, notes=[note])
                self.diagnostic_engine.process(error)
        else:
            error = diagnostic.Diagnostic(
                "fatal", "mismatched '{delimiter}'",
                {"delimiter": range.source()},
                range)
            self.diagnostic_engine.process(error)

    def __iter__(self):
        return self

    def __next__(self):
        return self.next()
