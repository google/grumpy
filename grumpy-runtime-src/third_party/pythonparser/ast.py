# encoding: utf-8

"""
The :mod:`ast` module contains the classes comprising the Python abstract syntax tree.

All attributes ending with ``loc`` contain instances of :class:`.source.Range`
or None. All attributes ending with ``_locs`` contain lists of instances of
:class:`.source.Range` or [].

The attribute ``loc``, present in every class except those inheriting :class:`boolop`,
has a special meaning: it encompasses the entire AST node, so that it is possible
to cut the range contained inside ``loc`` of a parsetree fragment and paste it
somewhere else without altering said parsetree fragment that.

The AST format for all supported versions is generally normalized to be a superset
of the native :mod:`..ast` module of the latest supported Python version.
In particular this affects:

    * :class:`With`: on 2.6-2.7 it uses the 3.0 format.
    * :class:`TryExcept` and :class:`TryFinally`: on 2.6-2.7 they're replaced with
      :class:`Try` from 3.0.
    * :class:`arguments`: on 2.6-3.1 it uses the 3.2 format, with dedicated
      :class:`arg` in ``vararg`` and ``kwarg`` slots.
"""

from __future__ import absolute_import, division, print_function, unicode_literals

# Location mixins

class commonloc(object):
    """
    A mixin common for all nodes.

    :cvar _locs: (tuple of strings)
        names of all attributes with location values

    :ivar loc: range encompassing all locations defined for this node
        or its children
    """

    _locs = ("loc",)

    def _reprfields(self):
        return self._fields + self._locs

    def __repr__(self):
        def value(name):
            try:
                loc = self.__dict__[name]
                if isinstance(loc, list):
                    return "[%s]" % (", ".join(map(repr, loc)))
                else:
                    return repr(loc)
            except:
                return "(!!!MISSING!!!)"
        fields = ", ".join(map(lambda name: "%s=%s" % (name, value(name)),
                           self._reprfields()))
        return "%s(%s)" % (self.__class__.__name__, fields)

    @property
    def lineno(self):
        return self.loc.line()

class keywordloc(commonloc):
    """
    A mixin common for all keyword statements, e.g. ``pass`` and ``yield expr``.

    :ivar keyword_loc: location of the keyword, e.g. ``yield``.
    """
    _locs = commonloc._locs + ("keyword_loc",)

class beginendloc(commonloc):
    """
    A mixin common for nodes with a opening and closing delimiters, e.g. tuples and lists.

    :ivar begin_loc: location of the opening delimiter, e.g. ``(``.
    :ivar end_loc: location of the closing delimiter, e.g. ``)``.
    """
    _locs = commonloc._locs + ("begin_loc", "end_loc")

# AST nodes

class AST(object):
    """
    An ancestor of all nodes.

    :cvar _fields: (tuple of strings)
        names of all attributes with semantic values
    """
    _fields = ()

    def __init__(self, **fields):
        for field in fields:
            setattr(self, field, fields[field])

class alias(AST, commonloc):
    """
    An import alias, e.g. ``x as y``.

    :ivar name: (string) value to import
    :ivar asname: (string) name to add to the environment
    :ivar name_loc: location of name
    :ivar as_loc: location of ``as``
    :ivar asname_loc: location of asname
    """
    _fields = ("name", "asname")
    _locs = commonloc._locs + ("name_loc", "as_loc", "asname_loc")

class arg(AST, commonloc):
    """
    A formal argument, e.g. in ``def f(x)`` or ``def f(x: T)``.

    :ivar arg: (string) argument name
    :ivar annotation: (:class:`AST`) type annotation, if any; **emitted since 3.0**
    :ivar arg_loc: location of argument name
    :ivar colon_loc: location of ``:``, if any; **emitted since 3.0**
    """
    _fields = ("arg", "annotation")
    _locs = commonloc._locs + ("arg_loc", "colon_loc")

class arguments(AST, beginendloc):
    """
    Function definition arguments, e.g. in ``def f(x, y=1, *z, **t)``.

    :ivar args: (list of :class:`arg`) regular formal arguments
    :ivar defaults: (list of :class:`AST`) values of default arguments
    :ivar vararg: (:class:`arg`) splat formal argument (if any), e.g. in ``*x``
    :ivar kwonlyargs: (list of :class:`arg`) keyword-only (post-\*) formal arguments;
        **emitted since 3.0**
    :ivar kw_defaults: (list of :class:`AST`) values of default keyword-only arguments;
        **emitted since 3.0**
    :ivar kwarg: (:class:`arg`) keyword splat formal argument (if any), e.g. in ``**x``
    :ivar star_loc: location of ``*``, if any
    :ivar dstar_loc: location of ``**``, if any
    :ivar equals_locs: locations of ``=``
    :ivar kw_equals_locs: locations of ``=`` of default keyword-only arguments;
        **emitted since 3.0**
    """
    _fields = ("args", "vararg", "kwonlyargs", "kwarg", "defaults", "kw_defaults")
    _locs = beginendloc._locs + ("star_loc", "dstar_loc", "equals_locs", "kw_equals_locs")

class boolop(AST, commonloc):
    """
    Base class for binary boolean operators.

    This class is unlike others in that it does not have the ``loc`` field.
    It serves only as an indicator of operation and corresponds to no source
    itself; locations are recorded in :class:`BoolOp`.
    """
    _locs = ()
class And(boolop):
    """The ``and`` operator."""
class Or(boolop):
    """The ``or`` operator."""

class cmpop(AST, commonloc):
    """Base class for comparison operators."""
class Eq(cmpop):
    """The ``==`` operator."""
class Gt(cmpop):
    """The ``>`` operator."""
class GtE(cmpop):
    """The ``>=`` operator."""
class In(cmpop):
    """The ``in`` operator."""
class Is(cmpop):
    """The ``is`` operator."""
class IsNot(cmpop):
    """The ``is not`` operator."""
class Lt(cmpop):
    """The ``<`` operator."""
class LtE(cmpop):
    """The ``<=`` operator."""
class NotEq(cmpop):
    """The ``!=`` (or deprecated ``<>``) operator."""
class NotIn(cmpop):
    """The ``not in`` operator."""

class comprehension(AST, commonloc):
    """
    A single ``for`` list comprehension clause.

    :ivar target: (assignable :class:`AST`) the variable(s) bound in comprehension body
    :ivar iter: (:class:`AST`) the expression being iterated
    :ivar ifs: (list of :class:`AST`) the ``if`` clauses
    :ivar for_loc: location of the ``for`` keyword
    :ivar in_loc: location of the ``in`` keyword
    :ivar if_locs: locations of ``if`` keywords
    """
    _fields = ("target", "iter", "ifs")
    _locs = commonloc._locs + ("for_loc", "in_loc", "if_locs")

class excepthandler(AST, commonloc):
    """Base class for the exception handler."""
class ExceptHandler(excepthandler):
    """
    An exception handler, e.g. ``except x as y:·  z``.

    :ivar type: (:class:`AST`) type of handled exception, if any
    :ivar name: (assignable :class:`AST` **until 3.0**, string **since 3.0**)
        variable bound to exception, if any
    :ivar body: (list of :class:`AST`) code to execute when exception is caught
    :ivar except_loc: location of ``except``
    :ivar as_loc: location of ``as``, if any
    :ivar name_loc: location of variable name
    :ivar colon_loc: location of ``:``
    """
    _fields = ("type", "name", "body")
    _locs = excepthandler._locs + ("except_loc", "as_loc", "name_loc", "colon_loc")

class expr(AST, commonloc):
    """Base class for expression nodes."""
class Attribute(expr):
    """
    An attribute access, e.g. ``x.y``.

    :ivar value: (:class:`AST`) left-hand side
    :ivar attr: (string) attribute name
    """
    _fields = ("value", "attr", "ctx")
    _locs = expr._locs + ("dot_loc", "attr_loc")
class BinOp(expr):
    """
    A binary operation, e.g. ``x + y``.

    :ivar left: (:class:`AST`) left-hand side
    :ivar op: (:class:`operator`) operator
    :ivar right: (:class:`AST`) right-hand side
    """
    _fields = ("left", "op", "right")
class BoolOp(expr):
    """
    A boolean operation, e.g. ``x and y``.

    :ivar op: (:class:`boolop`) operator
    :ivar values: (list of :class:`AST`) operands
    :ivar op_locs: locations of operators
    """
    _fields = ("op", "values")
    _locs = expr._locs + ("op_locs",)
class Call(expr, beginendloc):
    """
    A function call, e.g. ``f(x, y=1, *z, **t)``.

    :ivar func: (:class:`AST`) function to call
    :ivar args: (list of :class:`AST`) regular arguments
    :ivar keywords: (list of :class:`keyword`) keyword arguments
    :ivar starargs: (:class:`AST`) splat argument (if any), e.g. in ``*x``
    :ivar kwargs: (:class:`AST`) keyword splat argument (if any), e.g. in ``**x``
    :ivar star_loc: location of ``*``, if any
    :ivar dstar_loc: location of ``**``, if any
    """
    _fields = ("func", "args", "keywords", "starargs", "kwargs")
    _locs = beginendloc._locs + ("star_loc", "dstar_loc")
class Compare(expr):
    """
    A comparison operation, e.g. ``x < y`` or ``x < y > z``.

    :ivar left: (:class:`AST`) left-hand
    :ivar ops: (list of :class:`cmpop`) compare operators
    :ivar comparators: (list of :class:`AST`) compare values
    """
    _fields = ("left", "ops", "comparators")
class Dict(expr, beginendloc):
    """
    A dictionary, e.g. ``{x: y}``.

    :ivar keys: (list of :class:`AST`) keys
    :ivar values: (list of :class:`AST`) values
    :ivar colon_locs: locations of ``:``
    """
    _fields = ("keys", "values")
    _locs = beginendloc._locs + ("colon_locs",)
class DictComp(expr, beginendloc):
    """
    A list comprehension, e.g. ``{x: y for x,y in z}``.

    **Emitted since 2.7.**

    :ivar key: (:class:`AST`) key part of comprehension body
    :ivar value: (:class:`AST`) value part of comprehension body
    :ivar generators: (list of :class:`comprehension`) ``for`` clauses
    :ivar colon_loc: location of ``:``
    """
    _fields = ("key", "value", "generators")
    _locs = beginendloc._locs + ("colon_loc",)
class Ellipsis(expr):
    """The ellipsis, e.g. in ``x[...]``."""
class GeneratorExp(expr, beginendloc):
    """
    A generator expression, e.g. ``(x for x in y)``.

    :ivar elt: (:class:`AST`) expression body
    :ivar generators: (list of :class:`comprehension`) ``for`` clauses
    """
    _fields = ("elt", "generators")
class IfExp(expr):
    """
    A conditional expression, e.g. ``x if y else z``.

    :ivar test: (:class:`AST`) condition
    :ivar body: (:class:`AST`) value if true
    :ivar orelse: (:class:`AST`) value if false
    :ivar if_loc: location of ``if``
    :ivar else_loc: location of ``else``
    """
    _fields = ("test", "body", "orelse")
    _locs = expr._locs + ("if_loc", "else_loc")
class Lambda(expr):
    """
    A lambda expression, e.g. ``lambda x: x*x``.

    :ivar args: (:class:`arguments`) arguments
    :ivar body: (:class:`AST`) body
    :ivar lambda_loc: location of ``lambda``
    :ivar colon_loc: location of ``:``
    """
    _fields = ("args", "body")
    _locs = expr._locs + ("lambda_loc", "colon_loc")
class List(expr, beginendloc):
    """
    A list, e.g. ``[x, y]``.

    :ivar elts: (list of :class:`AST`) elements
    """
    _fields = ("elts", "ctx")
class ListComp(expr, beginendloc):
    """
    A list comprehension, e.g. ``[x for x in y]``.

    :ivar elt: (:class:`AST`) comprehension body
    :ivar generators: (list of :class:`comprehension`) ``for`` clauses
    """
    _fields = ("elt", "generators")
class Name(expr):
    """
    An identifier, e.g. ``x``.

    :ivar id: (string) name
    """
    _fields = ("id", "ctx")
class NameConstant(expr):
    """
    A named constant, e.g. ``None``.

    :ivar value: Python value, one of ``None``, ``True`` or ``False``
    """
    _fields = ("value",)
class Num(expr):
    """
    An integer, floating point or complex number, e.g. ``1``, ``1.0`` or ``1.0j``.

    :ivar n: (int, float or complex) value
    """
    _fields = ("n",)
class Repr(expr, beginendloc):
    """
    A repr operation, e.g. ``\`x\```

    **Emitted until 3.0.**

    :ivar value: (:class:`AST`) value
    """
    _fields = ("value",)
class Set(expr, beginendloc):
    """
    A set, e.g. ``{x, y}``.

    **Emitted since 2.7.**

    :ivar elts: (list of :class:`AST`) elements
    """
    _fields = ("elts",)
class SetComp(expr, beginendloc):
    """
    A set comprehension, e.g. ``{x for x in y}``.

    **Emitted since 2.7.**

    :ivar elt: (:class:`AST`) comprehension body
    :ivar generators: (list of :class:`comprehension`) ``for`` clauses
    """
    _fields = ("elt", "generators")
class Str(expr, beginendloc):
    """
    A string, e.g. ``"x"``.

    :ivar s: (string) value
    """
    _fields = ("s",)
class Starred(expr):
    """
    A starred expression, e.g. ``*x`` in ``*x, y = z``.

    :ivar value: (:class:`AST`) expression
    :ivar star_loc: location of ``*``
    """
    _fields = ("value", "ctx")
    _locs = expr._locs + ("star_loc",)
class Subscript(expr, beginendloc):
    """
    A subscript operation, e.g. ``x[1]``.

    :ivar value: (:class:`AST`) object being sliced
    :ivar slice: (:class:`slice`) slice
    """
    _fields = ("value", "slice", "ctx")
class Tuple(expr, beginendloc):
    """
    A tuple, e.g. ``(x,)`` or ``x,y``.

    :ivar elts: (list of nodes) elements
    """
    _fields = ("elts", "ctx")
class UnaryOp(expr):
    """
    An unary operation, e.g. ``+x``.

    :ivar op: (:class:`unaryop`) operator
    :ivar operand: (:class:`AST`) operand
    """
    _fields = ("op", "operand")
class Yield(expr):
    """
    A yield expression, e.g. ``yield x``.

    :ivar value: (:class:`AST`) yielded value
    :ivar yield_loc: location of ``yield``
    """
    _fields = ("value",)
    _locs = expr._locs + ("yield_loc",)
class YieldFrom(expr):
    """
    A yield from expression, e.g. ``yield from x``.

    :ivar value: (:class:`AST`) yielded value
    :ivar yield_loc: location of ``yield``
    :ivar from_loc: location of ``from``
    """
    _fields = ("value",)
    _locs = expr._locs + ("yield_loc", "from_loc")

# expr_context
#     AugLoad
#     AugStore
#     Del
#     Load
#     Param
#     Store

class keyword(AST, commonloc):
    """
    A keyword actual argument, e.g. in ``f(x=1)``.

    :ivar arg: (string) name
    :ivar value: (:class:`AST`) value
    :ivar equals_loc: location of ``=``
    """
    _fields = ("arg", "value")
    _locs = commonloc._locs + ("arg_loc", "equals_loc")

class mod(AST, commonloc):
    """Base class for modules (groups of statements)."""
    _fields = ("body",)
class Expression(mod):
    """A group of statements parsed as if for :func:`eval`."""
class Interactive(mod):
    """A group of statements parsed as if it was REPL input."""
class Module(mod):
    """A group of statements parsed as if it was a file."""

class operator(AST, commonloc):
    """Base class for numeric binary operators."""
class Add(operator):
    """The ``+`` operator."""
class BitAnd(operator):
    """The ``&`` operator."""
class BitOr(operator):
    """The ``|`` operator."""
class BitXor(operator):
    """The ``^`` operator."""
class Div(operator):
    """The ``\\`` operator."""
class FloorDiv(operator):
    """The ``\\\\`` operator."""
class LShift(operator):
    """The ``<<`` operator."""
class MatMult(operator):
    """The ``@`` operator."""
class Mod(operator):
    """The ``%`` operator."""
class Mult(operator):
    """The ``*`` operator."""
class Pow(operator):
    """The ``**`` operator."""
class RShift(operator):
    """The ``>>`` operator."""
class Sub(operator):
    """The ``-`` operator."""

class slice(AST, commonloc):
    """Base class for slice operations."""
class ExtSlice(slice):
    """
    The multiple slice, e.g. in ``x[0:1, 2:3]``.
    Note that multiple slices with only integer indexes
    will appear as instances of :class:`Index`.

    :ivar dims: (:class:`slice`) sub-slices
    """
    _fields = ("dims",)
class Index(slice):
    """
    The index, e.g. in ``x[1]`` or ``x[1, 2]``.

    :ivar value: (:class:`AST`) index
    """
    _fields = ("value",)
class Slice(slice):
    """
    The slice, e.g. in ``x[0:1]`` or ``x[0:1:2]``.

    :ivar lower: (:class:`AST`) lower bound, if any
    :ivar upper: (:class:`AST`) upper bound, if any
    :ivar step: (:class:`AST`) iteration step, if any
    :ivar bound_colon_loc: location of first semicolon
    :ivar step_colon_loc: location of second semicolon, if any
    """
    _fields = ("lower", "upper", "step")
    _locs = slice._locs + ("bound_colon_loc", "step_colon_loc")

class stmt(AST, commonloc):
    """Base class for statement nodes."""
class Assert(stmt, keywordloc):
    """
    The ``assert x, msg`` statement.

    :ivar test: (:class:`AST`) condition
    :ivar msg: (:class:`AST`) message, if any
    """
    _fields = ("test", "msg")
class Assign(stmt):
    """
    The ``=`` statement, e.g. in ``x = 1`` or ``x = y = 1``.

    :ivar targets: (list of assignable :class:`AST`) left-hand sides
    :ivar value: (:class:`AST`) right-hand side
    :ivar op_locs: location of equality signs corresponding to ``targets``
    """
    _fields = ("targets", "value")
    _locs = stmt._locs + ("op_locs",)
class AugAssign(stmt):
    """
    The operator-assignment statement, e.g. ``+=``.

    :ivar target: (assignable :class:`AST`) left-hand side
    :ivar op: (:class:`operator`) operator
    :ivar value: (:class:`AST`) right-hand side
    """
    _fields = ("target", "op", "value")
class Break(stmt, keywordloc):
    """The ``break`` statement."""
class ClassDef(stmt, keywordloc):
    """
    The ``class x(z, y):·  t`` (2.6) or
    ``class x(y, z=1, *t, **u):·  v`` (3.0) statement.

    :ivar name: (string) name
    :ivar bases: (list of :class:`AST`) base classes
    :ivar keywords: (list of :class:`keyword`) keyword arguments; **emitted since 3.0**
    :ivar starargs: (:class:`AST`) splat argument (if any), e.g. in ``*x``; **emitted since 3.0**
    :ivar kwargs: (:class:`AST`) keyword splat argument (if any), e.g. in ``**x``; **emitted since 3.0**
    :ivar body: (list of :class:`AST`) body
    :ivar decorator_list: (list of :class:`AST`) decorators
    :ivar keyword_loc: location of ``class``
    :ivar name_loc: location of name
    :ivar lparen_loc: location of ``(``, if any
    :ivar star_loc: location of ``*``, if any; **emitted since 3.0**
    :ivar dstar_loc: location of ``**``, if any; **emitted since 3.0**
    :ivar rparen_loc: location of ``)``, if any
    :ivar colon_loc: location of ``:``
    :ivar at_locs: locations of decorator ``@``
    """
    _fields = ("name", "bases", "keywords", "starargs", "kwargs", "body", "decorator_list")
    _locs = keywordloc._locs + ("name_loc", "lparen_loc", "star_loc", "dstar_loc", "rparen_loc",
                                "colon_loc", "at_locs")
class Continue(stmt, keywordloc):
    """The ``continue`` statement."""
class Delete(stmt, keywordloc):
    """
    The ``del x, y`` statement.

    :ivar targets: (list of :class:`Name`)
    """
    _fields = ("targets",)
class Exec(stmt, keywordloc):
    """
    The ``exec code in locals, globals`` statement.

    **Emitted until 3.0.**

    :ivar body: (:class:`AST`) code
    :ivar locals: (:class:`AST`) locals
    :ivar globals: (:class:`AST`) globals
    :ivar keyword_loc: location of ``exec``
    :ivar in_loc: location of ``in``
    """
    _fields = ("body", "locals", "globals")
    _locs = keywordloc._locs + ("in_loc",)
class Expr(stmt):
    """
    An expression in statement context. The value of expression is discarded.

    :ivar value: (:class:`expr`) value
    """
    _fields = ("value",)
class For(stmt, keywordloc):
    """
    The ``for x in y:·  z·else:·  t`` statement.

    :ivar target: (assignable :class:`AST`) loop variable
    :ivar iter: (:class:`AST`) loop collection
    :ivar body: (list of :class:`AST`) code for every iteration
    :ivar orelse: (list of :class:`AST`) code if empty
    :ivar keyword_loc: location of ``for``
    :ivar in_loc: location of ``in``
    :ivar for_colon_loc: location of colon after ``for``
    :ivar else_loc: location of ``else``, if any
    :ivar else_colon_loc: location of colon after ``else``, if any
    """
    _fields = ("target", "iter", "body", "orelse")
    _locs = keywordloc._locs + ("in_loc", "for_colon_loc", "else_loc", "else_colon_loc")
class FunctionDef(stmt, keywordloc):
    """
    The ``def f(x):·  y`` (2.6) or ``def f(x) -> t:·  y`` (3.0) statement.

    :ivar name: (string) name
    :ivar args: (:class:`arguments`) formal arguments
    :ivar returns: (:class:`AST`) return type annotation; **emitted since 3.0**
    :ivar body: (list of :class:`AST`) body
    :ivar decorator_list: (list of :class:`AST`) decorators
    :ivar keyword_loc: location of ``def``
    :ivar name_loc: location of name
    :ivar arrow_loc: location of ``->``, if any; **emitted since 3.0**
    :ivar colon_loc: location of ``:``, if any
    :ivar at_locs: locations of decorator ``@``
    """
    _fields = ("name", "args", "returns", "body", "decorator_list")
    _locs = keywordloc._locs + ("name_loc", "arrow_loc", "colon_loc", "at_locs")
class Global(stmt, keywordloc):
    """
    The ``global x, y`` statement.

    :ivar names: (list of string) names
    :ivar name_locs: locations of names
    """
    _fields = ("names",)
    _locs = keywordloc._locs + ("name_locs",)
class If(stmt, keywordloc):
    """
    The ``if x:·  y·else:·  z`` or ``if x:·  y·elif: z·  t`` statement.

    :ivar test: (:class:`AST`) condition
    :ivar body: (list of :class:`AST`) code if true
    :ivar orelse: (list of :class:`AST`) code if false
    :ivar if_colon_loc: location of colon after ``if`` or ``elif``
    :ivar else_loc: location of ``else``, if any
    :ivar else_colon_loc: location of colon after ``else``, if any
    """
    _fields = ("test", "body", "orelse")
    _locs = keywordloc._locs + ("if_colon_loc", "else_loc", "else_colon_loc")
class Import(stmt, keywordloc):
    """
    The ``import x, y`` statement.

    :ivar names: (list of :class:`alias`) names
    """
    _fields = ("names",)
class ImportFrom(stmt, keywordloc):
    """
    The ``from ...x import y, z`` or ``from x import (y, z)`` or
    ``from x import *`` statement.

    :ivar names: (list of :class:`alias`) names
    :ivar module: (string) module name, if any
    :ivar level: (integer) amount of dots before module name
    :ivar keyword_loc: location of ``from``
    :ivar dots_loc: location of dots, if any
    :ivar module_loc: location of module name, if any
    :ivar import_loc: location of ``import``
    :ivar lparen_loc: location of ``(``, if any
    :ivar rparen_loc: location of ``)``, if any
    """
    _fields = ("names", "module", "level")
    _locs = keywordloc._locs + ("dots_loc", "module_loc", "import_loc", "lparen_loc", "rparen_loc")
class Nonlocal(stmt, keywordloc):
    """
    The ``nonlocal x, y`` statement.

    **Emitted since 3.0.**

    :ivar names: (list of string) names
    :ivar name_locs: locations of names
    """
    _fields = ("names",)
    _locs = keywordloc._locs + ("name_locs",)
class Pass(stmt, keywordloc):
    """The ``pass`` statement."""
class Print(stmt, keywordloc):
    """
    The ``print >>x, y, z,`` statement.

    **Emitted until 3.0 or until print_function future flag is activated.**

    :ivar dest: (:class:`AST`) destination stream, if any
    :ivar values: (list of :class:`AST`) values to print
    :ivar nl: (boolean) whether to print newline after values
    :ivar dest_loc: location of ``>>``
    """
    _fields = ("dest", "values", "nl")
    _locs = keywordloc._locs + ("dest_loc",)
class Raise(stmt, keywordloc):
    """
    The ``raise exc, arg, traceback`` (2.6) or
    or ``raise exc from cause`` (3.0) statement.

    :ivar exc: (:class:`AST`) exception type or instance
    :ivar cause: (:class:`AST`) cause of exception, if any; **emitted since 3.0**
    :ivar inst: (:class:`AST`) exception instance or argument list, if any; **emitted until 3.0**
    :ivar tback: (:class:`AST`) traceback, if any; **emitted until 3.0**
    :ivar from_loc: location of ``from``, if any; **emitted since 3.0**
    """
    _fields = ("exc", "cause", "inst", "tback")
    _locs = keywordloc._locs + ("from_loc",)
class Return(stmt, keywordloc):
    """
    The ``return x`` statement.

    :ivar value: (:class:`AST`) return value, if any
    """
    _fields = ("value",)
class Try(stmt, keywordloc):
    """
    The ``try:·  x·except y:·  z·else:·  t`` or
    ``try:·  x·finally:·  y`` statement.

    :ivar body: (list of :class:`AST`) code to try
    :ivar handlers: (list of :class:`ExceptHandler`) exception handlers
    :ivar orelse: (list of :class:`AST`) code if no exception
    :ivar finalbody: (list of :class:`AST`) code to finalize
    :ivar keyword_loc: location of ``try``
    :ivar try_colon_loc: location of ``:`` after ``try``
    :ivar else_loc: location of ``else``
    :ivar else_colon_loc: location of ``:`` after ``else``
    :ivar finally_loc: location of ``finally``
    :ivar finally_colon_loc: location of ``:`` after ``finally``
    """
    _fields = ("body", "handlers", "orelse", "finalbody")
    _locs = keywordloc._locs + ("try_colon_loc", "else_loc", "else_colon_loc",
                                "finally_loc", "finally_colon_loc",)
class While(stmt, keywordloc):
    """
    The ``while x:·  y·else:·  z`` statement.

    :ivar test: (:class:`AST`) condition
    :ivar body: (list of :class:`AST`) code for every iteration
    :ivar orelse: (list of :class:`AST`) code if empty
    :ivar keyword_loc: location of ``while``
    :ivar while_colon_loc: location of colon after ``while``
    :ivar else_loc: location of ``else``, if any
    :ivar else_colon_loc: location of colon after ``else``, if any
    """
    _fields = ("test", "body", "orelse")
    _locs = keywordloc._locs + ("while_colon_loc", "else_loc", "else_colon_loc")
class With(stmt, keywordloc):
    """
    The ``with x as y:·  z`` statement.

    :ivar items: (list of :class:`withitem`) bindings
    :ivar body: (:class:`AST`) body
    :ivar keyword_loc: location of ``with``
    :ivar colon_loc: location of ``:``
    """
    _fields = ("items", "body")
    _locs = keywordloc._locs + ("colon_loc",)

class unaryop(AST, commonloc):
    """Base class for unary numeric and boolean operators."""
class Invert(unaryop):
    """The ``~`` operator."""
class Not(unaryop):
    """The ``not`` operator."""
class UAdd(unaryop):
    """The unary ``+`` operator."""
class USub(unaryop):
    """The unary ``-`` operator."""

class withitem(AST, commonloc):
    """
    The ``x as y`` clause in ``with x as y:``.

    :ivar context_expr: (:class:`AST`) context
    :ivar optional_vars: (assignable :class:`AST`) context binding, if any
    :ivar as_loc: location of ``as``, if any
    """
    _fields = ("context_expr", "optional_vars")
    _locs = commonloc._locs + ("as_loc",)
