"""
The :mod:`Diagnostic` module concerns itself with processing
and presentation of diagnostic messages.
"""

from __future__ import absolute_import, division, print_function, unicode_literals
from functools import reduce
from contextlib import contextmanager
import sys, re

class Diagnostic:
    """
    A diagnostic message highlighting one or more locations
    in a single source buffer.

    :ivar level: (one of ``LEVELS``) severity level
    :ivar reason: (format string) diagnostic message
    :ivar arguments: (dictionary) substitutions for ``reason``
    :ivar location: (:class:`pythonparser.source.Range`) most specific
        location of the problem
    :ivar highlights: (list of :class:`pythonparser.source.Range`)
        secondary locations related to the problem that are
        likely to be on the same line
    :ivar notes: (list of :class:`Diagnostic`)
        secondary diagnostics highlighting relevant source
        locations that are unlikely to be on the same line
    """

    LEVELS = ["note", "warning", "error", "fatal"]
    """
    Available diagnostic levels:
        * ``fatal`` indicates an unrecoverable error.
        * ``error`` indicates an error that leaves a possibility of
          processing more code, e.g. a recoverable parsing error.
        * ``warning`` indicates a potential problem.
        * ``note`` level diagnostics do not appear by itself,
          but are attached to other diagnostics to refer to
          and describe secondary source locations.
    """

    def __init__(self, level, reason, arguments, location,
                 highlights=None, notes=None):
        if level not in self.LEVELS:
            raise ValueError("level must be one of Diagnostic.LEVELS")

        if highlights is None:
            highlights = []
        if notes is None:
            notes = []

        if len(set(map(lambda x: x.source_buffer,
                       [location] + highlights))) > 1:
            raise ValueError("location and highlights must refer to the same source buffer")

        self.level, self.reason, self.arguments = \
            level, reason, arguments
        self.location, self.highlights, self.notes = \
            location, highlights, notes

    def message(self):
        """
        Returns the formatted message.
        """
        return self.reason.format(**self.arguments)

    def render(self, only_line=False, colored=False):
        """
        Returns the human-readable location of the diagnostic in the source,
        the formatted message, the source line corresponding
        to ``location`` and a line emphasizing the problematic
        locations in the source line using ASCII art, as a list of lines.
        Appends the result of calling :meth:`render` on ``notes``, if any.

        For example: ::

            <input>:1:8-9: error: cannot add integer and string
            x + (1 + "a")
                 ~ ^ ~~~

        :param only_line: (bool) If true, only print line number, not line and column range
        """
        source_line = self.location.source_line().rstrip("\n")
        highlight_line = bytearray(re.sub(r"[^\t]", " ", source_line), "utf-8")

        for hilight in self.highlights:
            if hilight.line() == self.location.line():
                lft, rgt = hilight.column_range()
                highlight_line[lft:rgt] = bytearray("~", "utf-8") * (rgt - lft)

        lft, rgt = self.location.column_range()
        if rgt == lft: # Expand zero-length ranges to one ^
            rgt = lft + 1
        highlight_line[lft:rgt] = bytearray("^", "utf-8") * (rgt - lft)

        if only_line:
            location = "%s:%s" % (self.location.source_buffer.name, self.location.line())
        else:
            location = str(self.location)

        notes = list(self.notes)
        if self.level != "note":
            expanded_location = self.location.expanded_from
            while expanded_location is not None:
                notes.insert(0, Diagnostic("note",
                    "expanded from here", {},
                    self.location.expanded_from))
                expanded_location = expanded_location.expanded_from

        rendered_notes = reduce(list.__add__, [note.render(only_line, colored)
                                               for note in notes], [])
        if colored:
            if self.level in ("error", "fatal"):
                level_color = 31 # red
            elif self.level == "warning":
                level_color = 35 # magenta
            else: # level == "note"
                level_color = 30 # gray
            return [
                "\x1b[1;37m{}: \x1b[{}m{}:\x1b[37m {}\x1b[0m".
                    format(location, level_color, self.level, self.message()),
                source_line,
                "\x1b[1;32m{}\x1b[0m".format(highlight_line.decode("utf-8"))
            ] + rendered_notes
        else:
            return [
                "{}: {}: {}".format(location, self.level, self.message()),
                source_line,
                highlight_line.decode("utf-8")
            ] + rendered_notes


class Error(Exception):
    """
    :class:`Error` is an exception which carries a :class:`Diagnostic`.

    :ivar diagnostic: (:class:`Diagnostic`) the diagnostic
    """
    def __init__(self, diagnostic):
        self.diagnostic = diagnostic

    def __str__(self):
        return "\n".join(self.diagnostic.render())

class Engine:
    """
    :class:`Engine` is a single point through which diagnostics from
    lexer, parser and any AST consumer are dispatched.

    :ivar all_errors_are_fatal: if true, an exception is raised not only
        for ``fatal`` diagnostic level, but also ``error``
    """
    def __init__(self, all_errors_are_fatal=False):
        self.all_errors_are_fatal = all_errors_are_fatal
        self._appended_notes = []

    def process(self, diagnostic):
        """
        The default implementation of :meth:`process` renders non-fatal
        diagnostics to ``sys.stderr``, and raises fatal ones as a :class:`Error`.
        """
        diagnostic.notes += self._appended_notes
        self.render_diagnostic(diagnostic)
        if diagnostic.level == "fatal" or \
                (self.all_errors_are_fatal and diagnostic.level == "error"):
            raise Error(diagnostic)

    @contextmanager
    def context(self, *notes):
        """
        A context manager that appends ``note`` to every diagnostic processed by
        this engine.
        """
        self._appended_notes += notes
        yield
        del self._appended_notes[-len(notes):]

    def render_diagnostic(self, diagnostic):
        sys.stderr.write("\n".join(diagnostic.render()) + "\n")
