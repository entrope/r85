# r85 Package

This package implements an encoding scheme similar to btoa's Ascii85 and
ZeroMQ's [32/Z85](https://rfc.zeromq.org/spec/32/).
The differences are in alphabet, handling of partial words and endianness.

## r85 Alphabet

The alphabet is encoded starting at ASCII `(` (decimal 40), but
replacing `<` with `}` and `` ` `` with `~`.
This minimizes escaping in most languages and HTML.
This avoids all of ``!"#$%&'` `` and their occasional uses as
metacharacters.

## r85 Binary-to-Text Translation

Each 4-octet block of binary data is interpreted as a little-endian 32-bit
number, and encoded into a 5-character little-endian block of r85 digits.
If the input array is not a multiple of four octets in length, the
remaining 1, 2 or 3 digits are treated as little-endian 8-, 16- or 24-bit
numbers and encoded into 2-, 3- or 4-character blocks instead.

## r85 Text-to-Binary Translation

Input text is processed by skipping characters outside the range of `(`
through `~`, and then collecting blocks up to 5 characters in length.
A shorter block is only allowed at the end of the input text.
`}` and `~` are replaced by `<` and `` ` `` respectively.
Translation fails if the final block is a single character long.
Otherwise, the 2-, 3-, 4- or 5-character block is interpreted as a
little-endian number.
If the number is larger than 8, 16, 24 or 32 bits (unsigned) respectively
then the input is considered corrupt.
Otherwise it is emitted as a 1-, 2-, 3- or 4-byte sequence in little-endian
order.
