#  Author:      Fred L. Drake, Jr.
#               fdrake@acm.org
#
#  This is a simple little module I wrote to make life easier.  I didn't
#  see anything quite like it in the library, though I may have overlooked
#  something.  I wrote this when I was trying to read some heavily nested
#  tuples with fairly non-descriptive content.  This is modeled very much
#  after Lisp/Scheme - style pretty-printing of lists.  If you find it
#  useful, thank small children who sleep at night.

"""Support to pretty-print lists, tuples, & dictionaries recursively.

Very simple, but useful, especially in debugging data structures.

Classes
-------

PrettyPrinter()
    Handle pretty-printing operations onto a stream using a configured
    set of formatting parameters.

Functions
---------

pformat()
    Format a Python object into a pretty-printed representation.

pprint()
    Pretty-print a Python object to a stream [default is sys.stdout].

saferepr()
    Generate a 'standard' repr()-like value, but protect against recursive
    data structures.

"""

import sys as _sys
import warnings
import StringIO
_StringIO = StringIO.StringIO

# try:
#     from cStringIO import StringIO as _StringIO
# except ImportError:
#     from StringIO import StringIO as _StringIO

__all__ = ["pprint","pformat","isreadable","isrecursive","saferepr",
           "PrettyPrinter"]

# cache these for faster access:
_commajoin = ", ".join
_id = id
_len = len
_type = type


def pprint(o, stream=None, indent=1, width=80, depth=None):
    """Pretty-print a Python o to a stream [default is sys.stdout]."""
    printer = PrettyPrinter(
        stream=stream, indent=indent, width=width, depth=depth)
    printer.pprint(o)

def pformat(o, indent=1, width=80, depth=None):
    """Format a Python o into a pretty-printed representation."""
    return PrettyPrinter(indent=indent, width=width, depth=depth).pformat(o)

def saferepr(o):
    """Version of repr() which can handle recursive data structures."""
    return _safe_repr(o, {}, None, 0)[0]

def isreadable(o):
    """Determine if saferepr(o) is readable by eval()."""
    return _safe_repr(o, {}, None, 0)[1]

def isrecursive(o):
    """Determine if o requires a recursive representation."""
    return _safe_repr(o, {}, None, 0)[2]

def _sorted(iterable):
    with warnings.catch_warnings():
        if _sys.py3kwarning:
            warnings.filterwarnings("ignore", "comparing unequal types "
                                    "not supported", DeprecationWarning)
        return sorted(iterable)

class PrettyPrinter(object):
    def __init__(self, indent=1, width=80, depth=None, stream=None):
        """Handle pretty printing operations onto a stream using a set of
        configured parameters.

        indent
            Number of spaces to indent for each level of nesting.

        width
            Attempted maximum number of columns in the output.

        depth
            The maximum depth to print out nested structures.

        stream
            The desired output stream.  If omitted (or false), the standard
            output stream available at construction will be used.

        """
        indent = int(indent)
        width = int(width)
        assert indent >= 0, "indent must be >= 0"
        assert depth is None or depth > 0, "depth must be > 0"
        assert width, "width must be != 0"
        self._depth = depth
        self._indent_per_level = indent
        self._width = width
        if stream is not None:
            self._stream = stream
        else:
            self._stream = _sys.stdout

    def pprint(self, o):
        self._format(o, self._stream, 0, 0, {}, 0)
        self._stream.write("\n")

    def pformat(self, o):
        sio = _StringIO()
        self._format(o, sio, 0, 0, {}, 0)
        return sio.getvalue()

    def isrecursive(self, o):
        return self.format(o, {}, 0, 0)[2]

    def isreadable(self, o):
        s, readable, recursive = self.format(o, {}, 0, 0)
        return readable and not recursive

    def _format(self, o, stream, indent, allowance, context, level):
        level = level + 1
        objid = _id(o)
        if objid in context:
            stream.write(_recursion(o))
            self._recursive = True
            self._readable = False
            return
        rep = self._repr(o, context, level - 1)
        typ = _type(o)
        sepLines = _len(rep) > (self._width - 1 - indent - allowance)
        write = stream.write

        if self._depth and level > self._depth:
            write(rep)
            return

        r = getattr(typ, "__repr__", None)
        if issubclass(typ, dict):# and r == dict.__repr__:
            write('{')
            if self._indent_per_level > 1:
                write((self._indent_per_level - 1) * ' ')
            length = _len(o)
            if length:
                context[objid] = 1
                indent = indent + self._indent_per_level
                items = _sorted(o.items())
                key, ent = items[0]
                rep = self._repr(key, context, level)
                write(rep)
                write(': ')
                self._format(ent, stream, indent + _len(rep) + 2,
                              allowance + 1, context, level)
                if length > 1:
                    for key, ent in items[1:]:
                        rep = self._repr(key, context, level)
                        if sepLines:
                            write(',\n%s%s: ' % (' '*indent, rep))
                        else:
                            write(', %s: ' % rep)
                        self._format(ent, stream, indent + _len(rep) + 2,
                                      allowance + 1, context, level)
                indent = indent - self._indent_per_level
                del context[objid]
            write('}')
            return

        # if ((issubclass(typ, list) and r == list.__repr__) or
        #     (issubclass(typ, tuple) and r == tuple.__repr__) or
        #     (issubclass(typ, set) and r == set.__repr__) or
        #     (issubclass(typ, frozenset) and r == frozenset.__repr__)
        if ((issubclass(typ, list)) or
            (issubclass(typ, tuple)) or
            (issubclass(typ, set)) or
            (issubclass(typ, frozenset))
           ):
            length = _len(o)
            if issubclass(typ, list):
                write('[')
                endchar = ']'
            elif issubclass(typ, tuple):
                write('(')
                endchar = ')'
            else:
                if not length:
                    write(rep)
                    return
                write(typ.__name__)
                write('([')
                endchar = '])'
                indent += len(typ.__name__) + 1
                o = _sorted(o)
            if self._indent_per_level > 1 and sepLines:
                write((self._indent_per_level - 1) * ' ')
            if length:
                context[objid] = 1
                indent = indent + self._indent_per_level
                self._format(o[0], stream, indent, allowance + 1,
                             context, level)
                if length > 1:
                    for ent in o[1:]:
                        if sepLines:
                            write(',\n' + ' '*indent)
                        else:
                            write(', ')
                        self._format(ent, stream, indent,
                                      allowance + 1, context, level)
                indent = indent - self._indent_per_level
                del context[objid]
            if issubclass(typ, tuple) and length == 1:
                write(',')
            write(endchar)
            return

        write(rep)

    def _repr(self, o, context, level):
        repr1, readable, recursive = self.format(o, dict(context),
                                                self._depth, level)
        if not readable:
            self._readable = False
        if recursive:
            self._recursive = True
        return repr1

    def format(self, o, context, maxlevels, level):
        """Format o for a specific context, returning a string
        and flags indicating whether the representation is 'readable'
        and whether the o represents a recursive construct.
        """
        return _safe_repr(o, context, maxlevels, level)


# Return triple (repr_string, isreadable, isrecursive).

def _safe_repr(o, context, maxlevels, level):
    typ = _type(o)
    if typ is str:
        # if 'locale' not in _sys.modules:
        #     return repr(o), True, False
        if "'" in o and '"' not in o:
            closure = '"'
            quotes = {'"': '\\"'}
        else:
            closure = "'"
            quotes = {"'": "\\'"}
        qget = quotes.get
        sio = _StringIO()
        write = sio.write
        for char in o:
            # if char.isalpha():
            if 'a' <= char <= 'z' or 'A' <= char <= 'Z':
                write(char)
            else:
                write(qget(char, repr(char)[1:-1]))
        return ("%s%s%s" % (closure, sio.getvalue(), closure)), True, False

    r = getattr(typ, "__repr__", None)
    if issubclass(typ, dict):# and r == dict.__repr__:
        if not o:
            return "{}", True, False
        objid = _id(o)
        if maxlevels and level >= maxlevels:
            return "{...}", False, objid in context
        if objid in context:
            return _recursion(o), False, True
        context[objid] = 1
        readable = True
        recursive = False
        components = []
        append = components.append
        level += 1
        saferepr = _safe_repr
        for k, v in _sorted(o.items()):
            krepr, kreadable, krecur = saferepr(k, context, maxlevels, level)
            vrepr, vreadable, vrecur = saferepr(v, context, maxlevels, level)
            append("%s: %s" % (krepr, vrepr))
            readable = readable and kreadable and vreadable
            if krecur or vrecur:
                recursive = True
        del context[objid]
        return "{%s}" % _commajoin(components), readable, recursive

    # if (issubclass(typ, list) and r == list.__repr__) or \
    #    (issubclass(typ, tuple) and r == tuple.__repr__):
    if (issubclass(typ, list)) or \
       (issubclass(typ, tuple)):
        if issubclass(typ, list):
            if not o:
                return "[]", True, False
            format = "[%s]"
        elif _len(o) == 1:
            format = "(%s,)"
        else:
            if not o:
                return "()", True, False
            format = "(%s)"
        objid = _id(o)
        if maxlevels and level >= maxlevels:
            return format % "...", False, objid in context
        if objid in context:
            return _recursion(o), False, True
        context[objid] = 1
        readable = True
        recursive = False
        components = []
        append = components.append
        level += 1
        for x in o:
            orepr, oreadable, orecur = _safe_repr(x, context, maxlevels, level)
            append(orepr)
            if not oreadable:
                readable = False
            if orecur:
                recursive = True
        del context[objid]
        return format % _commajoin(components), readable, recursive

    rep = repr(o)
    return rep, (rep and not rep.startswith('<')), False


def _recursion(o):
    return ("<Recursion on %s with id=%s>"
            % (_type(o).__name__, _id(o)))


# def _perfcheck(o=None):
#     import time
#     if o is None:
#         o = [("string", (1, 2), [3, 4], {5: 6, 7: 8})] * 100000
#     p = PrettyPrinter()
#     t1 = time.time()
#     _safe_repr(o, {}, None, 0)
#     t2 = time.time()
#     p.pformat(o)
#     t3 = time.time()
#     print "_safe_repr:", t2 - t1
#     print "pformat:", t3 - t2

# if __name__ == "__main__":
#     _perfcheck()
