from __go__.sync import type_Mutex as Mutex


def get_ident():
  f = __frame__()
  while f.f_back:
    f = f.f_back
  return id(f)


class LockType(object):
  def __init__(self):
    self._mutex = Mutex.new()

  def acquire(self):
    self._mutex.Lock()

  def release(self):
    self._mutex.Unlock()

  def __enter__(self):
    self.acquire()

  def __exit__(self, *args):
    self.release()


def allocate_lock():
    """Dummy implementation of thread.allocate_lock()."""
    return LockType()
