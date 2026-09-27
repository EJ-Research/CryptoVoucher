import json
import unittest
from pathlib import Path

from CryptoVoucher import CryptoVoucher

VECTORS = Path(__file__).resolve().parent.parent / "vectors.json"


class VectorTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.vectors = json.loads(VECTORS.read_text(encoding="utf-8"))
        cls.cv = CryptoVoucher()

    def test_valid(self):
        for v in self.vectors["valid"]:
            with self.subTest(private_key=v["private_key"], network=v["network"]):
                self.assertEqual(self.cv.create_voucher(v["private_key"], v["network"]),
                                 (v["voucher_key"], v["voucher_code"], None))
                self.assertEqual(self.cv.restore_private_key(v["voucher_key"], v["voucher_code"]),
                                 (v["restored_key"], v["network"], None))

    def test_create_errors(self):
        for v in self.vectors["create_errors"]:
            with self.subTest(private_key=v["private_key"], network=v["network"]):
                error = self.cv.create_voucher(v["private_key"], v["network"])[2]
                self.assertEqual((error or "").lower(), v["error"].lower())

    def test_restore_errors(self):
        for v in self.vectors["restore_errors"]:
            with self.subTest(v["note"]):
                error = self.cv.restore_private_key(v["voucher_key"], v["voucher_code"])[2]
                self.assertEqual((error or "").lower(), v["error"].lower())


if __name__ == "__main__":
    unittest.main()
