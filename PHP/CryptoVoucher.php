<?php

// Author: (EJ)
// Description: This code provides functionalities for encoding private keys into vouchers 
// and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
// across multiple programming languages and platforms.
//
// License: MIT
// Feel free to use, modify, or distribute this code under the terms of the MIT License.

class CryptoVoucher {

    // Characters used for Base62 encoding
    private $base62Chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';

    // secp256k1 curve order (decimal); valid private keys are in [1, N-1]
    private const SECP256K1_N = '115792089237316195423570985008687907852837564279074904382605163141518161494337';
    private const VOUCHER_KEY_LENGTH = 28;
    // 62^43 > 2^256, so every private key fits in 43 Base62 characters
    private const ENCODED_LENGTH = 43;
    // Remaining Base62 characters followed by one check character
    private const VOUCHER_CODE_LENGTH = self::ENCODED_LENGTH - self::VOUCHER_KEY_LENGTH + 1;

    // Encodes a hexadecimal private key to a fixed-length Base62 string
    private function _base62Encode($hex) {
        // Validate hex input
        $validation = $this->_validateHex($hex, 64);
        if ($validation['status'] === 'error') {
            return $validation;
        }

        $decimal = $this->_hexToDecimal($hex); // Convert hex to decimal
        if (!$this->_isInKeyRange($decimal)) {
            return ['status' => 'error', 'message' => 'Private key is out of secp256k1 range!'];
        }

        $encoded = '';

        // Perform Base62 encoding (explicit scale 0 keeps results integral regardless of bcscale())
        while (bccomp($decimal, '0', 0) > 0) {
            $remainder = (int) bcmod($decimal, '62', 0);
            $encoded = $this->base62Chars[$remainder] . $encoded;
            $decimal = bcdiv($decimal, '62', 0);
        }

        // Leading zeros keep the length fixed and do not change the decoded value
        return ['status' => 'success', 'data' => str_pad($encoded, self::ENCODED_LENGTH, '0', STR_PAD_LEFT)];
    }

    // Decodes a Base62 string back to a hexadecimal private key
    private function _base62Decode($base62) {
        // Validate Base62 input
        if (strlen($base62) !== self::ENCODED_LENGTH) {
            return ['status' => 'error', 'message' => 'Base62 input must be ' . self::ENCODED_LENGTH . ' characters long!'];
        }
        if (!preg_match('/\A[0-9A-Za-z]+\z/', $base62)) {
            return ['status' => 'error', 'message' => 'Invalid Base62 input. Only alphanumeric characters are allowed!'];
        }

        $decimal = '0';

        // Convert Base62 to decimal
        for ($i = 0; $i < strlen($base62); $i++) {
            $decimal = bcmul($decimal, '62', 0);
            $decimal = bcadd($decimal, strpos($this->base62Chars, $base62[$i]), 0);
        }

        if (!$this->_isInKeyRange($decimal)) {
            return ['status' => 'error', 'message' => 'Private key is out of secp256k1 range!'];
        }

        // Convert decimal back to hexadecimal
        return ['status' => 'success', 'data' => $this->_decimalToHex($decimal)];
    }

    // Computes the check character of a validated Base62 string: the sum of each digit
    // times its 1-based position, modulo 61 (prime, so every position weight is
    // invertible and any single swap changes the sum)
    private function _checkChar($base62) {
        $total = 0;
        for ($i = 0; $i < strlen($base62); $i++) {
            $total += ($i + 1) * strpos($this->base62Chars, $base62[$i]);
        }
        return $this->base62Chars[$total % 61];
    }

    // Checks that a decimal string is a valid secp256k1 private key (1 <= key < N)
    private function _isInKeyRange($decimal) {
        return bccomp($decimal, '0', 0) > 0 && bccomp($decimal, self::SECP256K1_N, 0) < 0;
    }

    // Converts a hexadecimal string to a decimal string
    private function _hexToDecimal($hex) {
        $decimal = '0';
        for ($i = 0; $i < strlen($hex); $i++) {
            $decimal = bcadd(bcmul($decimal, '16', 0), hexdec($hex[$i]), 0);
        }
        return $decimal;
    }

    // Converts a decimal string to a hexadecimal string
    private function _decimalToHex($decimal) {
        $hex = '';
        while (bccomp($decimal, '0', 0) > 0) {
            $remainder = (int) bcmod($decimal, '16', 0);
            $hex = dechex($remainder) . $hex;
            $decimal = bcdiv($decimal, '16', 0);
        }

        // Pad the result to ensure it is 64 characters long
        return str_pad($hex, 64, '0', STR_PAD_LEFT);
    }

    // Validates that the input is a valid hexadecimal string of the specified length
    private function _validateHex($input, $length) {
        if (!is_string($input) || strlen($input) !== $length || !ctype_xdigit($input)) {
            return ['status' => 'error', 'message' => "Input must be a $length-character hexadecimal string!"];
        }
        return ['status' => 'success'];
    }

    // Creates a voucher from a private key
    public function createVoucher($privateKey) {
        // Validate and encode private key to Base62
        $compressedKey = $this->_base62Encode($privateKey);
        if ($compressedKey['status'] === 'error') {
            return $compressedKey;
        }

        // Append the check character and split into a voucher key and voucher code
        $voucher = $compressedKey['data'] . $this->_checkChar($compressedKey['data']);
        $voucherKey = substr($voucher, 0, self::VOUCHER_KEY_LENGTH);
        $voucherCode = substr($voucher, self::VOUCHER_KEY_LENGTH);

        return ['status' => 'success', 'voucher_key' => $voucherKey, 'voucher_code' => $voucherCode];
    }

    // Restores the private key from a voucher key and voucher code
    public function restorePrivateKey($voucherKey, $voucherCode) {
        if (!is_string($voucherKey) || !is_string($voucherCode)) {
            return ['status' => 'error', 'message' => 'Voucher key and voucher code must be strings!'];
        }
        // Checking each part also catches the two parts entered in swapped order
        if (strlen($voucherKey) !== self::VOUCHER_KEY_LENGTH) {
            return ['status' => 'error', 'message' => 'Voucher key must be ' . self::VOUCHER_KEY_LENGTH . ' characters long!'];
        }
        if (strlen($voucherCode) !== self::VOUCHER_CODE_LENGTH) {
            return ['status' => 'error', 'message' => 'Voucher code must be ' . self::VOUCHER_CODE_LENGTH . ' characters long!'];
        }

        // Reconstruct the full voucher
        $voucher = $voucherKey . $voucherCode;

        // Decode the Base62 string back to hexadecimal (validates characters and key range)
        $compressedKey = substr($voucher, 0, self::ENCODED_LENGTH);
        $data = $this->_base62Decode($compressedKey);
        if ($data['status'] === 'error') {
            return $data;
        }

        if ($this->_checkChar($compressedKey) !== $voucher[self::ENCODED_LENGTH]) {
            return ['status' => 'error', 'message' => 'Invalid voucher check character!'];
        }

        return ['status' => 'success', 'data' => $data['data']];
    }
}

// Sample usage (runs only when this file is executed directly)
if (PHP_SAPI === 'cli' && !empty($_SERVER['SCRIPT_FILENAME']) && realpath($_SERVER['SCRIPT_FILENAME']) === __FILE__) {
    $privateKey = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B";
    $cryptoVoucher = new CryptoVoucher();

    // Create a voucher from the private key
    $voucher = $cryptoVoucher->createVoucher($privateKey);
    if ($voucher['status'] === 'success') {
        echo "Voucher Key: " . $voucher['voucher_key'] . PHP_EOL;
        echo "Voucher Code: " . $voucher['voucher_code'] . PHP_EOL;

        // Restore the private key from the voucher
        $restored = $cryptoVoucher->restorePrivateKey($voucher['voucher_key'], $voucher['voucher_code']);
        if ($restored['status'] === 'success') {
            echo "Restored Private Key: " . strtoupper($restored['data']) . PHP_EOL;
            echo strtoupper($privateKey) === strtoupper($restored['data']) ? "Success!" . PHP_EOL : "Failed!" . PHP_EOL;
        } else {
            echo "Error: " . $restored['message'] . PHP_EOL;
        }
    } else {
        echo "Error: " . $voucher['message'] . PHP_EOL;
    }
}
