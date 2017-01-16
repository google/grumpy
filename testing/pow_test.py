


assert 2.0 ** -2 == 0.25, "2.0 ** -2"
assert 2.0 ** -1 == 0.5, "2.0 ** -1"
assert 2.0 ** 0 == 1, "2.0 ** 0"
assert 2.0 ** 1 == 2, "2.0 ** 1"
assert 2.0 ** 2 == 4, "2.0 ** 2"

assert (-2.0) ** -2 == 0.25, "(-2.0) ** -2"
assert (-2.0) ** -1 == -0.5, "(-2.0) ** -1"
assert (-2.0) ** 0 == 1, "(-2.0) ** 0"
assert (-2.0) ** 1 == -2, "(-2.0) ** 1"
assert (-2.0) ** 2 == 4, "(-2.0) ** 2"

assert 2 ** -2 == 0.25, "2 ** -2"
assert 2 ** -1 == 0.5, "2 ** -1"
assert 2 ** 0 == 1, "2 ** 0"
assert 2 ** 1 == 2, "2 ** 1"
assert 2 ** 2 == 4, "2 ** 2"

for zero in (0, 0L, 0.0):
    try:
        result = zero ** -2
        assert "0 ** -2"
    except ZeroDivisionError:
        pass

    try:
        result = zero ** -1
        assert "0 ** -1"
    except ZeroDivisionError:
        pass

    assert zero ** 0 == 1, '0 ** 0'
    assert zero ** 1 == 0, '0 ** 1'
    assert zero ** 2 == 0, '0 ** 2'


    assert 2 ** zero == 1
    assert (-2.0) ** zero == 1
    assert 3L ** zero == 1



assert (-2) ** -2 == 0.25, '(-2) ** -2'
assert (-2) ** -1 == -0.5, '(-2) ** -1'
assert (-2) ** 0 == 1, '(-2) ** 0'
assert (-2) ** 1 == -2, '(-2) ** 1'
assert (-2) ** 2 == 4, '(-2) ** 2'

assert 2 ** 128 == 340282366920938463463374607431768211456, "2 ** 128"

large_number = 2 ** 128

assert large_number ** -1 == (1.0  / large_number)
assert large_number ** 0 == 1
assert large_number ** 1 == large_number



