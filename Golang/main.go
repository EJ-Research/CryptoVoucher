
// Author: (EJ)
// Description: This code provides functionalities for encoding private keys into vouchers 
// and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
// across multiple programming languages and platforms.
//
// License: MIT
// Feel free to use, modify, or distribute this code under the terms of the MIT License.


package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// CryptoVoucher provides methods for encoding private keys into vouchers
// and restoring private keys from vouchers using Base62 encoding
type CryptoVoucher struct {
	base62Chars string
}

// NewCryptoVoucher initializes a new CryptoVoucher instance
func NewCryptoVoucher() *CryptoVoucher {
	return &CryptoVoucher{
		base62Chars: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
	}
}

// validateHex ensures the input is a valid hexadecimal string of the specified length
func (cv *CryptoVoucher) validateHex(input string, length int) error {
	if len(input) != length {
		return fmt.Errorf("input must be %d characters long!", length)
	}
	if _, err := hex.DecodeString(input); err != nil {
		return errors.New("input must be a valid hexadecimal string!")
	}
	return nil
}

// base62Encode converts a hexadecimal string into a Base62 encoded string
func (cv *CryptoVoucher) base62Encode(hexInput string) (string, error) {
	// Validate the hexadecimal input
	if err := cv.validateHex(hexInput, 64); err != nil {
		return "", err
	}

	hexNum := new(big.Int)
	hexNum.SetString(hexInput, 16)
	encoded := ""
	base := big.NewInt(62)
	zero := big.NewInt(0)

	// Convert the number from base 16 to base 62
	for hexNum.Cmp(zero) > 0 {
		remainder := new(big.Int)
		hexNum.DivMod(hexNum, base, remainder)
		encoded = string(cv.base62Chars[remainder.Int64()]) + encoded
	}

	return encoded, nil
}

// base62Decode converts a Base62 encoded string back into a hexadecimal string
func (cv *CryptoVoucher) base62Decode(base62Input string) (string, error) {
	// Validate the Base62 input
	if matched, _ := regexp.MatchString("^[0-9A-Za-z]+$", base62Input); !matched {
		return "", errors.New("invalid Base62 input. Only alphanumeric characters are allowed!")
	}

	decimal := big.NewInt(0)
	base := big.NewInt(62)

	// Convert the Base62 string back to a decimal number
	for _, char := range base62Input {
		index := strings.Index(cv.base62Chars, string(char))
		if index == -1 {
			return "", errors.New("invalid character in Base62 input!")
		}
		digit := big.NewInt(int64(index))
		decimal.Mul(decimal, base).Add(decimal, digit)
	}

	// Convert the decimal number back to a hexadecimal string
	hexOutput := fmt.Sprintf("%064x", decimal)
	return hexOutput, nil
}

// CreateVoucher generates a voucher key and voucher code from a private key
func (cv *CryptoVoucher) CreateVoucher(privateKey string) (string, string, error) {
	// Convert the private key to a Base62 encoded string
	compressedKey, err := cv.base62Encode(privateKey)
	if err != nil {
		return "", "", err
	}

	// Ensure the encoded string is long enough to split
	if len(compressedKey) < 28 {
		return "", "", errors.New("encoded key is too short!")
	}

	// Split the encoded string into a voucher key and voucher code
	voucherKey := compressedKey[:28]
	voucherCode := compressedKey[28:]
	return voucherKey, voucherCode, nil
}

// RestorePrivateKey reconstructs a private key from a voucher key and voucher code
func (cv *CryptoVoucher) RestorePrivateKey(voucherKey, voucherCode string) (string, error) {
	// Combine the voucher key and voucher code to reconstruct the Base62 string
	compressedKey := voucherKey + voucherCode

	// Decode the Base62 string back to the original private key
	decodedKey, err := cv.base62Decode(compressedKey)
	if err != nil {
		return "", err
	}
	return decodedKey, nil
}

func main() {
	// Example private key to encode and decode
	privateKey := "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"
	cryptoVoucher := NewCryptoVoucher()

	// Create a voucher from the private key
	voucherKey, voucherCode, err := cryptoVoucher.CreateVoucher(privateKey)
	if err != nil {
		fmt.Println("Error creating voucher:", err)
		return
	}

	fmt.Println("Voucher Key:", voucherKey)
	fmt.Println("Voucher Code:", voucherCode)

	// Restore the private key from the voucher
	restoredKey, err := cryptoVoucher.RestorePrivateKey(voucherKey, voucherCode)
	if err != nil {
		fmt.Println("Error restoring private key:", err)
		return
	}

	fmt.Println("Restored Private Key:", strings.ToUpper(restoredKey))

	// Verify the restored key matches the original private key
	if strings.ToUpper(privateKey) == strings.ToUpper(restoredKey) {
		fmt.Println("Success!")
	} else {
		fmt.Println("Failed!")
	}
}
