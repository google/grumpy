# encoding:utf-8

"""
The :mod:`parser` module concerns itself with parsing Python source.
"""

from __future__ import absolute_import, division, print_function, unicode_literals
from functools import reduce
from . import source, diagnostic, lexer, ast

# A few notes about our approach to parsing:
#
# Python uses an LL(1) parser generator. It's a bit weird, because
# the usual reason to choose LL(1) is to make a handwritten parser
# possible, however Python's grammar is formulated in a way that
# is much more easily recognized if you make an FSM rather than
# the usual "if accept(token)..." ladder. So in a way it is
# the worst of both worlds.
#
# We don't use a parser generator because we want to have an unified
# grammar for all Python versions, and also have grammar coverage
# analysis and nice error recovery. To make the grammar compact,
# we use combinators to compose it from predefined fragments,
# such as "sequence" or "alternation" or "Kleene star". This easily
# gives us one token of lookahead in most cases, but e.g. not
# in the following one:
#
#     argument: test | test '=' test
#
# There are two issues with this. First, in an alternation, the first
# variant will be tried (and accepted) earlier. Second, if we reverse
# them, by the point it is clear ``'='`` will not be accepted, ``test``
# has already been consumed.
#
# The way we fix this is by reordering rules so that longest match
# comes first, and adding backtracking on alternations (as well as
# plus and star, since those have a hidden alternation inside).
#
# While backtracking can in principle make asymptotical complexity
# worse, it never makes parsing syntactically correct code supralinear
# with Python's LL(1) grammar, and we could not come up with any
# pathological incorrect input as well.

# Coverage data
_all_rules = []
_all_stmts = {}

# Generic LL parsing combinators
class Unmatched:
    pass

unmatched = Unmatched()

def llrule(loc, expected, cases=1):
    if loc is None:
        def decorator(rule):
            rule.expected = expected
            return rule
    else:
        def decorator(inner_rule):
            if cases == 1:
                def rule(*args, **kwargs):
                    result = inner_rule(*args, **kwargs)
                    if result is not unmatched:
                        rule.covered[0] = True
                    return result
            else:
                rule = inner_rule

            rule.loc, rule.expected, rule.covered = \
                loc, expected, [False] * cases
            _all_rules.append(rule)

            return rule
    return decorator

def action(inner_rule, loc=None):
    """
    A decorator returning a function that first runs ``inner_rule`` and then, if its
    return value is not None, maps that value using ``mapper``.

    If the value being mapped is a tuple, it is expanded into multiple arguments.

    Similar to attaching semantic actions to rules in traditional parser generators.
    """
    def decorator(mapper):
        @llrule(loc, inner_rule.expected)
        def outer_rule(parser):
            result = inner_rule(parser)
            if result is unmatched:
                return result
            if isinstance(result, tuple):
                return mapper(parser, *result)
            else:
                return mapper(parser, result)
        return outer_rule
    return decorator

def Eps(value=None, loc=None):
    """A rule that accepts no tokens (epsilon) and returns ``value``."""
    @llrule(loc, lambda parser: [])
    def rule(parser):
        return value
    return rule

def Tok(kind, loc=None):
    """A rule that accepts a token of kind ``kind`` and returns it, or returns None."""
    @llrule(loc, lambda parser: [kind])
    def rule(parser):
        return parser._accept(kind)
    return rule

def Loc(kind, loc=None):
    """A rule that accepts a token of kind ``kind`` and returns its location, or returns None."""
    @llrule(loc, lambda parser: [kind])
    def rule(parser):
        result = parser._accept(kind)
        if result is unmatched:
            return result
        return result.loc
    return rule

def Rule(name, loc=None):
    """A proxy for a rule called ``name`` which may not be yet defined."""
    @llrule(loc, lambda parser: getattr(parser, name).expected(parser))
    def rule(parser):
        return getattr(parser, name)()
    return rule

def Expect(inner_rule, loc=None):
    """A rule that executes ``inner_rule`` and emits a diagnostic error if it returns None."""
    @llrule(loc, inner_rule.expected)
    def rule(parser):
        result = inner_rule(parser)
        if result is unmatched:
            expected = reduce(list.__add__, [rule.expected(parser) for rule in parser._errrules])
            expected = list(sorted(set(expected)))

            if len(expected) > 1:
                expected = " or ".join([", ".join(expected[0:-1]), expected[-1]])
            elif len(expected) == 1:
                expected = expected[0]
            else:
                expected = "(impossible)"

            error_tok = parser._tokens[parser._errindex]
            error = diagnostic.Diagnostic(
                "fatal", "unexpected {actual}: expected {expected}",
                {"actual": error_tok.kind, "expected": expected},
                error_tok.loc)
            parser.diagnostic_engine.process(error)
        return result
    return rule

def Seq(first_rule, *rest_of_rules, **kwargs):
    """
    A rule that accepts a sequence of tokens satisfying ``rules`` and returns a tuple
    containing their return values, or None if the first rule was not satisfied.
    """
    @llrule(kwargs.get("loc", None), first_rule.expected)
    def rule(parser):
        result = first_rule(parser)
        if result is unmatched:
            return result

        results = [result]
        for rule in rest_of_rules:
            result = rule(parser)
            if result is unmatched:
                return result
            results.append(result)
        return tuple(results)
    return rule

def SeqN(n, *inner_rules, **kwargs):
    """
    A rule that accepts a sequence of tokens satisfying ``rules`` and returns
    the value returned by rule number ``n``, or None if the first rule was not satisfied.
    """
    @action(Seq(*inner_rules), loc=kwargs.get("loc", None))
    def rule(parser, *values):
        return values[n]
    return rule

def Alt(*inner_rules, **kwargs):
    """
    A rule that expects a sequence of tokens satisfying one of ``rules`` in sequence
    (a rule is satisfied when it returns anything but None) and returns the return
    value of that rule, or None if no rules were satisfied.
    """
    loc = kwargs.get("loc", None)
    expected = lambda parser: reduce(list.__add__, map(lambda x: x.expected(parser), inner_rules))
    if loc is not None:
        @llrule(loc, expected, cases=len(inner_rules))
        def rule(parser):
            data = parser._save()
            for idx, inner_rule in enumerate(inner_rules):
                result = inner_rule(parser)
                if result is unmatched:
                    parser._restore(data, rule=inner_rule)
                else:
                    rule.covered[idx] = True
                    return result
            return unmatched
    else:
        @llrule(loc, expected, cases=len(inner_rules))
        def rule(parser):
            data = parser._save()
            for inner_rule in inner_rules:
                result = inner_rule(parser)
                if result is unmatched:
                    parser._restore(data, rule=inner_rule)
                else:
                    return result
            return unmatched
    return rule

def Opt(inner_rule, loc=None):
    """Shorthand for ``Alt(inner_rule, Eps())``"""
    return Alt(inner_rule, Eps(), loc=loc)

def Star(inner_rule, loc=None):
    """
    A rule that accepts a sequence of tokens satisfying ``inner_rule`` zero or more times,
    and returns the returned values in a :class:`list`.
    """
    @llrule(loc, lambda parser: [])
    def rule(parser):
        results = []
        while True:
            data = parser._save()
            result = inner_rule(parser)
            if result is unmatched:
                parser._restore(data, rule=inner_rule)
                return results
            results.append(result)
    return rule

def Plus(inner_rule, loc=None):
    """
    A rule that accepts a sequence of tokens satisfying ``inner_rule`` one or more times,
    and returns the returned values in a :class:`list`.
    """
    @llrule(loc, inner_rule.expected)
    def rule(parser):
        result = inner_rule(parser)
        if result is unmatched:
            return result

        results = [result]
        while True:
            data = parser._save()
            result = inner_rule(parser)
            if result is unmatched:
                parser._restore(data, rule=inner_rule)
                return results
            results.append(result)
    return rule

class commalist(list):
    __slots__ = ("trailing_comma",)

def List(inner_rule, separator_tok, trailing, leading=True, loc=None):
    if not trailing:
        @action(Seq(inner_rule, Star(SeqN(1, Tok(separator_tok), inner_rule))), loc=loc)
        def outer_rule(parser, first, rest):
            return [first] + rest
        return outer_rule
    else:
        # A rule like this: stmt (';' stmt)* [';']
        # This doesn't yield itself to combinators above, because disambiguating
        # another iteration of the Kleene star and the trailing separator
        # requires two lookahead tokens (naively).
        separator_rule = Tok(separator_tok)
        @llrule(loc, inner_rule.expected)
        def rule(parser):
            results = commalist()

            if leading:
                result = inner_rule(parser)
                if result is unmatched:
                    return result
                else:
                    results.append(result)

            while True:
                result = separator_rule(parser)
                if result is unmatched:
                    results.trailing_comma = None
                    return results

                result_1 = inner_rule(parser)
                if result_1 is unmatched:
                    results.trailing_comma = result
                    return results
                else:
                    results.append(result_1)
        return rule

# Python AST specific parser combinators
def Newline(loc=None):
    """A rule that accepts token of kind ``newline`` and returns an empty list."""
    @llrule(loc, lambda parser: ["newline"])
    def rule(parser):
        result = parser._accept("newline")
        if result is unmatched:
            return result
        return []
    return rule

def Oper(klass, *kinds, **kwargs):
    """
    A rule that accepts a sequence of tokens of kinds ``kinds`` and returns
    an instance of ``klass`` with ``loc`` encompassing the entire sequence
    or None if the first token is not of ``kinds[0]``.
    """
    @action(Seq(*map(Loc, kinds)), loc=kwargs.get("loc", None))
    def rule(parser, *tokens):
        return klass(loc=tokens[0].join(tokens[-1]))
    return rule

def BinOper(expr_rulename, op_rule, node=ast.BinOp, loc=None):
    @action(Seq(Rule(expr_rulename), Star(Seq(op_rule, Rule(expr_rulename)))), loc=loc)
    def rule(parser, lhs, trailers):
        for (op, rhs) in trailers:
            lhs = node(left=lhs, op=op, right=rhs,
                       loc=lhs.loc.join(rhs.loc))
        return lhs
    return rule

def BeginEnd(begin_tok, inner_rule, end_tok, empty=None, loc=None):
    @action(Seq(Loc(begin_tok), inner_rule, Loc(end_tok)), loc=loc)
    def rule(parser, begin_loc, node, end_loc):
        if node is None:
            node = empty(parser)

        # Collection nodes don't have loc yet. If a node has loc at this
        # point, it means it's an expression passed in parentheses.
        if node.loc is None and type(node) in [
                ast.List, ast.ListComp,
                ast.Dict, ast.DictComp,
                ast.Set, ast.SetComp,
                ast.GeneratorExp,
                ast.Tuple, ast.Repr,
                ast.Call, ast.Subscript,
                ast.arguments]:
            node.begin_loc, node.end_loc, node.loc = \
                begin_loc, end_loc, begin_loc.join(end_loc)
        return node
    return rule

class Parser:

    # Generic LL parsing methods
    def __init__(self, lexer, version, diagnostic_engine):
        self._init_version(version)
        self.diagnostic_engine = diagnostic_engine

        self.lexer     = lexer
        self._tokens   = []
        self._index    = -1
        self._errindex = -1
        self._errrules = []
        self._advance()

    def _save(self):
        return self._index

    def _restore(self, data, rule):
        self._index = data
        self._token = self._tokens[self._index]

        if self._index > self._errindex:
            # We have advanced since last error
            self._errindex = self._index
            self._errrules = [rule]
        elif self._index == self._errindex:
            # We're at the same place as last error
            self._errrules.append(rule)
        else:
            # We've backtracked far and are now just failing the
            # whole parse
            pass

    def _advance(self):
        self._index += 1
        if self._index == len(self._tokens):
            self._tokens.append(self.lexer.next(eof_token=True))
        self._token = self._tokens[self._index]

    def _accept(self, expected_kind):
        if self._token.kind == expected_kind:
            result = self._token
            self._advance()
            return result
        return unmatched

    # Python-specific methods
    def _init_version(self, version):
        if version in ((2, 6), (2, 7)):
            if version == (2, 6):
                self.with_stmt       = self.with_stmt__26
                self.atom_6          = self.atom_6__26
            else:
                self.with_stmt       = self.with_stmt__27
                self.atom_6          = self.atom_6__27
            self.except_clause_1 = self.except_clause_1__26
            self.classdef        = self.classdef__26
            self.subscript       = self.subscript__26
            self.raise_stmt      = self.raise_stmt__26
            self.comp_if         = self.comp_if__26
            self.atom            = self.atom__26
            self.funcdef         = self.funcdef__26
            self.parameters      = self.parameters__26
            self.varargslist     = self.varargslist__26
            self.comparison_1    = self.comparison_1__26
            self.exprlist_1      = self.exprlist_1__26
            self.testlist_comp_1 = self.testlist_comp_1__26
            self.expr_stmt_1     = self.expr_stmt_1__26
            self.yield_expr      = self.yield_expr__26
            return
        elif version in ((3, 0), (3, 1), (3, 2), (3, 3), (3, 4), (3, 5)):
            if version == (3, 0):
                self.with_stmt       = self.with_stmt__26 # lol
            else:
                self.with_stmt       = self.with_stmt__27
            self.except_clause_1 = self.except_clause_1__30
            self.classdef        = self.classdef__30
            self.subscript       = self.subscript__30
            self.raise_stmt      = self.raise_stmt__30
            self.comp_if         = self.comp_if__30
            self.atom            = self.atom__30
            self.funcdef         = self.funcdef__30
            self.parameters      = self.parameters__30
            if version < (3, 2):
                self.varargslist     = self.varargslist__30
                self.typedargslist   = self.typedargslist__30
                self.comparison_1    = self.comparison_1__30
                self.star_expr       = self.star_expr__30
                self.exprlist_1      = self.exprlist_1__30
                self.testlist_comp_1 = self.testlist_comp_1__26
                self.expr_stmt_1     = self.expr_stmt_1__26
            else:
                self.varargslist     = self.varargslist__32
                self.typedargslist   = self.typedargslist__32
                self.comparison_1    = self.comparison_1__32
                self.star_expr       = self.star_expr__32
                self.exprlist_1      = self.exprlist_1__32
                self.testlist_comp_1 = self.testlist_comp_1__32
                self.expr_stmt_1     = self.expr_stmt_1__32
            if version < (3, 3):
                self.yield_expr      = self.yield_expr__26
            else:
                self.yield_expr      = self.yield_expr__33
            return

        raise NotImplementedError("pythonparser.parser.Parser cannot parse Python %s" %
                                  str(version))

    def _arguments(self, args=None, defaults=None, kwonlyargs=None, kw_defaults=None,
                   vararg=None, kwarg=None,
                   star_loc=None, dstar_loc=None, begin_loc=None, end_loc=None,
                   equals_locs=None, kw_equals_locs=None, loc=None):
        if args is None:
            args = []
        if defaults is None:
            defaults = []
        if kwonlyargs is None:
            kwonlyargs = []
        if kw_defaults is None:
            kw_defaults = []
        if equals_locs is None:
            equals_locs = []
        if kw_equals_locs is None:
            kw_equals_locs = []
        return ast.arguments(args=args, defaults=defaults,
                             kwonlyargs=kwonlyargs, kw_defaults=kw_defaults,
                             vararg=vararg, kwarg=kwarg,
                             star_loc=star_loc, dstar_loc=dstar_loc,
                             begin_loc=begin_loc, end_loc=end_loc,
                             equals_locs=equals_locs, kw_equals_locs=kw_equals_locs,
                             loc=loc)

    def _arg(self, tok, colon_loc=None, annotation=None):
        loc = tok.loc
        if annotation:
            loc = loc.join(annotation.loc)
        return ast.arg(arg=tok.value, annotation=annotation,
                       arg_loc=tok.loc, colon_loc=colon_loc, loc=loc)

    def _empty_arglist(self):
        return ast.Call(args=[], keywords=[], starargs=None, kwargs=None,
                        star_loc=None, dstar_loc=None, loc=None)

    def _wrap_tuple(self, elts):
        assert len(elts) > 0
        if len(elts) > 1:
            return ast.Tuple(ctx=None, elts=elts,
                             loc=elts[0].loc.join(elts[-1].loc), begin_loc=None, end_loc=None)
        else:
            return elts[0]

    def _assignable(self, node, is_delete=False):
        if isinstance(node, ast.Name) or isinstance(node, ast.Subscript) or \
                isinstance(node, ast.Attribute) or isinstance(node, ast.Starred):
            return node
        elif (isinstance(node, ast.List) or isinstance(node, ast.Tuple)) and \
                any(node.elts):
            node.elts = [self._assignable(elt, is_delete) for elt in node.elts]
            return node
        else:
            if is_delete:
                error = diagnostic.Diagnostic(
                    "fatal", "cannot delete this expression", {}, node.loc)
            else:
                error = diagnostic.Diagnostic(
                    "fatal", "cannot assign to this expression", {}, node.loc)
            self.diagnostic_engine.process(error)

    def add_flags(self, flags):
        if "print_function" in flags:
            self.lexer.print_function = True
        if "unicode_literals" in flags:
            self.lexer.unicode_literals = True

    # Grammar
    @action(Expect(Alt(Newline(),
                       Rule("simple_stmt"),
                       SeqN(0, Rule("compound_stmt"), Newline()))))
    def single_input(self, body):
        """single_input: NEWLINE | simple_stmt | compound_stmt NEWLINE"""
        loc = None
        if body != []:
            loc = body[0].loc
        return ast.Interactive(body=body, loc=loc)

    @action(Expect(SeqN(0, Star(Alt(Newline(), Rule("stmt"))), Tok("eof"))))
    def file_input(parser, body):
        """file_input: (NEWLINE | stmt)* ENDMARKER"""
        body = reduce(list.__add__, body, [])
        loc = None
        if body != []:
            loc = body[0].loc
        return ast.Module(body=body, loc=loc)

    @action(Expect(SeqN(0, Rule("testlist"), Star(Tok("newline")), Tok("eof"))))
    def eval_input(self, expr):
        """eval_input: testlist NEWLINE* ENDMARKER"""
        return ast.Expression(body=[expr], loc=expr.loc)

    @action(Seq(Loc("@"), List(Tok("ident"), ".", trailing=False),
                Opt(BeginEnd("(", Opt(Rule("arglist")), ")",
                             empty=_empty_arglist)),
                Loc("newline")))
    def decorator(self, at_loc, idents, call_opt, newline_loc):
        """decorator: '@' dotted_name [ '(' [arglist] ')' ] NEWLINE"""
        root = idents[0]
        dec_loc = root.loc
        expr = ast.Name(id=root.value, ctx=None, loc=root.loc)
        for ident in idents[1:]:
          dot_loc = ident.loc.begin()
          dot_loc.begin_pos -= 1
          dec_loc = dec_loc.join(ident.loc)
          expr = ast.Attribute(value=expr, attr=ident.value, ctx=None,
                               loc=expr.loc.join(ident.loc),
                               attr_loc=ident.loc, dot_loc=dot_loc)

        if call_opt:
            call_opt.func = expr
            call_opt.loc = dec_loc.join(call_opt.loc)
            expr = call_opt
        return at_loc, expr

    decorators = Plus(Rule("decorator"))
    """decorators: decorator+"""

    @action(Seq(Rule("decorators"), Alt(Rule("classdef"), Rule("funcdef"))))
    def decorated(self, decorators, classfuncdef):
        """decorated: decorators (classdef | funcdef)"""
        classfuncdef.at_locs = list(map(lambda x: x[0], decorators))
        classfuncdef.decorator_list = list(map(lambda x: x[1], decorators))
        classfuncdef.loc = classfuncdef.loc.join(decorators[0][0])
        return classfuncdef

    @action(Seq(Loc("def"), Tok("ident"), Rule("parameters"), Loc(":"), Rule("suite")))
    def funcdef__26(self, def_loc, ident_tok, args, colon_loc, suite):
        """(2.6, 2.7) funcdef: 'def' NAME parameters ':' suite"""
        return ast.FunctionDef(name=ident_tok.value, args=args, returns=None,
                               body=suite, decorator_list=[],
                               at_locs=[], keyword_loc=def_loc, name_loc=ident_tok.loc,
                               colon_loc=colon_loc, arrow_loc=None,
                               loc=def_loc.join(suite[-1].loc))

    @action(Seq(Loc("def"), Tok("ident"), Rule("parameters"),
                Opt(Seq(Loc("->"), Rule("test"))),
                Loc(":"), Rule("suite")))
    def funcdef__30(self, def_loc, ident_tok, args, returns_opt, colon_loc, suite):
        """(3.0-) funcdef: 'def' NAME parameters ['->' test] ':' suite"""
        arrow_loc = returns = None
        if returns_opt:
            arrow_loc, returns = returns_opt
        return ast.FunctionDef(name=ident_tok.value, args=args, returns=returns,
                               body=suite, decorator_list=[],
                               at_locs=[], keyword_loc=def_loc, name_loc=ident_tok.loc,
                               colon_loc=colon_loc, arrow_loc=arrow_loc,
                               loc=def_loc.join(suite[-1].loc))

    parameters__26 = BeginEnd("(", Opt(Rule("varargslist")), ")", empty=_arguments)
    """(2.6, 2.7) parameters: '(' [varargslist] ')'"""

    parameters__30 = BeginEnd("(", Opt(Rule("typedargslist")), ")", empty=_arguments)
    """(3.0) parameters: '(' [typedargslist] ')'"""

    varargslist__26_1 = Seq(Rule("fpdef"), Opt(Seq(Loc("="), Rule("test"))))

    @action(Seq(Loc("**"), Tok("ident")))
    def varargslist__26_2(self, dstar_loc, kwarg_tok):
        return self._arguments(kwarg=self._arg(kwarg_tok),
                               dstar_loc=dstar_loc, loc=dstar_loc.join(kwarg_tok.loc))

    @action(Seq(Loc("*"), Tok("ident"),
                Opt(Seq(Tok(","), Loc("**"), Tok("ident")))))
    def varargslist__26_3(self, star_loc, vararg_tok, kwarg_opt):
        dstar_loc = kwarg = None
        loc = star_loc.join(vararg_tok.loc)
        vararg = self._arg(vararg_tok)
        if kwarg_opt:
            _, dstar_loc, kwarg_tok = kwarg_opt
            kwarg = self._arg(kwarg_tok)
            loc = star_loc.join(kwarg_tok.loc)
        return self._arguments(vararg=vararg, kwarg=kwarg,
                               star_loc=star_loc, dstar_loc=dstar_loc, loc=loc)

    @action(Eps(value=()))
    def varargslist__26_4(self):
        return self._arguments()

    @action(Alt(Seq(Star(SeqN(0, varargslist__26_1, Tok(","))),
                    Alt(varargslist__26_2, varargslist__26_3)),
                Seq(List(varargslist__26_1, ",", trailing=True),
                    varargslist__26_4)))
    def varargslist__26(self, fparams, args):
        """
        (2.6, 2.7)
        varargslist: ((fpdef ['=' test] ',')*
                      ('*' NAME [',' '**' NAME] | '**' NAME) |
                      fpdef ['=' test] (',' fpdef ['=' test])* [','])
        """
        for fparam, default_opt in fparams:
            if default_opt:
                equals_loc, default = default_opt
                args.equals_locs.append(equals_loc)
                args.defaults.append(default)
            elif len(args.defaults) > 0:
                error = diagnostic.Diagnostic(
                    "fatal", "non-default argument follows default argument", {},
                    fparam.loc, [args.args[-1].loc.join(args.defaults[-1].loc)])
                self.diagnostic_engine.process(error)

            args.args.append(fparam)

        def fparam_loc(fparam, default_opt):
            if default_opt:
                equals_loc, default = default_opt
                return fparam.loc.join(default.loc)
            else:
                return fparam.loc

        if args.loc is None:
            args.loc = fparam_loc(*fparams[0]).join(fparam_loc(*fparams[-1]))
        elif len(fparams) > 0:
            args.loc = args.loc.join(fparam_loc(*fparams[0]))

        return args

    @action(Tok("ident"))
    def fpdef_1(self, ident_tok):
        return ast.arg(arg=ident_tok.value, annotation=None,
                       arg_loc=ident_tok.loc, colon_loc=None,
                       loc=ident_tok.loc)

    fpdef = Alt(fpdef_1, BeginEnd("(", Rule("fplist"), ")",
                                  empty=lambda self: ast.Tuple(elts=[], ctx=None, loc=None)))
    """fpdef: NAME | '(' fplist ')'"""

    def _argslist(fpdef_rule, old_style=False):
        argslist_1 = Seq(fpdef_rule, Opt(Seq(Loc("="), Rule("test"))))

        @action(Seq(Loc("**"), Tok("ident")))
        def argslist_2(self, dstar_loc, kwarg_tok):
            return self._arguments(kwarg=self._arg(kwarg_tok),
                                   dstar_loc=dstar_loc, loc=dstar_loc.join(kwarg_tok.loc))

        @action(Seq(Loc("*"), Tok("ident"),
                    Star(SeqN(1, Tok(","), argslist_1)),
                    Opt(Seq(Tok(","), Loc("**"), Tok("ident")))))
        def argslist_3(self, star_loc, vararg_tok, fparams, kwarg_opt):
            dstar_loc = kwarg = None
            loc = star_loc.join(vararg_tok.loc)
            vararg = self._arg(vararg_tok)
            if kwarg_opt:
                _, dstar_loc, kwarg_tok = kwarg_opt
                kwarg = self._arg(kwarg_tok)
                loc = star_loc.join(kwarg_tok.loc)
            kwonlyargs, kw_defaults, kw_equals_locs = [], [], []
            for fparam, default_opt in fparams:
                if default_opt:
                    equals_loc, default = default_opt
                    kw_equals_locs.append(equals_loc)
                    kw_defaults.append(default)
                else:
                    kw_defaults.append(None)
                kwonlyargs.append(fparam)
            if any(kw_defaults):
                loc = loc.join(kw_defaults[-1].loc)
            elif any(kwonlyargs):
                loc = loc.join(kwonlyargs[-1].loc)
            return self._arguments(vararg=vararg, kwarg=kwarg,
                                   kwonlyargs=kwonlyargs, kw_defaults=kw_defaults,
                                   star_loc=star_loc, dstar_loc=dstar_loc,
                                   kw_equals_locs=kw_equals_locs, loc=loc)

        argslist_4 = Alt(argslist_2, argslist_3)

        @action(Eps(value=()))
        def argslist_5(self):
            return self._arguments()

        if old_style:
            argslist = Alt(Seq(Star(SeqN(0, argslist_1, Tok(","))),
                               argslist_4),
                           Seq(List(argslist_1, ",", trailing=True),
                               argslist_5))
        else:
            argslist = Alt(Seq(Eps(value=[]), argslist_4),
                           Seq(List(argslist_1, ",", trailing=False),
                               Alt(SeqN(1, Tok(","), Alt(argslist_4, argslist_5)),
                                   argslist_5)))

        def argslist_action(self, fparams, args):
            for fparam, default_opt in fparams:
                if default_opt:
                    equals_loc, default = default_opt
                    args.equals_locs.append(equals_loc)
                    args.defaults.append(default)
                elif len(args.defaults) > 0:
                    error = diagnostic.Diagnostic(
                        "fatal", "non-default argument follows default argument", {},
                        fparam.loc, [args.args[-1].loc.join(args.defaults[-1].loc)])
                    self.diagnostic_engine.process(error)

                args.args.append(fparam)

            def fparam_loc(fparam, default_opt):
                if default_opt:
                    equals_loc, default = default_opt
                    return fparam.loc.join(default.loc)
                else:
                    return fparam.loc

            if args.loc is None:
                args.loc = fparam_loc(*fparams[0]).join(fparam_loc(*fparams[-1]))
            elif len(fparams) > 0:
                args.loc = args.loc.join(fparam_loc(*fparams[0]))

            return args

        return action(argslist)(argslist_action)

    typedargslist__30 = _argslist(Rule("tfpdef"), old_style=True)
    """
    (3.0, 3.1)
    typedargslist: ((tfpdef ['=' test] ',')*
                    ('*' [tfpdef] (',' tfpdef ['=' test])* [',' '**' tfpdef] | '**' tfpdef)
                    | tfpdef ['=' test] (',' tfpdef ['=' test])* [','])
    """

    typedargslist__32 = _argslist(Rule("tfpdef"))
    """
    (3.2-)
    typedargslist: (tfpdef ['=' test] (',' tfpdef ['=' test])* [','
           ['*' [tfpdef] (',' tfpdef ['=' test])* [',' '**' tfpdef] | '**' tfpdef]]
         |  '*' [tfpdef] (',' tfpdef ['=' test])* [',' '**' tfpdef] | '**' tfpdef)
    """

    varargslist__30 = _argslist(Rule("vfpdef"), old_style=True)
    """
    (3.0, 3.1)
    varargslist: ((vfpdef ['=' test] ',')*
                  ('*' [vfpdef] (',' vfpdef ['=' test])*  [',' '**' vfpdef] | '**' vfpdef)
                  | vfpdef ['=' test] (',' vfpdef ['=' test])* [','])
    """

    varargslist__32 = _argslist(Rule("vfpdef"))
    """
    (3.2-)
    varargslist: (vfpdef ['=' test] (',' vfpdef ['=' test])* [','
           ['*' [vfpdef] (',' vfpdef ['=' test])* [',' '**' vfpdef] | '**' vfpdef]]
         |  '*' [vfpdef] (',' vfpdef ['=' test])* [',' '**' vfpdef] | '**' vfpdef)
    """

    @action(Seq(Tok("ident"), Opt(Seq(Loc(":"), Rule("test")))))
    def tfpdef(self, ident_tok, annotation_opt):
        """(3.0-) tfpdef: NAME [':' test]"""
        if annotation_opt:
            colon_loc, annotation = annotation_opt
            return self._arg(ident_tok, colon_loc, annotation)
        return self._arg(ident_tok)

    vfpdef = fpdef_1
    """(3.0-) vfpdef: NAME"""

    @action(List(Rule("fpdef"), ",", trailing=True))
    def fplist(self, elts):
        """fplist: fpdef (',' fpdef)* [',']"""
        return ast.Tuple(elts=elts, ctx=None, loc=None)

    stmt = Alt(Rule("simple_stmt"), Rule("compound_stmt"))
    """stmt: simple_stmt | compound_stmt"""

    simple_stmt = SeqN(0, List(Rule("small_stmt"), ";", trailing=True), Tok("newline"))
    """simple_stmt: small_stmt (';' small_stmt)* [';'] NEWLINE"""

    small_stmt = Alt(Rule("expr_stmt"), Rule("print_stmt"),  Rule("del_stmt"),
                     Rule("pass_stmt"), Rule("flow_stmt"), Rule("import_stmt"),
                     Rule("global_stmt"), Rule("nonlocal_stmt"), Rule("exec_stmt"),
                     Rule("assert_stmt"))
    """
    (2.6, 2.7)
    small_stmt: (expr_stmt | print_stmt  | del_stmt | pass_stmt | flow_stmt |
                 import_stmt | global_stmt | exec_stmt | assert_stmt)
    (3.0-)
    small_stmt: (expr_stmt | del_stmt | pass_stmt | flow_stmt |
                 import_stmt | global_stmt | nonlocal_stmt | assert_stmt)
    """

    expr_stmt_1__26 = Rule("testlist")
    expr_stmt_1__32 = Rule("testlist_star_expr")

    @action(Seq(Rule("augassign"), Alt(Rule("yield_expr"), Rule("testlist"))))
    def expr_stmt_2(self, augassign, rhs_expr):
        return ast.AugAssign(op=augassign, value=rhs_expr)

    @action(Star(Seq(Loc("="), Alt(Rule("yield_expr"), Rule("expr_stmt_1")))))
    def expr_stmt_3(self, seq):
        if len(seq) > 0:
            return ast.Assign(targets=list(map(lambda x: x[1], seq[:-1])), value=seq[-1][1],
                              op_locs=list(map(lambda x: x[0], seq)))
        else:
            return None

    @action(Seq(Rule("expr_stmt_1"), Alt(expr_stmt_2, expr_stmt_3)))
    def expr_stmt(self, lhs, rhs):
        """
        (2.6, 2.7, 3.0, 3.1)
        expr_stmt: testlist (augassign (yield_expr|testlist) |
                             ('=' (yield_expr|testlist))*)
        (3.2-)
        expr_stmt: testlist_star_expr (augassign (yield_expr|testlist) |
                             ('=' (yield_expr|testlist_star_expr))*)
        """
        if isinstance(rhs, ast.AugAssign):
            if isinstance(lhs, ast.Tuple) or isinstance(lhs, ast.List):
                error = diagnostic.Diagnostic(
                    "fatal", "illegal expression for augmented assignment", {},
                    rhs.op.loc, [lhs.loc])
                self.diagnostic_engine.process(error)
            else:
                rhs.target = self._assignable(lhs)
                rhs.loc = rhs.target.loc.join(rhs.value.loc)
                return rhs
        elif rhs is not None:
            rhs.targets = list(map(self._assignable, [lhs] + rhs.targets))
            rhs.loc = lhs.loc.join(rhs.value.loc)
            return rhs
        else:
            return ast.Expr(value=lhs, loc=lhs.loc)

    testlist_star_expr = action(
        List(Alt(Rule("test"), Rule("star_expr")), ",", trailing=True)) \
        (_wrap_tuple)
    """(3.2-) testlist_star_expr: (test|star_expr) (',' (test|star_expr))* [',']"""

    augassign = Alt(Oper(ast.Add, "+="), Oper(ast.Sub, "-="), Oper(ast.MatMult, "@="),
                    Oper(ast.Mult, "*="), Oper(ast.Div, "/="), Oper(ast.Mod, "%="),
                    Oper(ast.BitAnd, "&="), Oper(ast.BitOr, "|="), Oper(ast.BitXor, "^="),
                    Oper(ast.LShift, "<<="), Oper(ast.RShift, ">>="),
                    Oper(ast.Pow, "**="), Oper(ast.FloorDiv, "//="))
    """augassign: ('+=' | '-=' | '*=' | '/=' | '%=' | '&=' | '|=' | '^=' |
                   '<<=' | '>>=' | '**=' | '//=')"""

    @action(List(Rule("test"), ",", trailing=True))
    def print_stmt_1(self, values):
        nl, loc = True, values[-1].loc
        if values.trailing_comma:
            nl, loc = False, values.trailing_comma.loc
        return ast.Print(dest=None, values=values, nl=nl,
                         dest_loc=None, loc=loc)

    @action(Seq(Loc(">>"), Rule("test"), Tok(","), List(Rule("test"), ",", trailing=True)))
    def print_stmt_2(self, dest_loc, dest, comma_tok, values):
        nl, loc = True, values[-1].loc
        if values.trailing_comma:
            nl, loc = False, values.trailing_comma.loc
        return ast.Print(dest=dest, values=values, nl=nl,
                         dest_loc=dest_loc, loc=loc)

    @action(Eps())
    def print_stmt_3(self, eps):
        return ast.Print(dest=None, values=[], nl=True,
                         dest_loc=None, loc=None)

    @action(Seq(Loc("print"), Alt(print_stmt_1, print_stmt_2, print_stmt_3)))
    def print_stmt(self, print_loc, stmt):
        """
        (2.6-2.7)
        print_stmt: 'print' ( [ test (',' test)* [','] ] |
                              '>>' test [ (',' test)+ [','] ] )
        """
        stmt.keyword_loc = print_loc
        if stmt.loc is None:
            stmt.loc = print_loc
        else:
            stmt.loc = print_loc.join(stmt.loc)
        return stmt

    @action(Seq(Loc("del"), List(Rule("expr"), ",", trailing=True)))
    def del_stmt(self, stmt_loc, exprs):
        # Python uses exprlist here, but does *not* obey the usual
        # tuple-wrapping semantics, so we embed the rule directly.
        """del_stmt: 'del' exprlist"""
        return ast.Delete(targets=[self._assignable(expr, is_delete=True) for expr in exprs],
                          loc=stmt_loc.join(exprs[-1].loc), keyword_loc=stmt_loc)

    @action(Loc("pass"))
    def pass_stmt(self, stmt_loc):
        """pass_stmt: 'pass'"""
        return ast.Pass(loc=stmt_loc, keyword_loc=stmt_loc)

    flow_stmt = Alt(Rule("break_stmt"), Rule("continue_stmt"), Rule("return_stmt"),
                    Rule("raise_stmt"), Rule("yield_stmt"))
    """flow_stmt: break_stmt | continue_stmt | return_stmt | raise_stmt | yield_stmt"""

    @action(Loc("break"))
    def break_stmt(self, stmt_loc):
        """break_stmt: 'break'"""
        return ast.Break(loc=stmt_loc, keyword_loc=stmt_loc)

    @action(Loc("continue"))
    def continue_stmt(self, stmt_loc):
        """continue_stmt: 'continue'"""
        return ast.Continue(loc=stmt_loc, keyword_loc=stmt_loc)

    @action(Seq(Loc("return"), Opt(Rule("testlist"))))
    def return_stmt(self, stmt_loc, values):
        """return_stmt: 'return' [testlist]"""
        loc = stmt_loc
        if values:
            loc = loc.join(values.loc)
        return ast.Return(value=values,
                          loc=loc, keyword_loc=stmt_loc)

    @action(Rule("yield_expr"))
    def yield_stmt(self, expr):
        """yield_stmt: yield_expr"""
        return ast.Expr(value=expr, loc=expr.loc)

    @action(Seq(Loc("raise"), Opt(Seq(Rule("test"),
                                      Opt(Seq(Tok(","), Rule("test"),
                                              Opt(SeqN(1, Tok(","), Rule("test")))))))))
    def raise_stmt__26(self, raise_loc, type_opt):
        """(2.6, 2.7) raise_stmt: 'raise' [test [',' test [',' test]]]"""
        type_ = inst = tback = None
        loc = raise_loc
        if type_opt:
            type_, inst_opt = type_opt
            loc = loc.join(type_.loc)
            if inst_opt:
                _, inst, tback = inst_opt
                loc = loc.join(inst.loc)
                if tback:
                    loc = loc.join(tback.loc)
        return ast.Raise(exc=type_, inst=inst, tback=tback, cause=None,
                         keyword_loc=raise_loc, from_loc=None, loc=loc)

    @action(Seq(Loc("raise"), Opt(Seq(Rule("test"), Opt(Seq(Loc("from"), Rule("test")))))))
    def raise_stmt__30(self, raise_loc, exc_opt):
        """(3.0-) raise_stmt: 'raise' [test ['from' test]]"""
        exc = from_loc = cause = None
        loc = raise_loc
        if exc_opt:
            exc, cause_opt = exc_opt
            loc = loc.join(exc.loc)
            if cause_opt:
                from_loc, cause = cause_opt
                loc = loc.join(cause.loc)
        return ast.Raise(exc=exc, inst=None, tback=None, cause=cause,
                         keyword_loc=raise_loc, from_loc=from_loc, loc=loc)

    import_stmt = Alt(Rule("import_name"), Rule("import_from"))
    """import_stmt: import_name | import_from"""

    @action(Seq(Loc("import"), Rule("dotted_as_names")))
    def import_name(self, import_loc, names):
        """import_name: 'import' dotted_as_names"""
        return ast.Import(names=names,
                          keyword_loc=import_loc, loc=import_loc.join(names[-1].loc))

    @action(Loc("."))
    def import_from_1(self, loc):
        return 1, loc

    @action(Loc("..."))
    def import_from_2(self, loc):
        return 3, loc

    @action(Seq(Star(Alt(import_from_1, import_from_2)), Rule("dotted_name")))
    def import_from_3(self, dots, dotted_name):
        dots_loc, dots_count = None, 0
        if any(dots):
            dots_loc = dots[0][1].join(dots[-1][1])
            dots_count = sum([count for count, loc in dots])
        return (dots_loc, dots_count), dotted_name

    @action(Plus(Alt(import_from_1, import_from_2)))
    def import_from_4(self, dots):
        dots_loc = dots[0][1].join(dots[-1][1])
        dots_count = sum([count for count, loc in dots])
        return (dots_loc, dots_count), None

    @action(Loc("*"))
    def import_from_5(self, star_loc):
        return (None, 0), \
               [ast.alias(name="*", asname=None,
                          name_loc=star_loc, as_loc=None, asname_loc=None, loc=star_loc)], \
               None

    @action(Rule("atom_5"))
    def import_from_7(self, string):
        return (None, 0), (string.loc, string.s)

    @action(Rule("import_as_names"))
    def import_from_6(self, names):
        return (None, 0), names, None

    @action(Seq(Loc("from"), Alt(import_from_3, import_from_4, import_from_7),
                Loc("import"), Alt(import_from_5,
                                   Seq(Loc("("), Rule("import_as_names"), Loc(")")),
                                   import_from_6)))
    def import_from(self, from_loc, module_name, import_loc, names):
        """
        (2.6, 2.7)
        import_from: ('from' ('.'* dotted_name | '.'+)
                      'import' ('*' | '(' import_as_names ')' | import_as_names))
        (3.0-)
        # note below: the ('.' | '...') is necessary because '...' is tokenized as ELLIPSIS
        import_from: ('from' (('.' | '...')* dotted_name | ('.' | '...')+)
                      'import' ('*' | '(' import_as_names ')' | import_as_names))
        """
        (dots_loc, dots_count), dotted_name_opt = module_name
        module_loc = module = None
        if dotted_name_opt:
            module_loc, module = dotted_name_opt
        lparen_loc, names, rparen_loc = names
        loc = from_loc.join(names[-1].loc)
        if rparen_loc:
            loc = loc.join(rparen_loc)

        if module == "__future__":
            self.add_flags([x.name for x in names])

        return ast.ImportFrom(names=names, module=module, level=dots_count,
                              keyword_loc=from_loc, dots_loc=dots_loc, module_loc=module_loc,
                              import_loc=import_loc, lparen_loc=lparen_loc, rparen_loc=rparen_loc,
                              loc=loc)

    @action(Seq(Tok("ident"), Opt(Seq(Loc("as"), Tok("ident")))))
    def import_as_name(self, name_tok, as_name_opt):
        """import_as_name: NAME ['as' NAME]"""
        asname_name = asname_loc = as_loc = None
        loc = name_tok.loc
        if as_name_opt:
            as_loc, asname = as_name_opt
            asname_name = asname.value
            asname_loc = asname.loc
            loc = loc.join(asname.loc)
        return ast.alias(name=name_tok.value, asname=asname_name,
                         loc=loc, name_loc=name_tok.loc, as_loc=as_loc, asname_loc=asname_loc)

    @action(Seq(Rule("dotted_name"), Opt(Seq(Loc("as"), Tok("ident")))))
    def dotted_as_name(self, dotted_name, as_name_opt):
        """dotted_as_name: dotted_name ['as' NAME]"""
        asname_name = asname_loc = as_loc = None
        dotted_name_loc, dotted_name_name = dotted_name
        loc = dotted_name_loc
        if as_name_opt:
            as_loc, asname = as_name_opt
            asname_name = asname.value
            asname_loc = asname.loc
            loc = loc.join(asname.loc)
        return ast.alias(name=dotted_name_name, asname=asname_name,
                         loc=loc, name_loc=dotted_name_loc, as_loc=as_loc, asname_loc=asname_loc)

    @action(Seq(Rule("atom_5"), Opt(Seq(Loc("as"), Tok("ident")))))
    def str_as_name(self, string, as_name_opt):
        asname_name = asname_loc = as_loc = None
        loc = string.loc
        if as_name_opt:
            as_loc, asname = as_name_opt
            asname_name = asname.value
            asname_loc = asname.loc
            loc = loc.join(asname.loc)
        return ast.alias(name=string.s, asname=asname_name,
                         loc=loc, name_loc=string.loc, as_loc=as_loc, asname_loc=asname_loc)

    import_as_names = List(Rule("import_as_name"), ",", trailing=True)
    """import_as_names: import_as_name (',' import_as_name)* [',']"""

    dotted_as_names = List(Alt(Rule("dotted_as_name"), Rule("str_as_name")), ",", trailing=False)
    """dotted_as_names: dotted_as_name (',' dotted_as_name)*"""

    @action(List(Tok("ident"), ".", trailing=False))
    def dotted_name(self, idents):
        """dotted_name: NAME ('.' NAME)*"""
        return idents[0].loc.join(idents[-1].loc), \
               ".".join(list(map(lambda x: x.value, idents)))

    @action(Seq(Loc("global"), List(Tok("ident"), ",", trailing=False)))
    def global_stmt(self, global_loc, names):
        """global_stmt: 'global' NAME (',' NAME)*"""
        return ast.Global(names=list(map(lambda x: x.value, names)),
                          name_locs=list(map(lambda x: x.loc, names)),
                          keyword_loc=global_loc, loc=global_loc.join(names[-1].loc))

    @action(Seq(Loc("exec"), Rule("expr"),
                Opt(Seq(Loc("in"), Rule("test"),
                        Opt(SeqN(1, Loc(","), Rule("test")))))))
    def exec_stmt(self, exec_loc, body, in_opt):
        """(2.6, 2.7) exec_stmt: 'exec' expr ['in' test [',' test]]"""
        in_loc, globals, locals = None, None, None
        loc = exec_loc.join(body.loc)
        if in_opt:
            in_loc, globals, locals = in_opt
            if locals:
                loc = loc.join(locals.loc)
            else:
                loc = loc.join(globals.loc)
        return ast.Exec(body=body, locals=locals, globals=globals,
                        loc=loc, keyword_loc=exec_loc, in_loc=in_loc)

    @action(Seq(Loc("nonlocal"), List(Tok("ident"), ",", trailing=False)))
    def nonlocal_stmt(self, nonlocal_loc, names):
        """(3.0-) nonlocal_stmt: 'nonlocal' NAME (',' NAME)*"""
        return ast.Nonlocal(names=list(map(lambda x: x.value, names)),
                            name_locs=list(map(lambda x: x.loc, names)),
                            keyword_loc=nonlocal_loc, loc=nonlocal_loc.join(names[-1].loc))

    @action(Seq(Loc("assert"), Rule("test"), Opt(SeqN(1, Tok(","), Rule("test")))))
    def assert_stmt(self, assert_loc, test, msg):
        """assert_stmt: 'assert' test [',' test]"""
        loc = assert_loc.join(test.loc)
        if msg:
            loc = loc.join(msg.loc)
        return ast.Assert(test=test, msg=msg,
                          loc=loc, keyword_loc=assert_loc)

    @action(Alt(Rule("if_stmt"), Rule("while_stmt"), Rule("for_stmt"),
                Rule("try_stmt"), Rule("with_stmt"), Rule("funcdef"),
                Rule("classdef"), Rule("decorated")))
    def compound_stmt(self, stmt):
        """compound_stmt: if_stmt | while_stmt | for_stmt | try_stmt | with_stmt |
                          funcdef | classdef | decorated"""
        return [stmt]

    @action(Seq(Loc("if"), Rule("test"), Loc(":"), Rule("suite"),
                Star(Seq(Loc("elif"), Rule("test"), Loc(":"), Rule("suite"))),
                Opt(Seq(Loc("else"), Loc(":"), Rule("suite")))))
    def if_stmt(self, if_loc, test, if_colon_loc, body, elifs, else_opt):
        """if_stmt: 'if' test ':' suite ('elif' test ':' suite)* ['else' ':' suite]"""
        stmt = ast.If(orelse=[],
                      else_loc=None, else_colon_loc=None)

        if else_opt:
            stmt.else_loc, stmt.else_colon_loc, stmt.orelse = else_opt

        for elif_ in reversed(elifs):
            stmt.keyword_loc, stmt.test, stmt.if_colon_loc, stmt.body = elif_
            stmt.loc = stmt.keyword_loc.join(stmt.body[-1].loc)
            if stmt.orelse:
                stmt.loc = stmt.loc.join(stmt.orelse[-1].loc)
            stmt = ast.If(orelse=[stmt],
                          else_loc=None, else_colon_loc=None)

        stmt.keyword_loc, stmt.test, stmt.if_colon_loc, stmt.body = \
            if_loc, test, if_colon_loc, body
        stmt.loc = stmt.keyword_loc.join(stmt.body[-1].loc)
        if stmt.orelse:
            stmt.loc = stmt.loc.join(stmt.orelse[-1].loc)
        return stmt

    @action(Seq(Loc("while"), Rule("test"), Loc(":"), Rule("suite"),
                Opt(Seq(Loc("else"), Loc(":"), Rule("suite")))))
    def while_stmt(self, while_loc, test, while_colon_loc, body, else_opt):
        """while_stmt: 'while' test ':' suite ['else' ':' suite]"""
        stmt = ast.While(test=test, body=body, orelse=[],
                         keyword_loc=while_loc, while_colon_loc=while_colon_loc,
                         else_loc=None, else_colon_loc=None,
                         loc=while_loc.join(body[-1].loc))
        if else_opt:
            stmt.else_loc, stmt.else_colon_loc, stmt.orelse = else_opt
            stmt.loc = stmt.loc.join(stmt.orelse[-1].loc)

        return stmt

    @action(Seq(Loc("for"), Rule("exprlist"), Loc("in"), Rule("testlist"),
                Loc(":"), Rule("suite"),
                Opt(Seq(Loc("else"), Loc(":"), Rule("suite")))))
    def for_stmt(self, for_loc, target, in_loc, iter, for_colon_loc, body, else_opt):
        """for_stmt: 'for' exprlist 'in' testlist ':' suite ['else' ':' suite]"""
        stmt = ast.For(target=self._assignable(target), iter=iter, body=body, orelse=[],
                       keyword_loc=for_loc, in_loc=in_loc, for_colon_loc=for_colon_loc,
                       else_loc=None, else_colon_loc=None,
                       loc=for_loc.join(body[-1].loc))
        if else_opt:
            stmt.else_loc, stmt.else_colon_loc, stmt.orelse = else_opt
            stmt.loc = stmt.loc.join(stmt.orelse[-1].loc)

        return stmt

    @action(Seq(Plus(Seq(Rule("except_clause"), Loc(":"), Rule("suite"))),
                Opt(Seq(Loc("else"), Loc(":"), Rule("suite"))),
                Opt(Seq(Loc("finally"), Loc(":"), Rule("suite")))))
    def try_stmt_1(self, clauses, else_opt, finally_opt):
        handlers = []
        for clause in clauses:
            handler, handler.colon_loc, handler.body = clause
            handler.loc = handler.loc.join(handler.body[-1].loc)
            handlers.append(handler)

        else_loc, else_colon_loc, orelse = None, None, []
        loc = handlers[-1].loc
        if else_opt:
            else_loc, else_colon_loc, orelse = else_opt
            loc = orelse[-1].loc

        finally_loc, finally_colon_loc, finalbody = None, None, []
        if finally_opt:
            finally_loc, finally_colon_loc, finalbody = finally_opt
            loc = finalbody[-1].loc
        stmt = ast.Try(body=None, handlers=handlers, orelse=orelse, finalbody=finalbody,
                       else_loc=else_loc, else_colon_loc=else_colon_loc,
                       finally_loc=finally_loc, finally_colon_loc=finally_colon_loc,
                       loc=loc)
        return stmt

    @action(Seq(Loc("finally"), Loc(":"), Rule("suite")))
    def try_stmt_2(self, finally_loc, finally_colon_loc, finalbody):
        return ast.Try(body=None, handlers=[], orelse=[], finalbody=finalbody,
                       else_loc=None, else_colon_loc=None,
                       finally_loc=finally_loc, finally_colon_loc=finally_colon_loc,
                       loc=finalbody[-1].loc)

    @action(Seq(Loc("try"), Loc(":"), Rule("suite"), Alt(try_stmt_1, try_stmt_2)))
    def try_stmt(self, try_loc, try_colon_loc, body, stmt):
        """
        try_stmt: ('try' ':' suite
                   ((except_clause ':' suite)+
                    ['else' ':' suite]
                    ['finally' ':' suite] |
                    'finally' ':' suite))
        """
        stmt.keyword_loc, stmt.try_colon_loc, stmt.body = \
            try_loc, try_colon_loc, body
        stmt.loc = stmt.loc.join(try_loc)
        return stmt

    @action(Seq(Loc("with"), Rule("test"), Opt(Rule("with_var")), Loc(":"), Rule("suite")))
    def with_stmt__26(self, with_loc, context, with_var, colon_loc, body):
        """(2.6, 3.0) with_stmt: 'with' test [ with_var ] ':' suite"""
        if with_var:
            as_loc, optional_vars = with_var
            item = ast.withitem(context_expr=context, optional_vars=optional_vars,
                                as_loc=as_loc, loc=context.loc.join(optional_vars.loc))
        else:
            item = ast.withitem(context_expr=context, optional_vars=None,
                                as_loc=None, loc=context.loc)
        return ast.With(items=[item], body=body,
                        keyword_loc=with_loc, colon_loc=colon_loc,
                        loc=with_loc.join(body[-1].loc))

    with_var = Seq(Loc("as"), Rule("expr"))
    """(2.6, 3.0) with_var: 'as' expr"""

    @action(Seq(Loc("with"), List(Rule("with_item"), ",", trailing=False), Loc(":"),
                Rule("suite")))
    def with_stmt__27(self, with_loc, items, colon_loc, body):
        """(2.7, 3.1-) with_stmt: 'with' with_item (',' with_item)*  ':' suite"""
        return ast.With(items=items, body=body,
                        keyword_loc=with_loc, colon_loc=colon_loc,
                        loc=with_loc.join(body[-1].loc))

    @action(Seq(Rule("test"), Opt(Seq(Loc("as"), Rule("expr")))))
    def with_item(self, context, as_opt):
        """(2.7, 3.1-) with_item: test ['as' expr]"""
        if as_opt:
            as_loc, optional_vars = as_opt
            return ast.withitem(context_expr=context, optional_vars=optional_vars,
                                as_loc=as_loc, loc=context.loc.join(optional_vars.loc))
        else:
            return ast.withitem(context_expr=context, optional_vars=None,
                                as_loc=None, loc=context.loc)

    @action(Seq(Alt(Loc("as"), Loc(",")), Rule("test")))
    def except_clause_1__26(self, as_loc, name):
        return as_loc, None, name

    @action(Seq(Loc("as"), Tok("ident")))
    def except_clause_1__30(self, as_loc, name):
        return as_loc, name, None

    @action(Seq(Loc("except"),
                Opt(Seq(Rule("test"),
                        Opt(Rule("except_clause_1"))))))
    def except_clause(self, except_loc, exc_opt):
        """
        (2.6, 2.7) except_clause: 'except' [test [('as' | ',') test]]
        (3.0-) except_clause: 'except' [test ['as' NAME]]
        """
        type_ = name = as_loc = name_loc = None
        loc = except_loc
        if exc_opt:
            type_, name_opt = exc_opt
            loc = loc.join(type_.loc)
            if name_opt:
                as_loc, name_tok, name_node = name_opt
                if name_tok:
                    name = name_tok.value
                    name_loc = name_tok.loc
                else:
                    name = name_node
                    name_loc = name_node.loc
                loc = loc.join(name_loc)
        return ast.ExceptHandler(type=type_, name=name,
                                 except_loc=except_loc, as_loc=as_loc, name_loc=name_loc,
                                 loc=loc)

    @action(Plus(Rule("stmt")))
    def suite_1(self, stmts):
        return reduce(list.__add__, stmts, [])

    suite = Alt(Rule("simple_stmt"),
                SeqN(2, Tok("newline"), Tok("indent"), suite_1, Tok("dedent")))
    """suite: simple_stmt | NEWLINE INDENT stmt+ DEDENT"""

    # 2.x-only backwards compatibility start
    testlist_safe = action(List(Rule("old_test"), ",", trailing=False))(_wrap_tuple)
    """(2.6, 2.7) testlist_safe: old_test [(',' old_test)+ [',']]"""

    old_test = Alt(Rule("or_test"), Rule("old_lambdef"))
    """(2.6, 2.7) old_test: or_test | old_lambdef"""

    @action(Seq(Loc("lambda"), Opt(Rule("varargslist")), Loc(":"), Rule("old_test")))
    def old_lambdef(self, lambda_loc, args_opt, colon_loc, body):
        """(2.6, 2.7) old_lambdef: 'lambda' [varargslist] ':' old_test"""
        if args_opt is None:
            args_opt = self._arguments()
            args_opt.loc = colon_loc.begin()
        return ast.Lambda(args=args_opt, body=body,
                          lambda_loc=lambda_loc, colon_loc=colon_loc,
                          loc=lambda_loc.join(body.loc))
    # 2.x-only backwards compatibility end

    @action(Seq(Rule("or_test"), Opt(Seq(Loc("if"), Rule("or_test"),
                                         Loc("else"), Rule("test")))))
    def test_1(self, lhs, rhs_opt):
        if rhs_opt is not None:
            if_loc, test, else_loc, orelse = rhs_opt
            return ast.IfExp(test=test, body=lhs, orelse=orelse,
                             if_loc=if_loc, else_loc=else_loc, loc=lhs.loc.join(orelse.loc))
        return lhs

    test = Alt(test_1, Rule("lambdef"))
    """test: or_test ['if' or_test 'else' test] | lambdef"""

    test_nocond = Alt(Rule("or_test"), Rule("lambdef_nocond"))
    """(3.0-) test_nocond: or_test | lambdef_nocond"""

    def lambdef_action(self, lambda_loc, args_opt, colon_loc, body):
        if args_opt is None:
            args_opt = self._arguments()
            args_opt.loc = colon_loc.begin()
        return ast.Lambda(args=args_opt, body=body,
                          lambda_loc=lambda_loc, colon_loc=colon_loc,
                          loc=lambda_loc.join(body.loc))

    lambdef = action(
        Seq(Loc("lambda"), Opt(Rule("varargslist")), Loc(":"), Rule("test"))) \
        (lambdef_action)
    """lambdef: 'lambda' [varargslist] ':' test"""

    lambdef_nocond = action(
        Seq(Loc("lambda"), Opt(Rule("varargslist")), Loc(":"), Rule("test_nocond"))) \
        (lambdef_action)
    """(3.0-) lambdef_nocond: 'lambda' [varargslist] ':' test_nocond"""

    @action(Seq(Rule("and_test"), Star(Seq(Loc("or"), Rule("and_test")))))
    def or_test(self, lhs, rhs):
        """or_test: and_test ('or' and_test)*"""
        if len(rhs) > 0:
            return ast.BoolOp(op=ast.Or(),
                              values=[lhs] + list(map(lambda x: x[1], rhs)),
                              loc=lhs.loc.join(rhs[-1][1].loc),
                              op_locs=list(map(lambda x: x[0], rhs)))
        else:
            return lhs

    @action(Seq(Rule("not_test"), Star(Seq(Loc("and"), Rule("not_test")))))
    def and_test(self, lhs, rhs):
        """and_test: not_test ('and' not_test)*"""
        if len(rhs) > 0:
            return ast.BoolOp(op=ast.And(),
                              values=[lhs] + list(map(lambda x: x[1], rhs)),
                              loc=lhs.loc.join(rhs[-1][1].loc),
                              op_locs=list(map(lambda x: x[0], rhs)))
        else:
            return lhs

    @action(Seq(Oper(ast.Not, "not"), Rule("not_test")))
    def not_test_1(self, op, operand):
        return ast.UnaryOp(op=op, operand=operand,
                           loc=op.loc.join(operand.loc))

    not_test = Alt(not_test_1, Rule("comparison"))
    """not_test: 'not' not_test | comparison"""

    comparison_1__26 = Seq(Rule("expr"), Star(Seq(Rule("comp_op"), Rule("expr"))))
    comparison_1__30 = Seq(Rule("star_expr"), Star(Seq(Rule("comp_op"), Rule("star_expr"))))
    comparison_1__32 = comparison_1__26

    @action(Rule("comparison_1"))
    def comparison(self, lhs, rhs):
        """
        (2.6, 2.7) comparison: expr (comp_op expr)*
        (3.0, 3.1) comparison: star_expr (comp_op star_expr)*
        (3.2-) comparison: expr (comp_op expr)*
        """
        if len(rhs) > 0:
            return ast.Compare(left=lhs, ops=list(map(lambda x: x[0], rhs)),
                               comparators=list(map(lambda x: x[1], rhs)),
                               loc=lhs.loc.join(rhs[-1][1].loc))
        else:
            return lhs

    @action(Seq(Opt(Loc("*")), Rule("expr")))
    def star_expr__30(self, star_opt, expr):
        """(3.0, 3.1) star_expr: ['*'] expr"""
        if star_opt:
            return ast.Starred(value=expr, ctx=None,
                               star_loc=star_opt, loc=expr.loc.join(star_opt))
        return expr

    @action(Seq(Loc("*"), Rule("expr")))
    def star_expr__32(self, star_loc, expr):
        """(3.0-) star_expr: '*' expr"""
        return ast.Starred(value=expr, ctx=None,
                           star_loc=star_loc, loc=expr.loc.join(star_loc))

    comp_op = Alt(Oper(ast.Lt, "<"), Oper(ast.Gt, ">"), Oper(ast.Eq, "=="),
                  Oper(ast.GtE, ">="), Oper(ast.LtE, "<="), Oper(ast.NotEq, "<>"),
                  Oper(ast.NotEq, "!="),
                  Oper(ast.In, "in"), Oper(ast.NotIn, "not", "in"),
                  Oper(ast.IsNot, "is", "not"), Oper(ast.Is, "is"))
    """
    (2.6, 2.7) comp_op: '<'|'>'|'=='|'>='|'<='|'<>'|'!='|'in'|'not' 'in'|'is'|'is' 'not'
    (3.0-) comp_op: '<'|'>'|'=='|'>='|'<='|'!='|'in'|'not' 'in'|'is'|'is' 'not'
    """

    expr = BinOper("xor_expr", Oper(ast.BitOr, "|"))
    """expr: xor_expr ('|' xor_expr)*"""

    xor_expr = BinOper("and_expr", Oper(ast.BitXor, "^"))
    """xor_expr: and_expr ('^' and_expr)*"""

    and_expr = BinOper("shift_expr", Oper(ast.BitAnd, "&"))
    """and_expr: shift_expr ('&' shift_expr)*"""

    shift_expr = BinOper("arith_expr", Alt(Oper(ast.LShift, "<<"), Oper(ast.RShift, ">>")))
    """shift_expr: arith_expr (('<<'|'>>') arith_expr)*"""

    arith_expr = BinOper("term", Alt(Oper(ast.Add, "+"), Oper(ast.Sub, "-")))
    """arith_expr: term (('+'|'-') term)*"""

    term = BinOper("factor", Alt(Oper(ast.Mult, "*"), Oper(ast.MatMult, "@"),
                                 Oper(ast.Div, "/"), Oper(ast.Mod, "%"),
                                 Oper(ast.FloorDiv, "//")))
    """term: factor (('*'|'/'|'%'|'//') factor)*"""

    @action(Seq(Alt(Oper(ast.UAdd, "+"), Oper(ast.USub, "-"), Oper(ast.Invert, "~")),
                Rule("factor")))
    def factor_1(self, op, factor):
        return ast.UnaryOp(op=op, operand=factor,
                           loc=op.loc.join(factor.loc))

    factor = Alt(factor_1, Rule("power"))
    """factor: ('+'|'-'|'~') factor | power"""

    @action(Seq(Rule("atom"), Star(Rule("trailer")), Opt(Seq(Loc("**"), Rule("factor")))))
    def power(self, atom, trailers, factor_opt):
        """power: atom trailer* ['**' factor]"""
        for trailer in trailers:
            if isinstance(trailer, ast.Attribute) or isinstance(trailer, ast.Subscript):
                trailer.value = atom
            elif isinstance(trailer, ast.Call):
                trailer.func = atom
            trailer.loc = atom.loc.join(trailer.loc)
            atom = trailer
        if factor_opt:
            op_loc, factor = factor_opt
            return ast.BinOp(left=atom, op=ast.Pow(loc=op_loc), right=factor,
                             loc=atom.loc.join(factor.loc))
        return atom

    @action(Rule("testlist1"))
    def atom_1(self, expr):
        return ast.Repr(value=expr, loc=None)

    @action(Tok("ident"))
    def atom_2(self, tok):
        return ast.Name(id=tok.value, loc=tok.loc, ctx=None)

    @action(Alt(Tok("int"), Tok("float"), Tok("complex")))
    def atom_3(self, tok):
        return ast.Num(n=tok.value, loc=tok.loc)

    @action(Seq(Tok("strbegin"), Tok("strdata"), Tok("strend")))
    def atom_4(self, begin_tok, data_tok, end_tok):
        return ast.Str(s=data_tok.value,
                       begin_loc=begin_tok.loc, end_loc=end_tok.loc,
                       loc=begin_tok.loc.join(end_tok.loc))

    @action(Plus(atom_4))
    def atom_5(self, strings):
        joint = ""
        if all(isinstance(x.s, bytes) for x in strings):
            joint = b""
        return ast.Str(s=joint.join([x.s for x in strings]),
                       begin_loc=strings[0].begin_loc, end_loc=strings[-1].end_loc,
                       loc=strings[0].loc.join(strings[-1].loc))

    atom_6__26 = Rule("dictmaker")
    atom_6__27 = Rule("dictorsetmaker")

    atom__26 = Alt(BeginEnd("(", Opt(Alt(Rule("yield_expr"), Rule("testlist_comp"))), ")",
                            empty=lambda self: ast.Tuple(elts=[], ctx=None, loc=None)),
                   BeginEnd("[", Opt(Rule("listmaker")), "]",
                            empty=lambda self: ast.List(elts=[], ctx=None, loc=None)),
                   BeginEnd("{", Opt(Rule("atom_6")), "}",
                            empty=lambda self: ast.Dict(keys=[], values=[], colon_locs=[],
                                                        loc=None)),
                   BeginEnd("`", atom_1, "`"),
                   atom_2, atom_3, atom_5)
    """
    (2.6)
    atom: ('(' [yield_expr|testlist_gexp] ')' |
           '[' [listmaker] ']' |
           '{' [dictmaker] '}' |
           '`' testlist1 '`' |
           NAME | NUMBER | STRING+)
    (2.7)
    atom: ('(' [yield_expr|testlist_comp] ')' |
           '[' [listmaker] ']' |
           '{' [dictorsetmaker] '}' |
           '`' testlist1 '`' |
           NAME | NUMBER | STRING+)
    """

    @action(Loc("..."))
    def atom_7(self, loc):
        return ast.Ellipsis(loc=loc)

    @action(Alt(Tok("None"), Tok("True"), Tok("False")))
    def atom_8(self, tok):
        if tok.kind == "None":
            value = None
        elif tok.kind == "True":
            value = True
        elif tok.kind == "False":
            value = False
        return ast.NameConstant(value=value, loc=tok.loc)

    atom__30 = Alt(BeginEnd("(", Opt(Alt(Rule("yield_expr"), Rule("testlist_comp"))), ")",
                            empty=lambda self: ast.Tuple(elts=[], ctx=None, loc=None)),
                   BeginEnd("[", Opt(Rule("testlist_comp__list")), "]",
                            empty=lambda self: ast.List(elts=[], ctx=None, loc=None)),
                   BeginEnd("{", Opt(Rule("dictorsetmaker")), "}",
                            empty=lambda self: ast.Dict(keys=[], values=[], colon_locs=[],
                                                        loc=None)),
                   atom_2, atom_3, atom_5, atom_7, atom_8)
    """
    (3.0-)
    atom: ('(' [yield_expr|testlist_comp] ')' |
           '[' [testlist_comp] ']' |
           '{' [dictorsetmaker] '}' |
           NAME | NUMBER | STRING+ | '...' | 'None' | 'True' | 'False')
    """

    def list_gen_action(self, lhs, rhs):
        if rhs is None: # (x)
            return lhs
        elif isinstance(rhs, ast.Tuple) or isinstance(rhs, ast.List):
            rhs.elts = [lhs] + rhs.elts
            return rhs
        elif isinstance(rhs, ast.ListComp) or isinstance(rhs, ast.GeneratorExp):
            rhs.elt = lhs
            return rhs

    @action(Rule("list_for"))
    def listmaker_1(self, compose):
        return ast.ListComp(generators=compose([]), loc=None)

    @action(List(Rule("test"), ",", trailing=True, leading=False))
    def listmaker_2(self, elts):
        return ast.List(elts=elts, ctx=None, loc=None)

    listmaker = action(
        Seq(Rule("test"),
            Alt(listmaker_1, listmaker_2))) \
        (list_gen_action)
    """listmaker: test ( list_for | (',' test)* [','] )"""

    testlist_comp_1__26 = Rule("test")
    testlist_comp_1__32 = Alt(Rule("test"), Rule("star_expr"))

    @action(Rule("comp_for"))
    def testlist_comp_2(self, compose):
        return ast.GeneratorExp(generators=compose([]), loc=None)

    @action(List(Rule("testlist_comp_1"), ",", trailing=True, leading=False))
    def testlist_comp_3(self, elts):
        if elts == [] and not elts.trailing_comma:
            return None
        else:
            return ast.Tuple(elts=elts, ctx=None, loc=None)

    testlist_comp = action(
        Seq(Rule("testlist_comp_1"), Alt(testlist_comp_2, testlist_comp_3))) \
        (list_gen_action)
    """
    (2.6) testlist_gexp: test ( gen_for | (',' test)* [','] )
    (2.7, 3.0, 3.1) testlist_comp: test ( comp_for | (',' test)* [','] )
    (3.2-) testlist_comp: (test|star_expr) ( comp_for | (',' (test|star_expr))* [','] )
    """

    @action(Rule("comp_for"))
    def testlist_comp__list_1(self, compose):
        return ast.ListComp(generators=compose([]), loc=None)

    @action(List(Rule("testlist_comp_1"), ",", trailing=True, leading=False))
    def testlist_comp__list_2(self, elts):
        return ast.List(elts=elts, ctx=None, loc=None)

    testlist_comp__list = action(
        Seq(Rule("testlist_comp_1"), Alt(testlist_comp__list_1, testlist_comp__list_2))) \
        (list_gen_action)
    """Same grammar as testlist_comp, but different semantic action."""

    @action(Seq(Loc("."), Tok("ident")))
    def trailer_1(self, dot_loc, ident_tok):
        return ast.Attribute(attr=ident_tok.value, ctx=None,
                             loc=dot_loc.join(ident_tok.loc),
                             attr_loc=ident_tok.loc, dot_loc=dot_loc)

    trailer = Alt(BeginEnd("(", Opt(Rule("arglist")), ")",
                           empty=_empty_arglist),
                  BeginEnd("[", Rule("subscriptlist"), "]"),
                  trailer_1)
    """trailer: '(' [arglist] ')' | '[' subscriptlist ']' | '.' NAME"""

    @action(List(Rule("subscript"), ",", trailing=True))
    def subscriptlist(self, subscripts):
        """subscriptlist: subscript (',' subscript)* [',']"""
        if len(subscripts) == 1:
            return ast.Subscript(slice=subscripts[0], ctx=None, loc=None)
        elif all([isinstance(x, ast.Index) for x in subscripts]):
            elts  = [x.value for x in subscripts]
            loc   = subscripts[0].loc.join(subscripts[-1].loc)
            index = ast.Index(value=ast.Tuple(elts=elts, ctx=None,
                                              begin_loc=None, end_loc=None, loc=loc),
                              loc=loc)
            return ast.Subscript(slice=index, ctx=None, loc=None)
        else:
            extslice = ast.ExtSlice(dims=subscripts,
                                    loc=subscripts[0].loc.join(subscripts[-1].loc))
            return ast.Subscript(slice=extslice, ctx=None, loc=None)

    @action(Seq(Loc("."), Loc("."), Loc(".")))
    def subscript_1(self, dot_1_loc, dot_2_loc, dot_3_loc):
        return ast.Ellipsis(loc=dot_1_loc.join(dot_3_loc))

    @action(Seq(Opt(Rule("test")), Loc(":"), Opt(Rule("test")), Opt(Rule("sliceop"))))
    def subscript_2(self, lower_opt, colon_loc, upper_opt, step_opt):
        loc = colon_loc
        if lower_opt:
            loc = loc.join(lower_opt.loc)
        if upper_opt:
            loc = loc.join(upper_opt.loc)
        step_colon_loc = step = None
        if step_opt:
            step_colon_loc, step = step_opt
            loc = loc.join(step_colon_loc)
            if step:
                loc = loc.join(step.loc)
        return ast.Slice(lower=lower_opt, upper=upper_opt, step=step,
                         loc=loc, bound_colon_loc=colon_loc, step_colon_loc=step_colon_loc)

    @action(Rule("test"))
    def subscript_3(self, expr):
        return ast.Index(value=expr, loc=expr.loc)

    subscript__26 = Alt(subscript_1, subscript_2, subscript_3)
    """(2.6, 2.7) subscript: '.' '.' '.' | test | [test] ':' [test] [sliceop]"""

    subscript__30 = Alt(subscript_2, subscript_3)
    """(3.0-) subscript: test | [test] ':' [test] [sliceop]"""

    sliceop = Seq(Loc(":"), Opt(Rule("test")))
    """sliceop: ':' [test]"""

    exprlist_1__26 = List(Rule("expr"), ",", trailing=True)
    exprlist_1__30 = List(Rule("star_expr"), ",", trailing=True)
    exprlist_1__32 = List(Alt(Rule("expr"), Rule("star_expr")), ",", trailing=True)

    @action(Rule("exprlist_1"))
    def exprlist(self, exprs):
        """
        (2.6, 2.7) exprlist: expr (',' expr)* [',']
        (3.0, 3.1) exprlist: star_expr (',' star_expr)* [',']
        (3.2-) exprlist: (expr|star_expr) (',' (expr|star_expr))* [',']
        """
        return self._wrap_tuple(exprs)

    @action(List(Rule("test"), ",", trailing=True))
    def testlist(self, exprs):
        """testlist: test (',' test)* [',']"""
        return self._wrap_tuple(exprs)

    @action(List(Seq(Rule("test"), Loc(":"), Rule("test")), ",", trailing=True))
    def dictmaker(self, elts):
        """(2.6) dictmaker: test ':' test (',' test ':' test)* [',']"""
        return ast.Dict(keys=list(map(lambda x: x[0], elts)),
                        values=list(map(lambda x: x[2], elts)),
                        colon_locs=list(map(lambda x: x[1], elts)),
                        loc=None)

    dictorsetmaker_1 = Seq(Rule("test"), Loc(":"), Rule("test"))

    @action(Seq(dictorsetmaker_1,
                Alt(Rule("comp_for"),
                    List(dictorsetmaker_1, ",", leading=False, trailing=True))))
    def dictorsetmaker_2(self, first, elts):
        if isinstance(elts, commalist):
            elts.insert(0, first)
            return ast.Dict(keys=list(map(lambda x: x[0], elts)),
                            values=list(map(lambda x: x[2], elts)),
                            colon_locs=list(map(lambda x: x[1], elts)),
                            loc=None)
        else:
            return ast.DictComp(key=first[0], value=first[2], generators=elts([]),
                                colon_loc=first[1],
                                begin_loc=None, end_loc=None, loc=None)

    @action(Seq(Rule("test"),
                Alt(Rule("comp_for"),
                    List(Rule("test"), ",", leading=False, trailing=True))))
    def dictorsetmaker_3(self, first, elts):
        if isinstance(elts, commalist):
            elts.insert(0, first)
            return ast.Set(elts=elts, loc=None)
        else:
            return ast.SetComp(elt=first, generators=elts([]),
                               begin_loc=None, end_loc=None, loc=None)

    dictorsetmaker = Alt(dictorsetmaker_2, dictorsetmaker_3)
    """
    (2.7-)
    dictorsetmaker: ( (test ':' test (comp_for | (',' test ':' test)* [','])) |
                      (test (comp_for | (',' test)* [','])) )
    """

    @action(Seq(Loc("class"), Tok("ident"),
                Opt(Seq(Loc("("), List(Rule("test"), ",", trailing=True), Loc(")"))),
                Loc(":"), Rule("suite")))
    def classdef__26(self, class_loc, name_tok, bases_opt, colon_loc, body):
        """(2.6, 2.7) classdef: 'class' NAME ['(' [testlist] ')'] ':' suite"""
        bases, lparen_loc, rparen_loc = [], None, None
        if bases_opt:
            lparen_loc, bases, rparen_loc = bases_opt

        return ast.ClassDef(name=name_tok.value, bases=bases, keywords=[],
                            starargs=None, kwargs=None, body=body,
                            decorator_list=[], at_locs=[],
                            keyword_loc=class_loc, lparen_loc=lparen_loc,
                            star_loc=None, dstar_loc=None, rparen_loc=rparen_loc,
                            name_loc=name_tok.loc, colon_loc=colon_loc,
                            loc=class_loc.join(body[-1].loc))

    @action(Seq(Loc("class"), Tok("ident"),
                Opt(Seq(Loc("("), Rule("arglist"), Loc(")"))),
                Loc(":"), Rule("suite")))
    def classdef__30(self, class_loc, name_tok, arglist_opt, colon_loc, body):
        """(3.0) classdef: 'class' NAME ['(' [testlist] ')'] ':' suite"""
        arglist, lparen_loc, rparen_loc = [], None, None
        bases, keywords, starargs, kwargs = [], [], None, None
        star_loc, dstar_loc = None, None
        if arglist_opt:
            lparen_loc, arglist, rparen_loc = arglist_opt
            bases, keywords, starargs, kwargs = \
                arglist.args, arglist.keywords, arglist.starargs, arglist.kwargs
            star_loc, dstar_loc = arglist.star_loc, arglist.dstar_loc

        return ast.ClassDef(name=name_tok.value, bases=bases, keywords=keywords,
                            starargs=starargs, kwargs=kwargs, body=body,
                            decorator_list=[], at_locs=[],
                            keyword_loc=class_loc, lparen_loc=lparen_loc,
                            star_loc=star_loc, dstar_loc=dstar_loc, rparen_loc=rparen_loc,
                            name_loc=name_tok.loc, colon_loc=colon_loc,
                            loc=class_loc.join(body[-1].loc))

    @action(Seq(Loc("*"), Rule("test"), Star(SeqN(1, Tok(","), Rule("argument"))),
                Opt(Seq(Tok(","), Loc("**"), Rule("test")))))
    def arglist_1(self, star_loc, stararg, postargs, kwarg_opt):
        dstar_loc = kwarg = None
        if kwarg_opt:
            _, dstar_loc, kwarg = kwarg_opt

        for postarg in postargs:
            if not isinstance(postarg, ast.keyword):
                error = diagnostic.Diagnostic(
                    "fatal", "only named arguments may follow *expression", {},
                    postarg.loc, [star_loc.join(stararg.loc)])
                self.diagnostic_engine.process(error)

        return postargs, \
               ast.Call(args=[], keywords=[], starargs=stararg, kwargs=kwarg,
                        star_loc=star_loc, dstar_loc=dstar_loc, loc=None)

    @action(Seq(Loc("**"), Rule("test")))
    def arglist_2(self, dstar_loc, kwarg):
        return [], \
               ast.Call(args=[], keywords=[], starargs=None, kwargs=kwarg,
                        star_loc=None, dstar_loc=dstar_loc, loc=None)

    @action(Seq(Rule("argument"),
                Alt(SeqN(1, Tok(","), Alt(Rule("arglist_1"),
                                          Rule("arglist_2"),
                                          Rule("arglist_3"),
                                          Eps())),
                    Eps())))
    def arglist_3(self, arg, cont):
        if cont is None:
            return [arg], self._empty_arglist()
        else:
            args, rest = cont
            return [arg] + args, rest

    @action(Alt(Rule("arglist_1"),
                Rule("arglist_2"),
                Rule("arglist_3")))
    def arglist(self, args, call):
        """arglist: (argument ',')* (argument [','] |
                                     '*' test (',' argument)* [',' '**' test] |
                                     '**' test)"""
        for arg in args:
            if isinstance(arg, ast.keyword):
                call.keywords.append(arg)
            elif len(call.keywords) > 0:
                error = diagnostic.Diagnostic(
                    "fatal", "non-keyword arg after keyword arg", {},
                    arg.loc, [call.keywords[-1].loc])
                self.diagnostic_engine.process(error)
            else:
                call.args.append(arg)
        return call

    @action(Seq(Loc("="), Rule("test")))
    def argument_1(self, equals_loc, rhs):
        def thunk(lhs):
            if not isinstance(lhs, ast.Name):
                error = diagnostic.Diagnostic(
                    "fatal", "keyword must be an identifier", {}, lhs.loc)
                self.diagnostic_engine.process(error)
            return ast.keyword(arg=lhs.id, value=rhs,
                               loc=lhs.loc.join(rhs.loc),
                               arg_loc=lhs.loc, equals_loc=equals_loc)
        return thunk

    @action(Opt(Rule("comp_for")))
    def argument_2(self, compose_opt):
        def thunk(lhs):
            if compose_opt:
                generators = compose_opt([])
                return ast.GeneratorExp(elt=lhs, generators=generators,
                                        begin_loc=None, end_loc=None,
                                        loc=lhs.loc.join(generators[-1].loc))
            return lhs
        return thunk

    @action(Seq(Rule("test"), Alt(argument_1, argument_2)))
    def argument(self, lhs, thunk):
        # This rule is reformulated to avoid exponential backtracking.
        """
        (2.6) argument: test [gen_for] | test '=' test  # Really [keyword '='] test
        (2.7-) argument: test [comp_for] | test '=' test
        """
        return thunk(lhs)

    list_iter = Alt(Rule("list_for"), Rule("list_if"))
    """(2.6, 2.7) list_iter: list_for | list_if"""

    def list_comp_for_action(self, for_loc, target, in_loc, iter, next_opt):
        def compose(comprehensions):
            comp = ast.comprehension(
                target=target, iter=iter, ifs=[],
                loc=for_loc.join(iter.loc), for_loc=for_loc, in_loc=in_loc, if_locs=[])
            comprehensions += [comp]
            if next_opt:
                return next_opt(comprehensions)
            else:
                return comprehensions
        return compose

    def list_comp_if_action(self, if_loc, cond, next_opt):
        def compose(comprehensions):
            comprehensions[-1].ifs.append(cond)
            comprehensions[-1].if_locs.append(if_loc)
            comprehensions[-1].loc = comprehensions[-1].loc.join(cond.loc)
            if next_opt:
                return next_opt(comprehensions)
            else:
                return comprehensions
        return compose

    list_for = action(
        Seq(Loc("for"), Rule("exprlist"),
            Loc("in"), Rule("testlist_safe"), Opt(Rule("list_iter")))) \
        (list_comp_for_action)
    """(2.6, 2.7) list_for: 'for' exprlist 'in' testlist_safe [list_iter]"""

    list_if = action(
        Seq(Loc("if"), Rule("old_test"), Opt(Rule("list_iter")))) \
        (list_comp_if_action)
    """(2.6, 2.7) list_if: 'if' old_test [list_iter]"""

    comp_iter = Alt(Rule("comp_for"), Rule("comp_if"))
    """
    (2.6) gen_iter: gen_for | gen_if
    (2.7-) comp_iter: comp_for | comp_if
    """

    comp_for = action(
        Seq(Loc("for"), Rule("exprlist"),
            Loc("in"), Rule("or_test"), Opt(Rule("comp_iter")))) \
        (list_comp_for_action)
    """
    (2.6) gen_for: 'for' exprlist 'in' or_test [gen_iter]
    (2.7-) comp_for: 'for' exprlist 'in' or_test [comp_iter]
    """

    comp_if__26 = action(
        Seq(Loc("if"), Rule("old_test"), Opt(Rule("comp_iter")))) \
        (list_comp_if_action)
    """
    (2.6) gen_if: 'if' old_test [gen_iter]
    (2.7) comp_if: 'if' old_test [comp_iter]
    """

    comp_if__30 = action(
        Seq(Loc("if"), Rule("test_nocond"), Opt(Rule("comp_iter")))) \
        (list_comp_if_action)
    """
    (3.0-) comp_if: 'if' test_nocond [comp_iter]
    """

    testlist1 = action(List(Rule("test"), ",", trailing=False))(_wrap_tuple)
    """testlist1: test (',' test)*"""

    @action(Seq(Loc("yield"), Opt(Rule("testlist"))))
    def yield_expr__26(self, yield_loc, exprs):
        """(2.6, 2.7, 3.0, 3.1, 3.2) yield_expr: 'yield' [testlist]"""
        if exprs is not None:
            return ast.Yield(value=exprs,
                             yield_loc=yield_loc, loc=yield_loc.join(exprs.loc))
        else:
            return ast.Yield(value=None,
                             yield_loc=yield_loc, loc=yield_loc)

    @action(Seq(Loc("yield"), Opt(Rule("yield_arg"))))
    def yield_expr__33(self, yield_loc, arg):
        """(3.3-) yield_expr: 'yield' [yield_arg]"""
        if isinstance(arg, ast.YieldFrom):
            arg.yield_loc = yield_loc
            arg.loc = arg.loc.join(arg.yield_loc)
            return arg
        elif arg is not None:
            return ast.Yield(value=arg,
                             yield_loc=yield_loc, loc=yield_loc.join(arg.loc))
        else:
            return ast.Yield(value=None,
                             yield_loc=yield_loc, loc=yield_loc)

    @action(Seq(Loc("from"), Rule("test")))
    def yield_arg_1(self, from_loc, value):
        return ast.YieldFrom(value=value,
                             from_loc=from_loc, loc=from_loc.join(value.loc))

    yield_arg = Alt(yield_arg_1, Rule("testlist"))
    """(3.3-) yield_arg: 'from' test | testlist"""
