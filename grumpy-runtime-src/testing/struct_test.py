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

import _struct as struct

# struct test
A = 0x67452301
B = 0xefcdab89
C = 0x98badcfe
D = 0x10325476

expected = '\x01#Eg\x89\xab\xcd\xef\xfe\xdc\xba\x98vT2\x10'

assert struct.pack("<IIII", A, B, C, D) == expected
