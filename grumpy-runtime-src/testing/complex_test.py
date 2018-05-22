# Copyright 2016 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

assert repr(1j) == "1j"
assert repr(complex()) == "0j"
assert repr(complex('nan-nanj')) == '(nan+nanj)'
assert repr(complex('-Nan+NaNj')) == '(nan+nanj)'
assert repr(complex('inf-infj')) == '(inf-infj)'
assert repr(complex('+inf+infj')) == '(inf+infj)'
assert repr(complex('-infINIty+infinityj')) == '(-inf+infj)'

assert complex(1.8456e3) == (1845.6+0j)
assert complex('1.8456e3') == (1845.6+0j)
assert complex(0, -365.12) == -365.12j
assert complex('-365.12j') == -365.12j
assert complex(-1.23E2, -45.678e1) == (-123-456.78j)
assert complex('-1.23e2-45.678e1j') == (-123-456.78j)
assert complex(21.98, -1) == (21.98-1j)
assert complex('21.98-j') == (21.98-1j)
assert complex('-j') == -1j
assert complex('+j') == 1j
assert complex('j') == 1j
assert complex(' \t \n \r ( \t \n \r 2.1-3.4j \t \n \r ) \t \n \r ') == (2.1-3.4j)
assert complex(complex(complex(3.14))) == (3.14+0j)
assert complex(complex(1, -2), .151692) == (1-1.848308j)
assert complex(complex(3.14), complex(-0.151692)) == (3.14-0.151692j)
assert complex(complex(-1, 2), complex(3, -4)) == (3+5j)

try:
  complex('((2.1-3.4j))')
except ValueError as e:
  assert str(e) == "complex() arg is a malformed string"
else:
  raise AssertionError('this was supposed to raise an exception')

try:
  complex('3.14 - 15.16 j')
except ValueError as e:
  assert str(e) == "complex() arg is a malformed string"
else:
  raise AssertionError('this was supposed to raise an exception')

try:
  complex('foo')
except ValueError as e:
  assert str(e) == "complex() arg is a malformed string"
else:
  raise AssertionError('this was supposed to raise an exception')

try:
  complex('foo', 1)
except TypeError as e:
  assert str(e) == "complex() can't take second arg if first is a string"
else:
  raise AssertionError('this was supposed to raise an exception')

try:
  complex(1, 'bar')
except TypeError as e:
  assert str(e) == "complex() second arg can't be a string"
else:
  raise AssertionError('this was supposed to raise an exception')

# __nonzero__

assert complex(0, 0).__nonzero__() == False
assert complex(.0, .0).__nonzero__() == False
assert complex(0.0, 0.1).__nonzero__() == True
assert complex(1, 0).__nonzero__() == True
assert complex(3.14, -0.001e+5).__nonzero__() == True
assert complex(float('nan'), float('nan')).__nonzero__() == True
assert complex(-float('inf'), float('inf')).__nonzero__() == True

# __pos__

assert complex(0, 0).__pos__() == 0j
assert complex(42, -0.1).__pos__() == (42-0.1j)
assert complex(-1.2, 375E+2).__pos__() == (-1.2+37500j)
assert repr(complex(5, float('nan')).__pos__()) == '(5+nanj)'
assert repr(complex(float('inf'), 0.618).__pos__()) == '(inf+0.618j)'