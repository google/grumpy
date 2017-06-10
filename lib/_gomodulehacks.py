# coding: utf-8

from __go__.grumpy import SysModules
import errno

# For some reason, importing and extending grumpy.Module does not work.
Module = type(errno)


def hybrid_module(modulename, modulefile, moduledict, all_attrs, globals_):
    """
    Augment native 'moduledict' with Python-sourced parts

    Allows 'modulename' to use 'moduledict' from outside,
    for example a Grumpy dict from native module.

    And does include the resulting module on sys.modules at the end.

    Should be called as:
        hybrid_module(__name__, __file__, YourmoduleDict, __all__, globals())
    On the last line of the Python-part of the module
    """
    class HybridModule(Module):
        def __init__(self):
            moduledict['__name__'] = modulename
            moduledict['__file__'] = modulefile
            for k in all_attrs:
                moduledict[k] = globals_[k]
        def __setattr__(self, name, value):
            moduledict[name] = value
        def __getattribute__(self, name):   # TODO: replace w/ __getattr__ when implemented
            resp = moduledict.get(name)
            if resp is None and name not in moduledict:
                return super(HybridModule, self).__getattribute__(name)
            return resp

    finalmodule = HybridModule()
    SysModules[modulename] = finalmodule
    return finalmodule
