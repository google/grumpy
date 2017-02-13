from __future__ import absolute_import, division, print_function, unicode_literals
import sys, pythonparser.source, pythonparser.lexer, pythonparser.parser, pythonparser.diagnostic

def parse_buffer(buffer, mode="exec", flags=[], version=None, engine=None):
    """
    Like :meth:`parse`, but accepts a :class:`source.Buffer` instead of
    source and filename, and returns comments as well.

    :see: :meth:`parse`
    :return: (:class:`ast.AST`, list of :class:`source.Comment`)
        Abstract syntax tree and comments
    """

    if version is None:
        version = sys.version_info[0:2]

    if engine is None:
        engine = pythonparser.diagnostic.Engine()

    lexer = pythonparser.lexer.Lexer(buffer, version, engine)
    if mode in ("single", "eval"):
        lexer.interactive = True

    parser = pythonparser.parser.Parser(lexer, version, engine)
    parser.add_flags(flags)

    if mode == "exec":
        return parser.file_input(), lexer.comments
    elif mode == "single":
        return parser.single_input(), lexer.comments
    elif mode == "eval":
        return parser.eval_input(), lexer.comments

def parse(source, filename="<unknown>", mode="exec",
          flags=[], version=None, engine=None):
    """
    Parse a string into an abstract syntax tree.
    This is the replacement for the built-in :meth:`..ast.parse`.

    :param source: (string) Source code in the correct encoding
    :param filename: (string) Filename of the source (used in diagnostics)
    :param mode: (string) Execution mode. Pass ``"exec"`` to parse a module,
        ``"single"`` to parse a single (interactive) statement,
        and ``"eval"`` to parse an expression. In the last two cases,
        ``source`` must be terminated with an empty line
        (i.e. end with ``"\\n\\n"``).
    :param flags: (list of string) Future flags.
        Equivalent to ``from __future__ import <flags>``.
    :param version: (2-tuple of int) Major and minor version of Python
        syntax to recognize, ``sys.version_info[0:2]`` by default.
    :param engine: (:class:`diagnostic.Engine`) Diagnostic engine,
        a fresh one is created by default
    :return: (:class:`ast.AST`) Abstract syntax tree
    :raise: :class:`diagnostic.Error`
        if the source code is not well-formed
    """
    ast, comments = parse_buffer(pythonparser.source.Buffer(source, filename),
                                 mode, flags, version, engine)
    return ast

