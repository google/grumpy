# Tests are borrowed from pypy

import binascii

# obscure case, for compability with CPython
# assert binascii.a2b_uu("") == "\x00" * 0x20

for s, expected in [
    ("!,_", "3"),
    (" ", ""),
    ("!", "\x00"),
    ("!6", "X"),
    ('"6', "X\x00"),
    ('"W', "\xdc\x00"),
    ('"WA', "\xde\x10"),
    ('"WAX', "\xde\x1e"),
    ('#WAX', "\xde\x1e\x00"),
    ('#WAXR', "\xde\x1e2"),
    ('$WAXR', "\xde\x1e2\x00"),
    ('$WAXR6', "\xde\x1e2X"),
    ('%WAXR6U', "\xde\x1e2[P"),
    ('&WAXR6UB', "\xde\x1e2[X\x80"),
    ("'WAXR6UBA3", "\xde\x1e2[X\xa1L"),
    ('(WAXR6UBA3#', "\xde\x1e2[X\xa1L0"),
    (')WAXR6UBA3#Q', "\xde\x1e2[X\xa1L<@"),
    ('*WAXR6UBA3#Q!5', "\xde\x1e2[X\xa1L<AT"),
    ('!,_', '\x33'),
    ]:
    assert binascii.a2b_uu(s) == expected
    assert binascii.a2b_uu(s + ' ') == expected
    assert binascii.a2b_uu(s + '  ') == expected
    assert binascii.a2b_uu(s + '   ') == expected
    assert binascii.a2b_uu(s + '    ') == expected
    assert binascii.a2b_uu(s + '\n') == expected
    assert binascii.a2b_uu(s + '\r\n') == expected
    assert binascii.a2b_uu(s + '  \r\n') == expected
    assert binascii.a2b_uu(s + '    \r\n') == expected


for s, expected in [
    ("", " "),
    ("\x00", "!    "),
    ("X", "!6   "),
    ("X\x00", '"6   '),
    ("\xdc\x00", '"W   '),
    ("\xde\x10", '"WA  '),
    ("\xde\x1e", '"WAX '),
    ("\xde\x1e\x00", '#WAX '),
    ("\xde\x1e2", '#WAXR'),
    ("\xde\x1e2\x00", '$WAXR    '),
    ("\xde\x1e2X", '$WAXR6   '),
    ("\xde\x1e2[P", '%WAXR6U  '),
    ("\xde\x1e2[X\x80", '&WAXR6UB '),
    ("\xde\x1e2[X\xa1L", "'WAXR6UBA3   "),
    ("\xde\x1e2[X\xa1L0", '(WAXR6UBA3#  '),
    ("\xde\x1e2[X\xa1L<@", ')WAXR6UBA3#Q '),
    ("\xde\x1e2[X\xa1L<AT", '*WAXR6UBA3#Q!5   '),
    ]:
    assert binascii.b2a_uu(s) == expected + '\n'

for s, expected in [
    ("", ""),
    ("\n", ""),
    ("Yg==\n", "b"),
    ("Y g = \n = \r", "b"),     # random spaces
    ("Y\x80g\xff=\xc4=", "b"),  # random junk chars, >= 0x80
    ("abcd", "i\xb7\x1d"),
    ("abcdef==", "i\xb7\x1dy"),
    ("abcdefg=", "i\xb7\x1dy\xf8"),
    ("abcdefgh", "i\xb7\x1dy\xf8!"),
    ("abcdef==FINISHED", "i\xb7\x1dy"),
    ("abcdef=   \n   =FINISHED", "i\xb7\x1dy"),
    ("abcdefg=FINISHED", "i\xb7\x1dy\xf8"),
    ("abcd=efgh", "i\xb7\x1dy\xf8!"),
    ("abcde=fgh", "i\xb7\x1dy\xf8!"),
    ("abcdef=gh", "i\xb7\x1dy\xf8!"),
    ]:
    assert binascii.a2b_base64(s) == expected

for s, expected in [
    ("", ""),
    ("b", "Yg=="),
    ("i\xb7\x1d", "abcd"),
    ("i\xb7\x1dy", "abcdeQ=="),
    ("i\xb7\x1dy\xf8", "abcdefg="),
    ("i\xb7\x1dy\xf8!", "abcdefgh"),
    ("i\xb7\x1d" * 345, "abcd" * 345),
    ]:
    assert binascii.b2a_base64(s) == expected + '\n'

# TODO: %02x str interpolation has to be implemented
# for s, expected in [
#     # these are the tests from CPython 2.7
#     ("= ", "= "),
#     ("==", "="),
#     ("=AX", "=AX"),
#     ("=00\r\n=00", "\x00\r\n\x00"),
#     # more tests follow
#     ("=", ""),
#     ("abc=", "abc"),
#     ("ab=\ncd", "abcd"),
#     ("ab=\r\ncd", "abcd"),
#     (''.join(["=%02x" % n for n in range(256)]),
#                   ''.join(map(chr, range(256)))),
#     (''.join(["=%02X" % n for n in range(256)]),
#                   ''.join(map(chr, range(256)))),
#     ]:
#     assert binascii.a2b_qp(s) == expected

for s, expected in [
    ("xyz", "xyz"),
    ("__", "  "),
    ("a_b", "a b"),
    ]:
    assert binascii.a2b_qp(s, header=True) == expected

# TODO: str.find has to be implemented
# for s, flags, expected in [
#     # these are the tests from CPython 2.7
#     ("\xff\r\n\xff\n\xff", {}, "=FF\r\n=FF\r\n=FF"),
#     ("0"*75+"\xff\r\n\xff\r\n\xff",{},"0"*75+"=\r\n=FF\r\n=FF\r\n=FF"),
#     ('\0\n', {}, '=00\n'),
#     ('\0\n', {'quotetabs': True}, '=00\n'),
#     ('foo\tbar\t\n', {}, 'foo\tbar=09\n'),
#     ('foo\tbar\t\n', {'quotetabs': True}, 'foo=09bar=09\n'),
#     ('.', {}, '=2E'),
#     ('.\n', {}, '=2E\n'),
#     ('a.\n', {}, 'a.\n'),
#     # more tests follow
#     ('_', {}, '_'),
#     ('_', {'header': True}, '=5F'),
#     ('.x', {}, '.x'),
#     ('.\r\nn', {}, '=2E\r\nn'),
#     ('\nn', {}, '\nn'),
#     ('\r\nn', {}, '\r\nn'),
#     ('\nn', {'istext': False}, '=0An'),
#     ('\r\nn', {'istext': False}, '=0D=0An'),
#     (' ', {}, '=20'),
#     ('\t', {}, '=09'),
#     (' x', {}, ' x'),
#     ('\tx', {}, '\tx'),
#     ('\x16x', {}, '=16x'),
#     (' x', {'quotetabs': True}, '=20x'),
#     ('\tx', {'quotetabs': True}, '=09x'),
#     (' \nn', {}, '=20\nn'),
#     ('\t\nn', {}, '=09\nn'),
#     ('x\nn', {}, 'x\nn'),
#     (' \r\nn', {}, '=20\r\nn'),
#     ('\t\r\nn', {}, '=09\r\nn'),
#     ('x\r\nn', {}, 'x\r\nn'),
#     ('x\nn', {'istext': False}, 'x=0An'),
#     ('   ', {}, '  =20'),
#     ('   ', {'header': True}, '__=20'),
#     ('   \nn', {}, '  =20\nn'),
#     ('   \nn', {'header': True}, '___\nn'),
#     ('   ', {}, '  =20'),
#     ('\t\t\t', {'header': True}, '\t\t=09'),
#     ('\t\t\t\nn', {}, '\t\t=09\nn'),
#     ('\t\t\t\nn', {'header': True}, '\t\t=09\nn'),
#     ]:
#     assert binascii.b2a_qp(s, **flags) == expected

for s, expected, done in [
    ("", "", 0),
    ("AAAA", "]u\xd7", 0),
    ("A\nA\rAA", "]u\xd7", 0),
    (":", "", 1),
    ("A:", "", 1),
    ("AA:", "]", 1),
    ("AAA:", "]u", 1),
    ("AAAA:", "]u\xd7", 1),
    ("AAAA:foobarbaz", "]u\xd7", 1),
    ("41-CZ:", "D\xe3\x19", 1),
    ("41-CZl:", "D\xe3\x19\xbb", 1),
    ("41-CZlm:", "D\xe3\x19\xbb\xbf", 1),
    ("41-CZlm@:", "D\xe3\x19\xbb\xbf\x16", 1),
    ]:
    assert binascii.a2b_hqx(s) == (expected, done)

for s, expected in [
    ("", ""),
    ("A", "33"),
    ("AB", "38)"),
    ("ABC", "38*$"),
    ("ABCD", "38*$4!"),
    ("ABCDE", "38*$4%8"),
    ("ABCDEF", "38*$4%9'"),
    ("ABCDEFG", "38*$4%9'4`"),
    ("]u\xd7", "AAAA"),
    ]:
    assert binascii.b2a_hqx(s) == expected

# TODO: fix failing tests
for s, expected in [
    ("", ""),
    ("hello world", "hello world"),
    ("\x90\x00", "\x90"),
    ("a\x90\x05", "a" * 5),
    ("a\x90\xff", "a" * 0xFF),
    ("abc\x90\x01def", "abcdef"),
    ("abc\x90\x02def", "abccdef"),
    # ("abc\x90\x03def", "abcccdef"),
    ("abc\x90\xa1def", "ab" + "c" * 0xA1 + "def"),
    # ("abc\x90\x03\x90\x02def", "abccccdef"),
    ("abc\x90\x00\x90\x03def", "abc\x90\x90\x90def"),
    ("abc\x90\x03\x90\x00def", "abccc\x90def"),
    ]:
    assert binascii.rledecode_hqx(s) == expected

for s, expected in [
    ("", ""),
    ("hello world", "hello world"),
    ("helllo world", "helllo world"),
    ("hellllo world", "hel\x90\x04o world"),
    ("helllllo world", "hel\x90\x05o world"),
    ("aaa", "aaa"),
    ("aaaa", "a\x90\x04"),
    ("a" * 0xff, "a\x90\xff"),
    ("a" * 0x100, "a\x90\xffa"),
    ("a" * 0x101, "a\x90\xffaa"),
    ("a" * 0x102, "a\x90\xffaaa"),      # see comments in the source
    ("a" * 0x103, "a\x90\xffa\x90\x04"),
    ("a" * 0x1fe, "a\x90\xffa\x90\xff"),
    ("a" * 0x1ff, "a\x90\xffa\x90\xffa"),
    ("\x90", "\x90\x00"),
    ("\x90" * 2, "\x90\x00" * 2),
    ("\x90" * 3, "\x90\x00" * 3),       # see comments in the source
    # TODO: fix this test
    # ("\x90" * 345, "\x90\x00" * 345),
    ]:
    assert binascii.rlecode_hqx(s) == expected

for s, initial, expected in [
    ("", 0, 0),
    ("", 123, 123),
    ("hello", 321, 28955),
    ("world", 65535, 12911),
    ("uh", 40102, 37544),
    ('a', 10000, 14338),
    ('b', 10000, 2145),
    ('c', 10000, 6208),
    ('d', 10000, 26791),
    ('e', 10000, 30854),
    ('f', 10000, 18661),
    ('g', 10000, 22724),
    ('h', 10000, 43307),
    ('i', 10000, 47370),
    ('j', 10000, 35177),
    ('k', 10000, 39240),
    ('l', 10000, 59823),
    ('m', 10000, 63886),
    ('n', 10000, 51693),
    ('o', 10000, 55756),
    ('p', 10000, 14866),
    ('q', 10000, 10803),
    ('r', 10000, 6736),
    ('s', 10000, 2673),
    ('t', 10000, 31382),
    ('u', 10000, 27319),
    ('v', 10000, 23252),
    ('w', 10000, 19189),
    ('x', 10000, 47898),
    ('y', 10000, 43835),
    ('z', 10000, 39768),
    ]:
    assert binascii.crc_hqx(s, initial) == expected

for s, initial, expected in [
    ("", 0, 0),
    ("", 123, 123),
    ("hello", 321, -348147686),
    ("world", -2147483648, 32803080),
    ("world", 2147483647, 942244330),
    ('a', 10000, -184504832),
    ('b', 10000, 1812594618),
    ('c', 10000, 453955372),
    ('d', 10000, -2056627569),
    ('e', 10000, -227710439),
    ('f', 10000, 1801730979),
    ('g', 10000, 476252981),
    ('h', 10000, -1931733340),
    ('i', 10000, -69523918),
    ('j', 10000, 1657960328),
    ('k', 10000, 366298910),
    ('l', 10000, -1951280451),
    ('m', 10000, -55123413),
    ('n', 10000, 1707062161),
    ('o', 10000, 314082055),
    ('p', 10000, -1615819022),
    ('q', 10000, -390611356),
    ('r', 10000, 1908338654),
    ('s', 10000, 112844616),
    ('t', 10000, -1730327829),
    ('u', 10000, -270894467),
    ('v', 10000, 1993550791),
    ('w', 10000, 30677841),
    ('x', 10000, -1855256896),
    ('y', 10000, -429115818),
    ('z', 10000, 2137352172),
    ('foo', 99999999999999999999999999, -1932704816),
    ('bar', -99999999999999999999999999, 2000545409),
    ]:
    assert binascii.crc32(s, initial) == expected

for s, expected in [
    ("", ""),
    ("0", "30"),
    ("1", "31"),
    ("2", "32"),
    ("8", "38"),
    ("9", "39"),
    ("A", "41"),
    ("O", "4f"),
    ("\xde", "de"),
    ("ABC", "414243"),
    ("\x00\x00\x00\xff\x00\x00", "000000ff0000"),
    ("\x28\x9c\xc8\xc0\x3d\x8e", "289cc8c03d8e"),
    ]:
    assert binascii.hexlify(s) == expected
    assert binascii.b2a_hex(s) == expected

for s, expected in [
    ("", ""),
    ("30", "0"),
    ("31", "1"),
    ("32", "2"),
    ("38", "8"),
    ("39", "9"),
    ("41", "A"),
    ("4F", "O"),
    ("4f", "O"),
    ("DE", "\xde"),
    ("De", "\xde"),
    ("dE", "\xde"),
    ("de", "\xde"),
    ("414243", "ABC"),
    ("000000FF0000", "\x00\x00\x00\xff\x00\x00"),
    ("289cc8C03d8e", "\x28\x9c\xc8\xc0\x3d\x8e"),
    ]:
    assert binascii.unhexlify(s) == expected
    assert binascii.a2b_hex(s) == expected
