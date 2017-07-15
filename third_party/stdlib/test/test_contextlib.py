"""Unit tests for contextlib.py, and other context managers."""

import sys
import tempfile
import unittest
# from contextlib import *  # Tests __all__
import contextlib
contextmanager = contextlib.contextmanager
nested = contextlib.nested
closing = contextlib.closing
from test import test_support
try:
    import threading
except ImportError:
    threading = None


class ContextManagerTestCase(unittest.TestCase):

    def test_contextmanager_plain(self):
        state = []
        @contextmanager
        def woohoo():
            state.append(1)
            yield 42
            state.append(999)
        with woohoo() as x:
            self.assertEqual(state, [1])
            self.assertEqual(x, 42)
            state.append(x)
        self.assertEqual(state, [1, 42, 999])

    @unittest.skip('grumpy')
    def test_contextmanager_finally(self):
        state = []
        @contextmanager
        def woohoo():
            state.append(1)
            try:
                yield 42
            finally:
                state.append(999)
        with self.assertRaises(ZeroDivisionError):
            with woohoo() as x:
                self.assertEqual(state, [1])
                self.assertEqual(x, 42)
                state.append(x)
                raise ZeroDivisionError()
        self.assertEqual(state, [1, 42, 999])

    @unittest.skip('grumpy')
    def test_contextmanager_no_reraise(self):
        @contextmanager
        def whee():
            yield
        ctx = whee()
        ctx.__enter__()
        # Calling __exit__ should not result in an exception
        self.assertFalse(ctx.__exit__(TypeError, TypeError("foo"), None))

    @unittest.skip('grumpy')
    def test_contextmanager_trap_yield_after_throw(self):
        @contextmanager
        def whoo():
            try:
                yield
            except:
                yield
        ctx = whoo()
        ctx.__enter__()
        self.assertRaises(
            RuntimeError, ctx.__exit__, TypeError, TypeError("foo"), None
        )

    @unittest.skip('grumpy')
    def test_contextmanager_except(self):
        state = []
        @contextmanager
        def woohoo():
            state.append(1)
            try:
                yield 42
            except ZeroDivisionError, e:
                state.append(e.args[0])
                self.assertEqual(state, [1, 42, 999])
        with woohoo() as x:
            self.assertEqual(state, [1])
            self.assertEqual(x, 42)
            state.append(x)
            raise ZeroDivisionError(999)
        self.assertEqual(state, [1, 42, 999])

    def _create_contextmanager_attribs(self):
        def attribs(**kw):
            def decorate(func):
                for k,v in kw.items():
                    setattr(func,k,v)
                return func
            return decorate
        @contextmanager
        @attribs(foo='bar')
        def baz(spam):
            """Whee!"""
        return baz

    @unittest.skip('grumpy')
    def test_contextmanager_attribs(self):
        baz = self._create_contextmanager_attribs()
        self.assertEqual(baz.__name__,'baz')
        self.assertEqual(baz.foo, 'bar')

    @unittest.skip('grumpy')
    @unittest.skipIf(sys.flags.optimize >= 2,
                     "Docstrings are omitted with -O2 and above")
    def test_contextmanager_doc_attrib(self):
        baz = self._create_contextmanager_attribs()
        self.assertEqual(baz.__doc__, "Whee!")

    def test_keywords(self):
        # Ensure no keyword arguments are inhibited
        @contextmanager
        def woohoo(self, func, args, kwds):
            yield (self, func, args, kwds)
        with woohoo(self=11, func=22, args=33, kwds=44) as target:
            self.assertEqual(target, (11, 22, 33, 44))

class NestedTestCase(unittest.TestCase):

    # XXX This needs more work

    def test_nested(self):
        @contextmanager
        def a():
            yield 1
        @contextmanager
        def b():
            yield 2
        @contextmanager
        def c():
            yield 3
        with nested(a(), b(), c()) as (x, y, z):
            self.assertEqual(x, 1)
            self.assertEqual(y, 2)
            self.assertEqual(z, 3)

    @unittest.skip('grumpy')
    def test_nested_cleanup(self):
        state = []
        @contextmanager
        def a():
            state.append(1)
            try:
                yield 2
            finally:
                state.append(3)
        @contextmanager
        def b():
            state.append(4)
            try:
                yield 5
            finally:
                state.append(6)
        with self.assertRaises(ZeroDivisionError):
            with nested(a(), b()) as (x, y):
                state.append(x)
                state.append(y)
                1 // 0
        self.assertEqual(state, [1, 4, 2, 5, 6, 3])

    def test_nested_right_exception(self):
        @contextmanager
        def a():
            yield 1
        class b(object):
            def __enter__(self):
                return 2
            def __exit__(self, *exc_info):
                try:
                    raise Exception()
                except:
                    pass
        with self.assertRaises(ZeroDivisionError):
            with nested(a(), b()) as (x, y):
                1 // 0
        self.assertEqual((x, y), (1, 2))

    @unittest.skip('grumpy')
    def test_nested_b_swallows(self):
        @contextmanager
        def a():
            yield
        @contextmanager
        def b():
            try:
                yield
            except:
                # Swallow the exception
                pass
        try:
            with nested(a(), b()):
                1 // 0
        except ZeroDivisionError:
            self.fail("Didn't swallow ZeroDivisionError")

    def test_nested_break(self):
        @contextmanager
        def a():
            yield
        state = 0
        while True:
            state += 1
            with nested(a(), a()):
                break
            state += 10
        self.assertEqual(state, 1)

    def test_nested_continue(self):
        @contextmanager
        def a():
            yield
        state = 0
        while state < 3:
            state += 1
            with nested(a(), a()):
                continue
            state += 10
        self.assertEqual(state, 3)

    def test_nested_return(self):
        @contextmanager
        def a():
            try:
                yield
            except:
                pass
        def foo():
            with nested(a(), a()):
                return 1
            return 10
        self.assertEqual(foo(), 1)

class ClosingTestCase(unittest.TestCase):

    # XXX This needs more work

    def test_closing(self):
        state = []
        class C(object):
            def close(self):
                state.append(1)
        x = C()
        self.assertEqual(state, [])
        with closing(x) as y:
            self.assertEqual(x, y)
        self.assertEqual(state, [1])

    def test_closing_error(self):
        state = []
        class C(object):
            def close(self):
                state.append(1)
        x = C()
        self.assertEqual(state, [])
        with self.assertRaises(ZeroDivisionError):
            with closing(x) as y:
                self.assertEqual(x, y)
                1 // 0
        self.assertEqual(state, [1])

class FileContextTestCase(unittest.TestCase):

    @unittest.skip('grumpy')
    def testWithOpen(self):
        tfn, _ = tempfile.mkstemp()
        try:
            f = None
            with open(tfn, "w") as f:
                self.assertFalse(f.closed)
                f.write("Booh\n")
            self.assertTrue(f.closed)
            f = None
            with self.assertRaises(ZeroDivisionError):
                with open(tfn, "r") as f:
                    self.assertFalse(f.closed)
                    self.assertEqual(f.read(), "Booh\n")
                    1 // 0
            self.assertTrue(f.closed)
        finally:
            test_support.unlink(tfn)

@unittest.skipUnless(threading, 'Threading required for this test.')
class LockContextTestCase(unittest.TestCase):

    def boilerPlate(self, lock, locked):
        self.assertFalse(locked())
        with lock:
            self.assertTrue(locked())
        self.assertFalse(locked())
        with self.assertRaises(ZeroDivisionError):
            with lock:
                self.assertTrue(locked())
                1 // 0
        self.assertFalse(locked())

    @unittest.skip('grumpy')
    def testWithLock(self):
        lock = threading.Lock()
        self.boilerPlate(lock, lock.locked)

    def testWithRLock(self):
        lock = threading.RLock()
        self.boilerPlate(lock, lock._is_owned)

    def testWithCondition(self):
        lock = threading.Condition()
        def locked():
            return lock._is_owned()
        self.boilerPlate(lock, locked)

    def testWithSemaphore(self):
        lock = threading.Semaphore()
        def locked():
            if lock.acquire(False):
                lock.release()
                return False
            else:
                return True
        self.boilerPlate(lock, locked)

    def testWithBoundedSemaphore(self):
        lock = threading.BoundedSemaphore()
        def locked():
            if lock.acquire(False):
                lock.release()
                return False
            else:
                return True
        self.boilerPlate(lock, locked)

# This is needed to make the test actually run under regrtest.py!
def test_main():
#    with test_support.check_warnings(("With-statements now directly support "
#                                      "multiple context managers",
#                                      DeprecationWarning)):
        test_support.run_unittest(__name__)

if __name__ == "__main__":
    test_main()
