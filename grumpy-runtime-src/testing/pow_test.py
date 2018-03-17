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

assert 2.0 ** -2 == 0.25, "2.0 ** -2"
assert 2.0 ** -1 == 0.5, "2.0 ** -1"
assert 2.0 ** 0 == 1, "2.0 ** 0"
assert 2.0 ** 1 == 2, "2.0 ** 1"
assert 2.0 ** 2 == 4, "2.0 ** 2"

assert (-2.0) ** -2 == 0.25, "(-2.0) ** -2"
assert (-2.0) ** -1 == -0.5, "(-2.0) ** -1"
assert (-2.0) ** 0 == 1, "(-2.0) ** 0"
assert (-2.0) ** 1 == -2, "(-2.0) ** 1"
assert (-2.0) ** 2 == 4, "(-2.0) ** 2"

assert 2 ** -2 == 0.25, "2 ** -2"
assert 2 ** -1 == 0.5, "2 ** -1"
assert 2 ** 0 == 1, "2 ** 0"
assert 2 ** 1 == 2, "2 ** 1"
assert 2 ** 2 == 4, "2 ** 2"

assert 2L ** -2 == 0.25, "2L ** -2"
assert 2L ** -1 == 0.5, "2L ** -1"
assert 2L ** 0 == 1, "2L ** 0"
assert 2L ** 1 == 2, "2L ** 1"
assert 2L ** 2 == 4, "2L ** 2"

# Test the rpow operator on long
assert 2 ** -2L == 0.25, "2 ** -2L"
assert 2 ** -1L == 0.5, "2 ** -1L"
assert 2 ** 0L == 1, "2 ** 0L"
assert 2 ** 1L == 2, "2 ** 1L"
assert 2 ** 2L == 4, "2 ** 2L"

for zero in (0, 0L, 0.0):
    try:
        result = zero ** -2
        assert "0 ** -2"
    except ZeroDivisionError:
        pass

    try:
        result = zero ** -1
        assert "0 ** -1"
    except ZeroDivisionError:
        pass

    assert zero ** 0 == 1, '0 ** 0'
    assert zero ** 1 == 0, '0 ** 1'
    assert zero ** 2 == 0, '0 ** 2'

    assert 2 ** zero == 1
    assert (-2.0) ** zero == 1
    assert 3L ** zero == 1

assert (-2) ** -2 == 0.25, '(-2) ** -2'
assert (-2) ** -1 == -0.5, '(-2) ** -1'
assert (-2) ** 0 == 1, '(-2) ** 0'
assert (-2) ** 1 == -2, '(-2) ** 1'
assert (-2) ** 2 == 4, '(-2) ** 2'

assert 2 ** 128 == 340282366920938463463374607431768211456, "2 ** 128"

# chose something which can be represented exact as an IEEE floating point number
large_number = (2 ** 128  + 2 ** 127)

assert large_number ** -1 == (1.0  / large_number), "large_number ** -1 == (1.0  / large_number)"
assert large_number ** 0 == 1, "large_number ** 0 == 1"
assert large_number ** 1 == large_number, "large_number ** 1 == large_number"

