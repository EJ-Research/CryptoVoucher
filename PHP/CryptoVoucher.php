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

    // Encodes a hexadecimal string to a Base62 string
    private function _base62Encode($hex) {
        // Validate hex input
        $validation = $this->_validateHex($hex, 64);
        if ($validation['status'] === 'error') {
            return $validation;
        }

        $decimal = $this->_hexToDecimal($hex); // Convert hex to decimal
        $encoded = '';

        // Perform Base62 encoding
        while (bccomp($decimal, '0') > 0) {
            $remainder = bcmod($decimal, '62');
            $encoded = $this->base62Chars[$remainder] . $encoded;
            $decimal = bcdiv($decimal, '62', 0);
        }

        return ['status' => 'success', 'data' => $encoded];
    }

    // Decodes a Base62 string back to a hexadecimal string
    private function _base62Decode($base62) {
        // Validate Base62 input
        if (!preg_match('/^[0-9A-Za-z]+$/', $base62)) {
            return ['status' => 'error', 'message' => 'Invalid Base62 input. Only alphanumeric characters are allowed!'];
        }

        $decimal = '0';

        // Convert Base62 to decimal
        for ($i = 0; $i < strlen($base62); $i++) {
            $decimal = bcmul($decimal, '62', 0);
            $decimal = bcadd($decimal, strpos($this->base62Chars, $base62[$i]), 0);
        }

        // Convert decimal back to hexadecimal
        return ['status' => 'success', 'data' => $this->_decimalToHex($decimal)];
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
        while (bccomp($decimal, '0') > 0) {
            $remainder = bcmod($decimal, '16');
            $hex = dechex($remainder) . $hex;
            $decimal = bcdiv($decimal, '16', 0);
        }

        // Pad the result to ensure it is 64 characters long
        return str_pad($hex, 64, '0', STR_PAD_LEFT);
    }

    // Validates that the input is a valid hexadecimal string of the specified length
    private function _validateHex($input, $length) {
        if (strlen($input) !== $length || !ctype_xdigit($input)) {
            return ['status' => 'error', 'message' => "Input must be a $length-character hexadecimal string!"];
        }
        return ['status' => 'success'];
    }

    // Creates a voucher from a private key
    public function createVoucher($privateKey) {
        // Validate private key
        $validation = $this->_validateHex($privateKey, 64);
        if ($validation['status'] === 'error') {
            return $validation;
        }

        // Encode private key to Base62
        $compressedKey = $this->_base62Encode($privateKey);
        if ($compressedKey['status'] === 'error') {
            return $compressedKey;
        }

        // Split the Base62 string into a voucher key and voucher code
        $voucherKey = substr($compressedKey['data'], 0, 28);
        $voucherCode = substr($compressedKey['data'], 28);

        return ['status' => 'success', 'voucher_key' => $voucherKey, 'voucher_code' => $voucherCode];
    }

    // Restores the private key from a voucher key and voucher code
    public function restorePrivateKey($voucherKey, $voucherCode) {
        // Reconstruct the compressed Base62 string
        $compressedKey = $voucherKey . $voucherCode;

        // Validate the reconstructed string
        if (!preg_match('/^[0-9A-Za-z]+$/', $compressedKey)) {
            return ['status' => 'error', 'message' => 'Invalid voucher data. Reconstructed key contains invalid character!.'];
        }

        // Decode the Base62 string back to hexadecimal
        $data = $this->_base62Decode($compressedKey);
        if ($data['status'] === 'error') {
            return $data;
        }

        return ['status' => 'success', 'data' => $data['data']];
    }
}

// Sample usage
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