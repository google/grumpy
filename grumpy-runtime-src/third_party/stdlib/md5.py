# $Id$
#
#  Copyright (C) 2005   Gregory P. Smith (greg@krypto.org)
#  Licensed to PSF under a Contributor Agreement.

# import warnings
# warnings.warn("the md5 module is deprecated; use hashlib instead",
#                 DeprecationWarning, 2)

# from hashlib import md5
import _md5

new = _md5.new
md5 = _md5.new
digest_size = 16
