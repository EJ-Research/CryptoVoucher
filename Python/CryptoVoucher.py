
# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers 
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.



import re

# secp256k1 curve order; valid private keys are in [1, N-1]
SECP256K1_N = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
# 62^43 > 2^256, so every private key fits in 43 Base62 characters
ENCODED_LENGTH = 43
# Network code followed by the first 28 Base62 characters
VOUCHER_KEY_LENGTH = 29
# Remaining 15 Base62 characters followed by one check character
VOUCHER_CODE_LENGTH = 16

# Network code (first character of every voucher) -> network ID.
# Codes are permanent: an assigned code is never changed or reused, new networks are only appended.
NETWORKS = {
    "1": "BITCOIN_BTC",
    "2": "ETHEREUM_ETH",
    "3": "ETHEREUM_USDT",
    "4": "ETHEREUM_USDC",
    "5": "TRON_TRX",
    "6": "TRON_USDT",
    "7": "BSC_BNB",
    "8": "BSC_USDT",
    "9": "BSC_USDC",
    "A": "POLYGON_POL",
    "B": "POLYGON_USDT",
    "C": "POLYGON_USDC",
    "D": "SOLANA_SOL",
    "E": "SOLANA_USDT",
    "F": "SOLANA_USDC",
    "G": "TON_TON",
    "H": "TON_USDT",
    "J": "ARBITRUM_ETH",
    "K": "ARBITRUM_USDT",
    "L": "ARBITRUM_USDC",
    "M": "OPTIMISM_ETH",
    "N": "OPTIMISM_USDT",
    "P": "OPTIMISM_USDC",
    "Q": "BASE_ETH",
    "R": "BASE_USDC",
    "S": "AVALANCHE_AVAX",
    "T": "AVALANCHE_USDT",
    "U": "AVALANCHE_USDC",
    "V": "LITECOIN_LTC",
    "W": "DOGECOIN_DOGE",
    "X": "BITCOINCASH_BCH",
    "Y": "XRPL_XRP",
    "a": "ETHEREUM_DAI",
    "b": "ETHEREUM_PYUSD",
}
NETWORK_CODES = {network: code for code, network in NETWORKS.items()}


class CryptoVoucher:
    def __init__(self):
        # Characters used for Base62 encoding
        self.base62_chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
        self.base62_index = {char: index for index, char in enumerate(self.base62_chars)}

    def validate_hex(self, input_hex, length):
        """
        Validates the input to ensure it is a hexadecimal string of the specified length
        """
        if not isinstance(input_hex, str):
            return False, "Input must be a string!"
        if len(input_hex) != length:
            return False, f"Input must be {length} characters long!"
        if not re.fullmatch(r'[0-9a-fA-F]+', input_hex):
            return False, "Input must be a valid hexadecimal string!"
        return True, None

    def base62_encode(self, hex_input):
        """
        Encodes a hexadecimal private key into a fixed-length Base62 string
        """
        is_valid, error = self.validate_hex(hex_input, 64)
        if not is_valid:
            return None, error

        # Convert hex string to a decimal number
        decimal_value = int(hex_input, 16)
        if not 0 < decimal_value < SECP256K1_N:
            return None, "Private key is out of secp256k1 range!"

        base62 = ""

        # Convert the decimal number to Base62
        while decimal_value > 0:
            remainder = decimal_value % 62
            base62 = self.base62_chars[remainder] + base62
            decimal_value //= 62

        # Leading zeros keep the length fixed and do not change the decoded value
        return base62.rjust(ENCODED_LENGTH, "0"), None

    def base62_decode(self, base62_input):
        """
        Decodes a Base62 string back into a hexadecimal private key
        """
        if not isinstance(base62_input, str):
            return None, "Base62 input must be a string!"
        if len(base62_input) != ENCODED_LENGTH:
            return None, f"Base62 input must be {ENCODED_LENGTH} characters long!"
        if not re.fullmatch(r'[0-9A-Za-z]+', base62_input):
            return None, "Invalid Base62 input. Only alphanumeric characters are allowed!"

        decimal_value = 0

        # Convert the Base62 string back to a decimal number
        for char in base62_input:
            decimal_value = decimal_value * 62 + self.base62_index[char]

        if not 0 < decimal_value < SECP256K1_N:
            return None, "Private key is out of secp256k1 range!"

        # Convert the decimal number to a hex string and pad to 64 characters
        hex_output = f"{decimal_value:064x}"
        return hex_output, None

    def check_char(self, base62_input):
        """
        Computes the check character of a validated Base62 string: the sum of
        each digit times its 1-based position, modulo 61 (prime, so every
        position weight is invertible and any single swap changes the sum)
        """
        total = 0
        for position, char in enumerate(base62_input, 1):
            total += position * self.base62_index[char]
        return self.base62_chars[total % 61]

    def create_voucher(self, private_key, network):
        """
        Creates a voucher key and voucher code from a private key and its network ID
        """
        code = NETWORK_CODES.get(network) if isinstance(network, str) else None
        if code is None:
            return None, None, "Unknown network!"

        base62_encoded, error = self.base62_encode(private_key)
        if error:
            return None, None, error

        voucher = code + base62_encoded
        voucher += self.check_char(voucher)
        voucher_key = voucher[:VOUCHER_KEY_LENGTH]
        voucher_code = voucher[VOUCHER_KEY_LENGTH:]
        return voucher_key, voucher_code, None

    def restore_private_key(self, voucher_key, voucher_code):
        """
        Restores the private key and its network ID from a voucher key and voucher code
        """
        if not isinstance(voucher_key, str) or not isinstance(voucher_code, str):
            return None, None, "Voucher key and voucher code must be strings!"
        # Checking each part also catches the two parts entered in swapped order
        if len(voucher_key) != VOUCHER_KEY_LENGTH:
            return None, None, f"Voucher key must be {VOUCHER_KEY_LENGTH} characters long!"
        if len(voucher_code) != VOUCHER_CODE_LENGTH:
            return None, None, f"Voucher code must be {VOUCHER_CODE_LENGTH} characters long!"

        voucher = voucher_key + voucher_code
        network = NETWORKS.get(voucher[0])
        if network is None:
            return None, None, "Unknown network code!"

        # Decode the Base62 string back to the original private key
        hex_decoded, error = self.base62_decode(voucher[1:ENCODED_LENGTH + 1])
        if error:
            return None, None, error

        # The check character covers the network code as well
        if self.check_char(voucher[:ENCODED_LENGTH + 1]) != voucher[ENCODED_LENGTH + 1]:
            return None, None, "Invalid voucher check character!"

        return hex_decoded, network, None

# Example usage


def main():
    private_key = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"
    crypto_voucher = CryptoVoucher()

    # Create a voucher for USDT on TRON
    voucher_key, voucher_code, error = crypto_voucher.create_voucher(
        private_key, "TRON_USDT")
    if error:
        print("Error creating voucher:", error)
        return

    print("Voucher Key:", voucher_key)
    print("Voucher Code:", voucher_code)

    # Restore the private key and network
    restored_key, network, error = crypto_voucher.restore_private_key(
        voucher_key, voucher_code)
    if error:
        print("Error restoring private key:", error)
        return

    print("Restored Private Key:", restored_key.upper())
    print("Network:", network)

    # Verify the restored key matches the original
    if restored_key.upper() == private_key.upper():
        print("Success!")
    else:
        print("Failed!")


if __name__ == "__main__":
    main()
