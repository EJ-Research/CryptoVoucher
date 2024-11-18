
# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers 
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.



import re
from math import log
from decimal import Decimal


class CryptoVoucher:
    def __init__(self):
        # Characters used for Base62 encoding
        self.base62_chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

    def validate_hex(self, input_hex, length):
        """
        Validates the input to ensure it is a hexadecimal string of the specified length
        """
        if len(input_hex) != length:
            return False, f"Input must be {length} characters long!"
        if not re.fullmatch(r'[0-9a-fA-F]+', input_hex):
            return False, "Input must be a valid hexadecimal string!"
        return True, None

    def base62_encode(self, hex_input):
        """
        Encodes a hexadecimal string into a Base62 string
        """
        is_valid, error = self.validate_hex(hex_input, 64)
        if not is_valid:
            return None, error

        # Convert hex string to a decimal number
        decimal_value = int(hex_input, 16)
        base62 = ""

        # Convert the decimal number to Base62
        while decimal_value > 0:
            remainder = decimal_value % 62
            base62 = self.base62_chars[remainder] + base62
            decimal_value //= 62

        return base62, None

    def base62_decode(self, base62_input):
        """
        Decodes a Base62 string back into a hexadecimal string
        """
        if not re.fullmatch(r'[0-9A-Za-z]+', base62_input):
            return None, "Invalid Base62 input. Only alphanumeric characters are allowed!"

        decimal_value = 0

        # Convert the Base62 string back to a decimal number
        for char in base62_input:
            index = self.base62_chars.index(char)
            decimal_value = decimal_value * 62 + index

        # Convert the decimal number to a hex string and pad to 64 characters
        hex_output = f"{decimal_value:064x}"
        return hex_output, None

    def create_voucher(self, private_key):
        """
        Creates a voucher key and voucher code from a private key
        """
        base62_encoded, error = self.base62_encode(private_key)
        if error:
            return None, None, error

        # Ensure the encoded key is long enough to split
        if len(base62_encoded) < 28:
            return None, None, "Encoded key is too short!"

        voucher_key = base62_encoded[:28]
        voucher_code = base62_encoded[28:]
        return voucher_key, voucher_code, None

    def restore_private_key(self, voucher_key, voucher_code):
        """
        Restores the private key from a voucher key and voucher code
        """
        base62_combined = voucher_key + voucher_code

        # Decode the Base62 string back to the original private key
        hex_decoded, error = self.base62_decode(base62_combined)
        if error:
            return None, error

        return hex_decoded, None

# Example usage


def main():
    private_key = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"
    crypto_voucher = CryptoVoucher()

    # Create a voucher
    voucher_key, voucher_code, error = crypto_voucher.create_voucher(
        private_key)
    if error:
        print("Error creating voucher:", error)
        return

    print("Voucher Key:", voucher_key)
    print("Voucher Code:", voucher_code)

    # Restore the private key
    restored_key, error = crypto_voucher.restore_private_key(
        voucher_key, voucher_code)
    if error:
        print("Error restoring private key:", error)
        return

    print("Restored Private Key:", restored_key.upper())

    # Verify the restored key matches the original
    if restored_key.upper() == private_key.upper():
        print("Success!")
    else:
        print("Failed!")


if __name__ == "__main__":
    main()
