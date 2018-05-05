import unittest
import os
import sys
from ctypes import *
from ctypes.util import find_library
from ctypes.test import is_resource_enabled

if sys.platform == "win32":
    lib_gl = find_library("OpenGL32")
    lib_glu = find_library("Glu32")
    lib_gle = None
elif sys.platform == "darwin":
    lib_gl = lib_glu = find_library("OpenGL")
    lib_gle = None
else:
    lib_gl = find_library("GL")
    lib_glu = find_library("GLU")
    lib_gle = find_library("gle")

## print, for debugging
if is_resource_enabled("printing"):
    if lib_gl or lib_glu or lib_gle:
        print "OpenGL libraries:"
        for item in (("GL", lib_gl),
                     ("GLU", lib_glu),
                     ("gle", lib_gle)):
            print "\t", item


# On some systems, loading the OpenGL libraries needs the RTLD_GLOBAL mode.
class Test_OpenGL_libs(unittest.TestCase):
    def setUp(self):
        self.gl = self.glu = self.gle = None
        if lib_gl:
            try:
                self.gl = CDLL(lib_gl, mode=RTLD_GLOBAL)
            except OSError:
                pass
        if lib_glu:
            try:
                self.glu = CDLL(lib_glu, RTLD_GLOBAL)
            except OSError:
                pass
        if lib_gle:
            try:
                self.gle = CDLL(lib_gle)
            except OSError:
                pass

    def tearDown(self):
        self.gl = self.glu = self.gle = None

    @unittest.skipUnless(lib_gl, 'lib_gl not available')
    def test_gl(self):
        if self.gl:
            self.gl.glClearIndex

    @unittest.skipUnless(lib_glu, 'lib_glu not available')
    def test_glu(self):
        if self.glu:
            self.glu.gluBeginCurve

    @unittest.skipUnless(lib_gle, 'lib_gle not available')
    def test_gle(self):
        if self.gle:
            self.gle.gleGetJoinStyle

# On platforms where the default shared library suffix is '.so',
# at least some libraries can be loaded as attributes of the cdll
# object, since ctypes now tries loading the lib again
# with '.so' appended of the first try fails.
#
# Won't work for libc, unfortunately.  OTOH, it isn't
# needed for libc since this is already mapped into the current
# process (?)
#
# On MAC OSX, it won't work either, because dlopen() needs a full path,
# and the default suffix is either none or '.dylib'.
@unittest.skip('test disabled')
@unittest.skipUnless(os.name=="posix" and sys.platform != "darwin",
                     'test not suitable for this platform')
class LoadLibs(unittest.TestCase):
    def test_libm(self):
        import math
        libm = cdll.libm
        sqrt = libm.sqrt
        sqrt.argtypes = (c_double,)
        sqrt.restype = c_double
        self.assertEqual(sqrt(2), math.sqrt(2))

if __name__ == "__main__":
    unittest.main()
