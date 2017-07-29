from '__go__/grumpy' import NewTryableMutex, StartThread, ThreadCount


class error(Exception):
  pass


def get_ident():
  f = __frame__()
  while f.f_back:
    f = f.f_back
  return id(f)


class LockType(object):
  def __init__(self):
    self._mutex = NewTryableMutex()

  def acquire(self, waitflag=1):
    if waitflag:
      self._mutex.Lock()
      return True
    return self._mutex.TryLock()

  def release(self):
    self._mutex.Unlock()

  def __enter__(self):
    self.acquire()

  def __exit__(self, *args):
    self.release()


def allocate_lock():
  """Dummy implementation of thread.allocate_lock()."""
  return LockType()


def start_new_thread(func, args, kwargs=None):
  if kwargs is None:
    kwargs = {}
  l = allocate_lock()
  ident = []
  def thread_func():
    ident.append(get_ident())
    l.release()
    func(*args, **kwargs)
  l.acquire()
  StartThread(thread_func)
  l.acquire()
  return ident[0]


def stack_size(n=0):
  if n:
    raise error('grumpy does not support setting stack size')
  return 0


def _count():
  return ThreadCount
