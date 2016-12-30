def get_ident():
  f = __frame__()
  while f.f_back:
    f = f.f_back
  return id(f)
