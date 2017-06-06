f = open('/tmp/file_test__someunlikelyexistingfile', 'w')
assert f.softspace == 0

f.softspace = 1
assert f.softspace == 1

try:
    f.softspace = '4321'     # should not be converted automatically
except TypeError as e:
    if not str(e).endswith('is required'):
        raise e     # Wrong exception arrived to us!
else:
    raise RuntimeError('a TypeError should had raised.')

assert f.softspace == 1
