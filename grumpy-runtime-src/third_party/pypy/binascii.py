"""A pure Python implementation of binascii.

Rather slow and buggy in corner cases.
PyPy provides an RPython version too.
"""

class Error(Exception):
    pass

class Done(Exception):
    pass

class Incomplete(Exception):
    pass

def a2b_uu(s):
    if not s:
        return ''

    length = (ord(s[0]) - 0x20) % 64

    def quadruplets_gen(s):
        while s:
            try:
                yield ord(s[0]), ord(s[1]), ord(s[2]), ord(s[3])
            except IndexError:
                s += '   '
                yield ord(s[0]), ord(s[1]), ord(s[2]), ord(s[3])
                return
            s = s[4:]

    try:
        result = [''.join(
            [chr((A - 0x20) << 2 | (((B - 0x20) >> 4) & 0x3)),
            chr(((B - 0x20) & 0xf) << 4 | (((C - 0x20) >> 2) & 0xf)),
            chr(((C - 0x20) & 0x3) << 6 | ((D - 0x20) & 0x3f))
            ]) for A, B, C, D in quadruplets_gen(s[1:].rstrip())]
    except ValueError:
        raise Error('Illegal char')
    result = ''.join(result)
    trailingdata = result[length:]
    # if trailingdata.strip('\x00'):
    #     raise Error('Trailing garbage')
    result = result[:length]
    if len(result) < length:
        result += ((length - len(result)) * '\x00')
    return result


def b2a_uu(s):
    length = len(s)
    if length > 45:
        raise Error('At most 45 bytes at once')

    def triples_gen(s):
        while s:
            try:
                yield ord(s[0]), ord(s[1]), ord(s[2])
            except IndexError:
                s += '\0\0'
                yield ord(s[0]), ord(s[1]), ord(s[2])
                return
            s = s[3:]

    result = [''.join(
        [chr(0x20 + (( A >> 2                    ) & 0x3F)),
         chr(0x20 + (((A << 4) | ((B >> 4) & 0xF)) & 0x3F)),
         chr(0x20 + (((B << 2) | ((C >> 6) & 0x3)) & 0x3F)),
         chr(0x20 + (( C                         ) & 0x3F))])
              for A, B, C in triples_gen(s)]
    return chr(ord(' ') + (length & 077)) + ''.join(result) + '\n'


table_a2b_base64 = {
    'A': 0,
    'B': 1,
    'C': 2,
    'D': 3,
    'E': 4,
    'F': 5,
    'G': 6,
    'H': 7,
    'I': 8,
    'J': 9,
    'K': 10,
    'L': 11,
    'M': 12,
    'N': 13,
    'O': 14,
    'P': 15,
    'Q': 16,
    'R': 17,
    'S': 18,
    'T': 19,
    'U': 20,
    'V': 21,
    'W': 22,
    'X': 23,
    'Y': 24,
    'Z': 25,
    'a': 26,
    'b': 27,
    'c': 28,
    'd': 29,
    'e': 30,
    'f': 31,
    'g': 32,
    'h': 33,
    'i': 34,
    'j': 35,
    'k': 36,
    'l': 37,
    'm': 38,
    'n': 39,
    'o': 40,
    'p': 41,
    'q': 42,
    'r': 43,
    's': 44,
    't': 45,
    'u': 46,
    'v': 47,
    'w': 48,
    'x': 49,
    'y': 50,
    'z': 51,
    '0': 52,
    '1': 53,
    '2': 54,
    '3': 55,
    '4': 56,
    '5': 57,
    '6': 58,
    '7': 59,
    '8': 60,
    '9': 61,
    '+': 62,
    '/': 63,
    '=': 0,
}


def a2b_base64(s):
    if not isinstance(s, (str, unicode)):
        raise TypeError("expected string or unicode, got %r" % (s,))
    s = s.rstrip()
    # clean out all invalid characters, this also strips the final '=' padding
    # check for correct padding

    def next_valid_char(s, pos):
        for i in range(pos + 1, len(s)):
            c = s[i]
            if c < '\x7f':
                try:
                    table_a2b_base64[c]
                    return c
                except KeyError:
                    pass
        return None

    quad_pos = 0
    leftbits = 0
    leftchar = 0
    res = []
    for i, c in enumerate(s):
        if c > '\x7f' or c == '\n' or c == '\r' or c == ' ':
            continue
        if c == '=':
            if quad_pos < 2 or (quad_pos == 2 and next_valid_char(s, i) != '='):
                continue
            else:
                leftbits = 0
                break
        try:
            next_c = table_a2b_base64[c]
        except KeyError:
            continue
        quad_pos = (quad_pos + 1) & 0x03
        leftchar = (leftchar << 6) | next_c
        leftbits += 6
        if leftbits >= 8:
            leftbits -= 8
            res.append((leftchar >> leftbits & 0xff))
            leftchar &= ((1 << leftbits) - 1)
    if leftbits != 0:
        raise Error('Incorrect padding')

    return ''.join([chr(i) for i in res])

table_b2a_base64 = \
"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

def b2a_base64(s):
    length = len(s)
    final_length = length % 3

    def triples_gen(s):
        while s:
            try:
                yield ord(s[0]), ord(s[1]), ord(s[2])
            except IndexError:
                s += '\0\0'
                yield ord(s[0]), ord(s[1]), ord(s[2])
                return
            s = s[3:]


    a = triples_gen(s[ :length - final_length])

    result = [''.join(
        [table_b2a_base64[( A >> 2                    ) & 0x3F],
         table_b2a_base64[((A << 4) | ((B >> 4) & 0xF)) & 0x3F],
         table_b2a_base64[((B << 2) | ((C >> 6) & 0x3)) & 0x3F],
         table_b2a_base64[( C                         ) & 0x3F]])
              for A, B, C in a]

    final = s[length - final_length:]
    if final_length == 0:
        snippet = ''
    elif final_length == 1:
        a = ord(final[0])
        snippet = table_b2a_base64[(a >> 2 ) & 0x3F] + \
                  table_b2a_base64[(a << 4 ) & 0x3F] + '=='
    else:
        a = ord(final[0])
        b = ord(final[1])
        snippet = table_b2a_base64[(a >> 2) & 0x3F] + \
                  table_b2a_base64[((a << 4) | (b >> 4) & 0xF) & 0x3F] + \
                  table_b2a_base64[(b << 2) & 0x3F] + '='
    return ''.join(result) + snippet + '\n'

def a2b_qp(s, header=False):
    inp = 0
    odata = []
    while inp < len(s):
        if s[inp] == '=':
            inp += 1
            if inp >= len(s):
                break
            # Soft line breaks
            if (s[inp] == '\n') or (s[inp] == '\r'):
                if s[inp] != '\n':
                    while inp < len(s) and s[inp] != '\n':
                        inp += 1
                if inp < len(s):
                    inp += 1
            elif s[inp] == '=':
                # broken case from broken python qp
                odata.append('=')
                inp += 1
            elif s[inp] in hex_numbers and s[inp + 1] in hex_numbers:
                ch = chr(int(s[inp:inp+2], 16))
                inp += 2
                odata.append(ch)
            else:
                odata.append('=')
        elif header and s[inp] == '_':
            odata.append(' ')
            inp += 1
        else:
            odata.append(s[inp])
            inp += 1
    return ''.join(odata)

def b2a_qp(data, quotetabs=False, istext=True, header=False):
    """quotetabs=True means that tab and space characters are always
       quoted.
       istext=False means that \r and \n are treated as regular characters
       header=True encodes space characters with '_' and requires
       real '_' characters to be quoted.
    """
    MAXLINESIZE = 76

    # See if this string is using CRLF line ends
    lf = data.find('\n')
    crlf = lf > 0 and data[lf-1] == '\r'

    inp = 0
    linelen = 0
    odata = []
    while inp < len(data):
        c = data[inp]
        if (c > '~' or
            c == '=' or
            (header and c == '_') or
            (c == '.' and linelen == 0 and (inp+1 == len(data) or
                                            data[inp+1] == '\n' or
                                            data[inp+1] == '\r')) or
            (not istext and (c == '\r' or c == '\n')) or
            ((c == '\t' or c == ' ') and (inp + 1 == len(data))) or
            (c <= ' ' and c != '\r' and c != '\n' and
             (quotetabs or (not quotetabs and (c != '\t' and c != ' '))))):
            linelen += 3
            if linelen >= MAXLINESIZE:
                odata.append('=')
                if crlf: odata.append('\r')
                odata.append('\n')
                linelen = 3
            odata.append('=' + two_hex_digits(ord(c)))
            inp += 1
        else:
            if (istext and
                (c == '\n' or (inp+1 < len(data) and c == '\r' and
                               data[inp+1] == '\n'))):
                linelen = 0
                # Protect against whitespace on end of line
                if (len(odata) > 0 and
                    (odata[-1] == ' ' or odata[-1] == '\t')):
                    ch = ord(odata[-1])
                    odata[-1] = '='
                    odata.append(two_hex_digits(ch))

                if crlf: odata.append('\r')
                odata.append('\n')
                if c == '\r':
                    inp += 2
                else:
                    inp += 1
            else:
                if (inp + 1 < len(data) and
                    data[inp+1] != '\n' and
                    (linelen + 1) >= MAXLINESIZE):
                    odata.append('=')
                    if crlf: odata.append('\r')
                    odata.append('\n')
                    linelen = 0

                linelen += 1
                if header and c == ' ':
                    c = '_'
                odata.append(c)
                inp += 1
    return ''.join(odata)

hex_numbers = '0123456789ABCDEF'
def hex(n):
    if n == 0:
        return '0'

    if n < 0:
        n = -n
        sign = '-'
    else:
        sign = ''
    arr = []

    def hex_gen(n):
        """ Yield a nibble at a time. """
        while n:
            yield n % 0x10
            n = n / 0x10

    for nibble in hex_gen(n):
        arr = [hex_numbers[nibble]] + arr
    return sign + ''.join(arr)

def two_hex_digits(n):
    return hex_numbers[n / 0x10] + hex_numbers[n % 0x10]


def strhex_to_int(s):
    i = 0
    for c in s:
        i = i * 0x10 + hex_numbers.index(c)
    return i

hqx_encoding = '!"#$%&\'()*+,-012345689@ABCDEFGHIJKLMNPQRSTUVXYZ[`abcdefhijklmpqr'

DONE = 0x7f
SKIP = 0x7e
FAIL = 0x7d

table_a2b_hqx = [
    #^@    ^A    ^B    ^C    ^D    ^E    ^F    ^G
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    #\b    \t    \n    ^K    ^L    \r    ^N    ^O
    FAIL, FAIL, SKIP, FAIL, FAIL, SKIP, FAIL, FAIL,
    #^P    ^Q    ^R    ^S    ^T    ^U    ^V    ^W
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    #^X    ^Y    ^Z    ^[    ^\    ^]    ^^    ^_
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    #      !     "     #     $     %     &     '
    FAIL, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
    #(     )     *     +     ,     -     .     /
    0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, FAIL, FAIL,
    #0     1     2     3     4     5     6     7
    0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, FAIL,
    #8     9     :     ;     <     =     >     ?
    0x14, 0x15, DONE, FAIL, FAIL, FAIL, FAIL, FAIL,
    #@     A     B     C     D     E     F     G
    0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D,
    #H     I     J     K     L     M     N     O
    0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, FAIL,
    #P     Q     R     S     T     U     V     W
    0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, FAIL,
    #X     Y     Z     [     \     ]     ^     _
    0x2C, 0x2D, 0x2E, 0x2F, FAIL, FAIL, FAIL, FAIL,
    #`     a     b     c     d     e     f     g
    0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, FAIL,
    #h     i     j     k     l     m     n     o
    0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, FAIL, FAIL,
    #p     q     r     s     t     u     v     w
    0x3D, 0x3E, 0x3F, FAIL, FAIL, FAIL, FAIL, FAIL,
    #x     y     z     {     |     }     ~    ^?
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
    FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL, FAIL,
]

def a2b_hqx(s):
    result = []

    def quadruples_gen(s):
        t = []
        for c in s:
            res = table_a2b_hqx[ord(c)]
            if res == SKIP:
                continue
            elif res == FAIL:
                raise Error('Illegal character')
            elif res == DONE:
                yield t
                raise Done
            else:
                t.append(res)
            if len(t) == 4:
                yield t
                t = []
        yield t

    done = 0
    try:
        for snippet in quadruples_gen(s):
            length = len(snippet)
            if length == 4:
                result.append(chr(((snippet[0] & 0x3f) << 2) | (snippet[1] >> 4)))
                result.append(chr(((snippet[1] & 0x0f) << 4) | (snippet[2] >> 2)))
                result.append(chr(((snippet[2] & 0x03) << 6) | (snippet[3])))
            elif length == 3:
                result.append(chr(((snippet[0] & 0x3f) << 2) | (snippet[1] >> 4)))
                result.append(chr(((snippet[1] & 0x0f) << 4) | (snippet[2] >> 2)))
            elif length == 2:
                result.append(chr(((snippet[0] & 0x3f) << 2) | (snippet[1] >> 4)))
    except Done:
        done = 1
    except Error:
        raise
    return (''.join(result), done)

def b2a_hqx(s):
    result =[]

    def triples_gen(s):
        while s:
            try:
                yield ord(s[0]), ord(s[1]), ord(s[2])
            except IndexError:
                yield tuple([ord(c) for c in s])
            s = s[3:]

    for snippet in triples_gen(s):
        length = len(snippet)
        if length == 3:
            result.append(
                hqx_encoding[(snippet[0] & 0xfc) >> 2])
            result.append(hqx_encoding[
                ((snippet[0] & 0x03) << 4) | ((snippet[1] & 0xf0) >> 4)])
            result.append(hqx_encoding[
                (snippet[1] & 0x0f) << 2 | ((snippet[2] & 0xc0) >> 6)])
            result.append(hqx_encoding[snippet[2] & 0x3f])
        elif length == 2:
            result.append(
                hqx_encoding[(snippet[0] & 0xfc) >> 2])
            result.append(hqx_encoding[
                ((snippet[0] & 0x03) << 4) | ((snippet[1] & 0xf0) >> 4)])
            result.append(hqx_encoding[
                (snippet[1] & 0x0f) << 2])
        elif length == 1:
            result.append(
                hqx_encoding[(snippet[0] & 0xfc) >> 2])
            result.append(hqx_encoding[
                ((snippet[0] & 0x03) << 4)])
    return ''.join(result)

crctab_hqx = [
        0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
        0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
        0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
        0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
        0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
        0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
        0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
        0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
        0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
        0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
        0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
        0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
        0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
        0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
        0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
        0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
        0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
        0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
        0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
        0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
        0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
        0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
        0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
        0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
        0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
        0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
        0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
        0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
        0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
        0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
        0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
        0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0,
]

def crc_hqx(s, crc):
    for c in s:
        crc = ((crc << 8) & 0xff00) ^ crctab_hqx[((crc >> 8) & 0xff) ^ ord(c)]

    return crc

def rlecode_hqx(s):
    """
    Run length encoding for binhex4.
    The CPython implementation does not do run length encoding
    of \x90 characters. This implementation does.
    """
    if not s:
        return ''
    result = []
    prev = s[0]
    count = 1
    # Add a dummy character to get the loop to go one extra round.
    # The dummy must be different from the last character of s.
    # In the same step we remove the first character, which has
    # already been stored in prev.
    if s[-1] == '!':
        s = s[1:] + '?'
    else:
        s = s[1:] + '!'

    for c in s:
        if c == prev and count < 255:
            count += 1
        else:
            if count == 1:
                if prev != '\x90':
                    result.append(prev)
                else:
                    result += ['\x90', '\x00']
            elif count < 4:
                if prev != '\x90':
                    result += [prev] * count
                else:
                    result += ['\x90', '\x00'] * count
            else:
                if prev != '\x90':
                    result += [prev, '\x90', chr(count)]
                else:
                    result += ['\x90', '\x00', '\x90', chr(count)]
            count = 1
            prev = c

    return ''.join(result)

def rledecode_hqx(s):
    s = s.split('\x90')
    result = [s[0]]
    prev = s[0]
    for snippet in s[1:]:
        count = ord(snippet[0])
        if count > 0:
            result.append(prev[-1] * (count-1))
            prev = snippet
        else:
            result. append('\x90')
            prev = '\x90'
        result.append(snippet[1:])

    return ''.join(result)

crc_32_tab = [
    0x00000000L, 0x77073096L, 0xee0e612cL, 0x990951baL, 0x076dc419L,
    0x706af48fL, 0xe963a535L, 0x9e6495a3L, 0x0edb8832L, 0x79dcb8a4L,
    0xe0d5e91eL, 0x97d2d988L, 0x09b64c2bL, 0x7eb17cbdL, 0xe7b82d07L,
    0x90bf1d91L, 0x1db71064L, 0x6ab020f2L, 0xf3b97148L, 0x84be41deL,
    0x1adad47dL, 0x6ddde4ebL, 0xf4d4b551L, 0x83d385c7L, 0x136c9856L,
    0x646ba8c0L, 0xfd62f97aL, 0x8a65c9ecL, 0x14015c4fL, 0x63066cd9L,
    0xfa0f3d63L, 0x8d080df5L, 0x3b6e20c8L, 0x4c69105eL, 0xd56041e4L,
    0xa2677172L, 0x3c03e4d1L, 0x4b04d447L, 0xd20d85fdL, 0xa50ab56bL,
    0x35b5a8faL, 0x42b2986cL, 0xdbbbc9d6L, 0xacbcf940L, 0x32d86ce3L,
    0x45df5c75L, 0xdcd60dcfL, 0xabd13d59L, 0x26d930acL, 0x51de003aL,
    0xc8d75180L, 0xbfd06116L, 0x21b4f4b5L, 0x56b3c423L, 0xcfba9599L,
    0xb8bda50fL, 0x2802b89eL, 0x5f058808L, 0xc60cd9b2L, 0xb10be924L,
    0x2f6f7c87L, 0x58684c11L, 0xc1611dabL, 0xb6662d3dL, 0x76dc4190L,
    0x01db7106L, 0x98d220bcL, 0xefd5102aL, 0x71b18589L, 0x06b6b51fL,
    0x9fbfe4a5L, 0xe8b8d433L, 0x7807c9a2L, 0x0f00f934L, 0x9609a88eL,
    0xe10e9818L, 0x7f6a0dbbL, 0x086d3d2dL, 0x91646c97L, 0xe6635c01L,
    0x6b6b51f4L, 0x1c6c6162L, 0x856530d8L, 0xf262004eL, 0x6c0695edL,
    0x1b01a57bL, 0x8208f4c1L, 0xf50fc457L, 0x65b0d9c6L, 0x12b7e950L,
    0x8bbeb8eaL, 0xfcb9887cL, 0x62dd1ddfL, 0x15da2d49L, 0x8cd37cf3L,
    0xfbd44c65L, 0x4db26158L, 0x3ab551ceL, 0xa3bc0074L, 0xd4bb30e2L,
    0x4adfa541L, 0x3dd895d7L, 0xa4d1c46dL, 0xd3d6f4fbL, 0x4369e96aL,
    0x346ed9fcL, 0xad678846L, 0xda60b8d0L, 0x44042d73L, 0x33031de5L,
    0xaa0a4c5fL, 0xdd0d7cc9L, 0x5005713cL, 0x270241aaL, 0xbe0b1010L,
    0xc90c2086L, 0x5768b525L, 0x206f85b3L, 0xb966d409L, 0xce61e49fL,
    0x5edef90eL, 0x29d9c998L, 0xb0d09822L, 0xc7d7a8b4L, 0x59b33d17L,
    0x2eb40d81L, 0xb7bd5c3bL, 0xc0ba6cadL, 0xedb88320L, 0x9abfb3b6L,
    0x03b6e20cL, 0x74b1d29aL, 0xead54739L, 0x9dd277afL, 0x04db2615L,
    0x73dc1683L, 0xe3630b12L, 0x94643b84L, 0x0d6d6a3eL, 0x7a6a5aa8L,
    0xe40ecf0bL, 0x9309ff9dL, 0x0a00ae27L, 0x7d079eb1L, 0xf00f9344L,
    0x8708a3d2L, 0x1e01f268L, 0x6906c2feL, 0xf762575dL, 0x806567cbL,
    0x196c3671L, 0x6e6b06e7L, 0xfed41b76L, 0x89d32be0L, 0x10da7a5aL,
    0x67dd4accL, 0xf9b9df6fL, 0x8ebeeff9L, 0x17b7be43L, 0x60b08ed5L,
    0xd6d6a3e8L, 0xa1d1937eL, 0x38d8c2c4L, 0x4fdff252L, 0xd1bb67f1L,
    0xa6bc5767L, 0x3fb506ddL, 0x48b2364bL, 0xd80d2bdaL, 0xaf0a1b4cL,
    0x36034af6L, 0x41047a60L, 0xdf60efc3L, 0xa867df55L, 0x316e8eefL,
    0x4669be79L, 0xcb61b38cL, 0xbc66831aL, 0x256fd2a0L, 0x5268e236L,
    0xcc0c7795L, 0xbb0b4703L, 0x220216b9L, 0x5505262fL, 0xc5ba3bbeL,
    0xb2bd0b28L, 0x2bb45a92L, 0x5cb36a04L, 0xc2d7ffa7L, 0xb5d0cf31L,
    0x2cd99e8bL, 0x5bdeae1dL, 0x9b64c2b0L, 0xec63f226L, 0x756aa39cL,
    0x026d930aL, 0x9c0906a9L, 0xeb0e363fL, 0x72076785L, 0x05005713L,
    0x95bf4a82L, 0xe2b87a14L, 0x7bb12baeL, 0x0cb61b38L, 0x92d28e9bL,
    0xe5d5be0dL, 0x7cdcefb7L, 0x0bdbdf21L, 0x86d3d2d4L, 0xf1d4e242L,
    0x68ddb3f8L, 0x1fda836eL, 0x81be16cdL, 0xf6b9265bL, 0x6fb077e1L,
    0x18b74777L, 0x88085ae6L, 0xff0f6a70L, 0x66063bcaL, 0x11010b5cL,
    0x8f659effL, 0xf862ae69L, 0x616bffd3L, 0x166ccf45L, 0xa00ae278L,
    0xd70dd2eeL, 0x4e048354L, 0x3903b3c2L, 0xa7672661L, 0xd06016f7L,
    0x4969474dL, 0x3e6e77dbL, 0xaed16a4aL, 0xd9d65adcL, 0x40df0b66L,
    0x37d83bf0L, 0xa9bcae53L, 0xdebb9ec5L, 0x47b2cf7fL, 0x30b5ffe9L,
    0xbdbdf21cL, 0xcabac28aL, 0x53b39330L, 0x24b4a3a6L, 0xbad03605L,
    0xcdd70693L, 0x54de5729L, 0x23d967bfL, 0xb3667a2eL, 0xc4614ab8L,
    0x5d681b02L, 0x2a6f2b94L, 0xb40bbe37L, 0xc30c8ea1L, 0x5a05df1bL,
    0x2d02ef8dL
]

def crc32(s, crc=0):
    result = 0
    crc = ~long(crc) & 0xffffffffL
    for c in s:
        crc = crc_32_tab[(crc ^ long(ord(c))) & 0xffL] ^ (crc >> 8)
        #/* Note:  (crc >> 8) MUST zero fill on left

    result = crc ^ 0xffffffffL

    if result > (1 << 31):
        result = ((result + (1<<31)) % (1<<32)) - (1<<31)

    return result

def b2a_hex(s):
    result = []
    for char in s:
        c = (ord(char) >> 4) & 0xf
        if c > 9:
            c = c + ord('a') - 10
        else:
            c = c + ord('0')
        result.append(chr(c))
        c = ord(char) & 0xf
        if c > 9:
            c = c + ord('a') - 10
        else:
            c = c + ord('0')
        result.append(chr(c))
    return ''.join(result)

hexlify = b2a_hex

table_hex = [
    -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    0, 1, 2, 3,  4, 5, 6, 7,  8, 9,-1,-1, -1,-1,-1,-1,
    -1,10,11,12, 13,14,15,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    -1,10,11,12, 13,14,15,-1, -1,-1,-1,-1, -1,-1,-1,-1,
    -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1
]


def a2b_hex(t):
    result = []

    def pairs_gen(s):
        while s:
            try:
                yield table_hex[ord(s[0])], table_hex[ord(s[1])]
            except IndexError:
                if len(s):
                    raise TypeError('Odd-length string')
                return
            s = s[2:]

    for a, b in pairs_gen(t):
        if a < 0 or b < 0:
            raise TypeError('Non-hexadecimal digit found')
        result.append(chr((a << 4) + b))
    return ''.join(result)


unhexlify = a2b_hex