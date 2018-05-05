from ctypes import *
import sys, unittest
import os
from ctypes.util import find_library
from ctypes.test import is_resource_enabled

libc_name = None
if os.name == "nt":
    libc_name = find_library("c")
elif os.name == "ce":
    libc_name = "coredll"
elif sys.platform == "cygwin":
    libc_name = "cygwin1.dll"
else:
    libc_name = find_library("c")

if is_resource_enabled("printing"):
    print "libc_name is", libc_name

class LoaderTest(unittest.TestCase):

    unknowndll = "xxrandomnamexx"

    @unittest.skipUnless(libc_name is not None, 'could not find libc')
    def test_load(self):
        CDLL(libc_name)
        CDLL(os.path.basename(libc_name))
        self.assertRaises(OSError, CDLL, self.unknowndll)

    @unittest.skipUnless(libc_name is not None, 'could not find libc')
    @unittest.skipUnless(libc_name is not None and
                         os.path.basename(libc_name) == "libc.so.6",
                         'wrong libc path for test')
    def test_load_version(self):
        cdll.LoadLibrary("libc.so.6")
        # linux uses version, libc 9 should not exist
        self.assertRaises(OSError, cdll.LoadLibrary, "libc.so.9")
        self.assertRaises(OSError, cdll.LoadLibrary, self.unknowndll)

    def test_find(self):
        for name in ("c", "m"):
            lib = find_library(name)
            if lib:
                cdll.LoadLibrary(lib)
                CDLL(lib)

    @unittest.skipUnless(os.name in ("nt", "ce"),
                         'test specific to Windows (NT/CE)')
    def test_load_library(self):
        self.assertIsNotNone(libc_name)
        if is_resource_enabled("printing"):
            print find_library("kernel32")
            print find_library("user32")

        if os.name == "nt":
            windll.kernel32.GetModuleHandleW
            windll["kernel32"].GetModuleHandleW
            windll.LoadLibrary("kernel32").GetModuleHandleW
            WinDLL("kernel32").GetModuleHandleW
        elif os.name == "ce":
            windll.coredll.GetModuleHandleW
            windll["coredll"].GetModuleHandleW
            windll.LoadLibrary("coredll").GetModuleHandleW
            WinDLL("coredll").GetModuleHandleW

    @unittest.skipUnless(os.name in ("nt", "ce"),
                         'test specific to Windows (NT/CE)')
    def test_load_ordinal_functions(self):
        import _ctypes_test
        dll = WinDLL(_ctypes_test.__file__)
        # We load the same function both via ordinal and name
        func_ord = dll[2]
        func_name = dll.GetString
        # addressof gets the address where the function pointer is stored
        a_ord = addressof(func_ord)
        a_name = addressof(func_name)
        f_ord_addr = c_void_p.from_address(a_ord).value
        f_name_addr = c_void_p.from_address(a_name).value
        self.assertEqual(hex(f_ord_addr), hex(f_name_addr))

        self.assertRaises(AttributeError, dll.__getitem__, 1234)

    @unittest.skipUnless(os.name == "nt", 'Windows-specific test')
    def test_1703286_A(self):
        from _ctypes import LoadLibrary, FreeLibrary
        # On winXP 64-bit, advapi32 loads at an address that does
        # NOT fit into a 32-bit integer.  FreeLibrary must be able
        # to accept this address.

        # These are tests for http://www.python.org/sf/1703286
        handle = LoadLibrary("advapi32")
        FreeLibrary(handle)

    @unittest.skipUnless(os.name == "nt", 'Windows-specific test')
    def test_1703286_B(self):
        # Since on winXP 64-bit advapi32 loads like described
        # above, the (arbitrarily selected) CloseEventLog function
        # also has a high address.  'call_function' should accept
        # addresses so large.
        from _ctypes import call_function
        advapi32 = windll.advapi32
        # Calling CloseEventLog with a NULL argument should fail,
        # but the call should not segfault or so.
        self.assertEqual(0, advapi32.CloseEventLog(None))
        windll.kernel32.GetProcAddress.argtypes = c_void_p, c_char_p
        windll.kernel32.GetProcAddress.restype = c_void_p
        proc = windll.kernel32.GetProcAddress(advapi32._handle,
                                              "CloseEventLog")
        self.assertTrue(proc)
        # This is the real test: call the function via 'call_function'
        self.assertEqual(0, call_function(proc, (None,)))

if __name__ == "__main__":
    unittest.main()
