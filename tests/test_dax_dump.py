import struct
import unittest

from scripts.dax_dump import decode_rle, parse_entries


class DaxDumpTests(unittest.TestCase):
    def test_parse_one_block_and_decode_rle(self):
        # Header size is 2 + one 9-byte entry; block expands to AAA BC.
        entry = struct.pack("<BIHH", 7, 0, 5, 5)
        data = struct.pack("<H", 9) + entry + bytes((0xFD, 0x41, 0x01, 0x42, 0x43))
        data_offset, entries = parse_entries(data)
        self.assertEqual(data_offset, 11)
        self.assertEqual(entries[0].block_id, 7)
        self.assertEqual(decode_rle(data[data_offset:], 5), b"AAABC")

    def test_rejects_truncated_literal_run(self):
        with self.assertRaises(ValueError):
            decode_rle(bytes((0x02, 0x41)), 3)


if __name__ == "__main__":
    unittest.main()
