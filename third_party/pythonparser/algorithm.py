"""
The :mod:`Diagnostic` module provides several commonly useful
algorithms that operate on abstract syntax trees.
"""

from __future__ import absolute_import, division, print_function, unicode_literals
from . import ast

class Visitor:
    """
    A node visitor base class that does a traversal
    of the abstract syntax tree.

    This class is meant to be subclassed, with the subclass adding
    visitor methods. The visitor method should call ``self.generic_visit(node)``
    to continue the traversal; this allows to perform arbitrary
    actions both before and after traversing the children of a node.

    The visitor methods for the nodes are ``'visit_'`` +
    class name of the node.  So a `Try` node visit function would
    be `visit_Try`.
    """

    def generic_visit(self, node):
        """Called if no explicit visitor function exists for a node."""
        for field_name in node._fields:
            self.visit(getattr(node, field_name))

    def _visit_one(self, node):
        visit_attr = "visit_" + type(node).__name__
        if hasattr(self, visit_attr):
            return getattr(self, visit_attr)(node)
        else:
            return self.generic_visit(node)

    def visit(self, obj):
        """Visit a node or a list of nodes. Other values are ignored"""
        if isinstance(obj, list):
            return [self.visit(elt) for elt in obj]
        elif isinstance(obj, ast.AST):
            return self._visit_one(obj)

class Transformer:
    """
    A node transformer base class that does a post-order traversal
    of the abstract syntax tree while allowing to replace or remove
    the nodes being traversed.

    The return value of the visitor methods is used to replace or remove
    the old node.  If the return value of the visitor method is ``None``,
    the node will be removed from its location, otherwise it is replaced
    with the return value.  The return value may be the original node
    in which case no replacement takes place.

    This class is meant to be subclassed, with the subclass adding
    visitor methods. The visitor method should call ``self.generic_visit(node)``
    to continue the traversal; this allows to perform arbitrary
    actions both before and after traversing the children of a node.

    The visitor methods for the nodes are ``'visit_'`` +
    class name of the node.  So a `Try` node visit function would
    be `visit_Try`.
    """

    def generic_visit(self, node):
        """Called if no explicit visitor function exists for a node."""
        for field_name in node._fields:
            setattr(node, field_name, self.visit(getattr(node, field_name)))
        return node

    def _visit_one(self, node):
        visit_attr = "visit_" + type(node).__name__
        if hasattr(self, visit_attr):
            return getattr(self, visit_attr)(node)
        else:
            return self.generic_visit(node)

    def visit(self, obj):
        """Visit a node or a list of nodes. Other values are ignored"""
        if isinstance(obj, list):
            return list(filter(lambda x: x is not None, map(self.visit, obj)))
        elif isinstance(obj, ast.AST):
            return self._visit_one(obj)
        else:
            return obj

def compare(left, right, compare_locs=False):
    """
    An AST comparison function. Returns ``True`` if all fields in
    ``left`` are equal to fields in ``right``; if ``compare_locs`` is
    true, all locations should match as well.
    """
    if type(left) != type(right):
        return False

    if isinstance(left, ast.AST):
        for field in left._fields:
            if not compare(getattr(left, field), getattr(right, field)):
                return False

        if compare_locs:
            for loc in left._locs:
                if getattr(left, loc) != getattr(right, loc):
                    return False

        return True
    elif isinstance(left, list):
        if len(left) != len(right):
            return False

        for left_elt, right_elt in zip(left, right):
            if not compare(left_elt, right_elt):
                return False

        return True
    else:
        return left == right
