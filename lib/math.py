#NOTE: This is ugly.
from __go__.math import (Pi, E, Ceil, Copysign, Abs, Floor, Mod, Frexp, IsInf,
    IsNaN, Exp2, Modf, Trunc, Exp, Expm1, Log, Log1p, Log10, Pow, Sqrt, Acos,
    Asin, Atan, Atan2, Hypot, Sin, Cos, Tan, Acosh, Asinh, Atanh, Sinh, Cosh,
    Tanh, Erf, Erfc, Gamma, Lgamma) # pylint: disable=g-multiple-import

# Constants

pi = Pi
e = E

# Number-theoretic and representation functions

def ceil(x):
    return Ceil(x)
    
def copysign(x,y):
    return Copysign(x,y)
    
def fabs(x):
    return Abs(x)

def factorial(x):
    
    def factorial_helper(x, acc):
        if x <= 1:
            return acc
        else:
            return factorial_helper(x - 1, acc * x)
    
    if x % 1 != 0 or x < 1:
        raise ValueError
    else:
        return factorial_helper(x,1)
        
def floor(x):
    return Floor(x)

def fmod(x):
    return Mod(x)

def frexp(x):
    return Frexp(x)

# NOTE: This function exists in python, but I don't know how to write it,
#and I don't see it anywhere in Go's math library.
#
# def fsum(x):
#     pass

def isinf(x):
    return IsInf(x,0)
    
def isnan(x):
    return IsNaN(x)

def ldexp(x,i):
    return x * Exp2(i)
    
def modf(x):
    #Modf returns (int, frac), but python should return (frac, int)
    (a, b) = Modf(x)
    return (b, a)
    
def trunc(x):
    return Trunc(x)

# Power and logarithmic functions

def exp(x):
    return Exp(x)

def expm1(x):
    return Expm1(x)

def log(x, b=None):
    if b is None:
        return Log(x)
    else:
        # NOTE: We can try and catch more special cases to delegate to specific
        # Go functions or maybe there is a function that does this and I missed it
        return Log(x)/Log(b)

def log1p(x):
    return Log1p(x)
    
def log10(x):
    return Log10(x)
    
def pow(x, y):
    return Pow(x,y)

def sqrt(x):
    return Sqrt(x)

# Trigonometric functions

def acos(x):
    return Acos(x)

def asin(x):
    return Asin(x)

def atan(x):
    return Atan(x)

def atan2(y, x):
    return Atan2(y, x)
    
def cos(x):
    return Cos(x)

def hypot(x, y):
    return Hypot(x,y)

def sin(x):
    return Sin(x)

def tan(x):
    return Tan(x)

# Angular conversion

def degrees(x):
    return (x * 180) / pi

def radians(x):
    return (x * pi) / 180

# Hyperbolic functions

def acosh(x):
    return Acosh(x)
    
def asinh(x):
    return Asinh(x)
    
def atanh(x):
    return Atanh(x)

def cosh(x):
    return Cosh(x)

def sinh(x):
    return Sinh(x)
    
def tanh(x):
    return Tanh(x)
    
# Special functions

def erf(x):
    return Erf(x)
    
def erfc(x):
    return Erfc(x)

def gamma(x):
    return Gamma(x)
    
def lgamma(x):
    return Lgamma(x)
