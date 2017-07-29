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

from '__go__/math' import (Pi, E, Ceil, Copysign, Abs, Floor, Mod, Frexp, IsInf,
    IsNaN, Exp2, Modf, Trunc, Exp, Expm1, Log, Log1p, Log10, Pow, Sqrt, Acos,
    Asin, Atan, Atan2, Hypot, Sin, Cos, Tan, Acosh, Asinh, Atanh, Sinh, Cosh,
    Tanh, Erf, Erfc, Gamma, Lgamma)  # pylint: disable=g-multiple-import

# Constants

pi = Pi


e = E


# Number-theoretic and representation functions

def ceil(x):
    return Ceil(float(x))


def copysign(x, y):
    return Copysign(float(x), float(y))


def fabs(x):
    return Abs(float(x))


def factorial(x):
    try:
        xi = int(x)
    except TypeError:
        xi = None

    try:
        xf = float(x)
    except TypeError:
        xf = None

    if xi is None:
        xi = int(xf)
        if xi != xf:
            raise ValueError("factorial() only accepts integral values")
    elif xf is None and xi is None:
        raise TypeError("an integer is required")
    elif xf is None:
        pass
    elif xf != xi:
        raise ValueError("factorial() only accepts integral values")

    x = xi

    if x < 0:
        raise ValueError("factorial() not defined for negative values")

    acc = 1

    for value in range(2, x+1):
        acc *= value

    return acc


def floor(x):
    return Floor(float(x))


def fmod(x):
    return Mod(float(x))


def frexp(x):
    return Frexp(float(x))


# TODO: Implement fsum()
# def fsum(x):
#    pass


def isinf(x):
    return IsInf(float(x), 0)


def isnan(x):
    return IsNaN(float(x))


def ldexp(x, i):
    return float(x) * Exp2(float(i))


def modf(x):
    # Modf returns (int, frac), but python should return (frac, int).
    a, b = Modf(float(x))
    return b, a


def trunc(x):
    return Trunc(float(x))


# Power and logarithmic functions

def exp(x):
    return Exp(float(x))


def expm1(x):
    return Expm1(float(x))


def log(x, b=None):
    if b is None:
        return Log(float(x))

    # NOTE: We can try and catch more special cases to delegate to specific
    # Go functions or maybe there is a function that does this and I missed it.
    return Log(float(x)) / Log(float(b))


def log1p(x):
    return Log1p(float(x))


def log10(x):
    return Log10(float(x))


def pow(x, y):
    return Pow(float(x), float(y))


def sqrt(x):
    return Sqrt(float(x))


# Trigonometric functions

def acos(x):
    return Acos(float(x))


def asin(x):
    return Asin(float(x))


def atan(x):
    return Atan(float(x))


def atan2(y, x):
    return Atan2(float(y), float(x))


def cos(x):
    return Cos(float(x))


def hypot(x, y):
    return Hypot(float(x), float(y))


def sin(x):
    return Sin(float(x))


def tan(x):
    return Tan(float(x))


# Angular conversion

def degrees(x):
    return (float(x) * 180) / pi


def radians(x):
    return (float(x) * pi) / 180


# Hyperbolic functions

def acosh(x):
    return Acosh(float(x))


def asinh(x):
    return Asinh(float(x))


def atanh(x):
    return Atanh(float(x))


def cosh(x):
    return Cosh(float(x))


def sinh(x):
    return Sinh(float(x))


def tanh(x):
    return Tanh(float(x))


# Special functions

def erf(x):
    return Erf(float(x))


def erfc(x):
    return Erfc(float(x))


def gamma(x):
    return Gamma(float(x))


def lgamma(x):
    return Lgamma(float(x))
