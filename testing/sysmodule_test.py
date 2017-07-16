from StringIO import StringIO
import sys

print 'To sys.stdout'

old_stdout = sys.stdout
sio = StringIO()
sys.stdout = sio

print 'To replaced sys.stdout'

sys.stdout = old_stdout
print 'To original sys.stdout'

assert sio.tell() == len('To replaced sys.stdout')+1, 'Should had printed to StringIO, not STDOUT'