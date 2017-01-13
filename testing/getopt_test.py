import getopt

args = '-a -b -cfoo -d bar a1 a2'.split()
optlist, args = getopt.getopt(args, 'abc:d:')
assert optlist == [('-a', ''), ('-b', ''), ('-c', 'foo'), ('-d', 'bar')]

# TODO: str.index has to be implemented
# s = '--condition=foo --testing --output-file abc.def -x a1 a2'
# args = s.split()
# optlist, args = getopt.getopt(
#     args, 'x', ['condition=', 'output-file=', 'testing'])

# assert optlist == [('--condition', 'foo'), ('--testing', ''),
#                    ('--output-file', 'abc.def'), ('-x', '')]
