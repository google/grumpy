""" Supplies the internal functions for functools.py in the standard library """

__all__ = ['reduce', 'partial']

# reduce() has moved to _functools in Python 2.6+.
def reduce(function, iterable, initializer=None):
    it = iter(iterable)
    if initializer is None:
        try:
            initializer = next(it)
        except StopIteration:
            raise TypeError('reduce() of empty sequence with no initial value')
    accum_value = initializer
    for x in it:
        accum_value = function(accum_value, x)
    return accum_value

class partial(object):
    """
    partial(func, *args, **keywords) - new function with partial application
    of the given arguments and keywords.
    """

    __slots__ = ('_func', '_args', '_keywords', '__dict__')

    def __init__(*args, **keywords):
        if len(args) < 2:
            raise TypeError('__init__() takes at least 2 arguments (%d given)'
                            % len(args))
        self, func, args = args[0], args[1], args[2:]
        if not callable(func):
            raise TypeError("the first argument must be callable")
        self._func = func
        self._args = args
        self._keywords = keywords

    def __delattr__(self, key):
        if key == '__dict__':
            raise TypeError("a partial object's dictionary may not be deleted")
        object.__delattr__(self, key)

    # @property
    def func(self):
        return self._func
    # TODO: Make this a decorator once they're implemented.
    func = property(func)

    # @property
    def args(self):
        return self._args
    # TODO: Make this a decorator once they're implemented.
    args = property(args)

    # @property
    def keywords(self):
        return self._keywords
    # TODO: Make this a decorator once they're implemented.
    keywords = property(keywords)

    def __call__(self, *fargs, **fkeywords):
        if self._keywords:
            fkeywords = dict(self._keywords, **fkeywords)
        return self._func(*(self._args + fargs), **fkeywords)

    def __reduce__(self):
        d = dict((k, v) for k, v in self.__dict__.iteritems() if k not in
                ('_func', '_args', '_keywords'))
        if len(d) == 0:
            d = None
        return (type(self), (self._func,),
                (self._func, self._args, self._keywords, d))

    def __setstate__(self, state):
        if not isinstance(state, tuple) or len(state) != 4:
            raise TypeError("invalid partial state")

        func, args, keywords, d = state

        if (not callable(func) or not isinstance(args, tuple) or
            (keywords is not None and not isinstance(keywords, dict))):
            raise TypeError("invalid partial state")

        self._func = func
        self._args = tuple(args)

        if keywords is None:
            keywords = {}
        elif type(keywords) is not dict:
            keywords = dict(keywords)
        self._keywords = keywords

        if d is None:
            self.__dict__.clear()
        else:
            self.__dict__.update(d)
