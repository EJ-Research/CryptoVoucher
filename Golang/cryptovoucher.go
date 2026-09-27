
// Author: (EJ)
// Description: This code provides functionalities for encoding private keys into vouchers
// and restoring them using Base62 encoding. It is designed for flexibility and adaptability
// across multiple programming languages and platforms.
//
// License: MIT
// Feel free to use, modify, or distribute this code under the terms of the MIT License.


package cryptovoucher

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	voucherKeyLength = 28
	// 62^43 > 2^256, so every private key fits in 43 Base62 characters
	encodedLength = 43
	// Encoded key followed by one check character
	voucherLength = encodedLength + 1
)

// secp256k1 curve order; valid private keys are in [1, N-1]
var secp256k1N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

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

// isInKeyRange reports whether n is a valid secp256k1 private key
func isInKeyRange(n *big.Int) bool {
	return n.Sign() > 0 && n.Cmp(secp256k1N) < 0
}

// base62Encode converts a hexadecimal private key into a fixed-length Base62 encoded string
func (cv *CryptoVoucher) base62Encode(hexInput string) (string, error) {
	// Validate the hexadecimal input
	if err := cv.validateHex(hexInput, 64); err != nil {
		return "", err
	}

	hexNum, _ := new(big.Int).SetString(hexInput, 16)
	if !isInKeyRange(hexNum) {
		return "", errors.New("private key is out of secp256k1 range!")
	}

	// Leading zeros keep the length fixed and do not change the decoded value
	encoded := bytes.Repeat([]byte{'0'}, encodedLength)
	base := big.NewInt(62)
	remainder := new(big.Int)

	// Convert the number from base 16 to base 62, filling from the right
	for i := encodedLength - 1; hexNum.Sign() > 0; i-- {
		hexNum.DivMod(hexNum, base, remainder)
		encoded[i] = cv.base62Chars[remainder.Int64()]
	}

	return string(encoded), nil
}

// base62Decode converts a Base62 encoded string back into a hexadecimal private key
func (cv *CryptoVoucher) base62Decode(base62Input string) (string, error) {
	if len(base62Input) != encodedLength {
		return "", fmt.Errorf("base62 input must be %d characters long!", encodedLength)
	}

	decimal := big.NewInt(0)
	base := big.NewInt(62)
	digit := new(big.Int)

	// Convert the Base62 string back to a decimal number
	for i := 0; i < len(base62Input); i++ {
		index := strings.IndexByte(cv.base62Chars, base62Input[i])
		if index == -1 {
			return "", errors.New("invalid Base62 input. Only alphanumeric characters are allowed!")
		}
		decimal.Mul(decimal, base).Add(decimal, digit.SetInt64(int64(index)))
	}

	if !isInKeyRange(decimal) {
		return "", errors.New("private key is out of secp256k1 range!")
	}

	// Convert the decimal number back to a hexadecimal string
	hexOutput := fmt.Sprintf("%064x", decimal)
	return hexOutput, nil
}

// checkChar computes the Luhn mod 62 check character of a validated Base62 string
func (cv *CryptoVoucher) checkChar(base62Input string) byte {
	total, factor := 0, 2
	for i := len(base62Input) - 1; i >= 0; i-- {
		addend := factor * strings.IndexByte(cv.base62Chars, base62Input[i])
		total += addend/62 + addend%62
		factor = 3 - factor
	}
	return cv.base62Chars[(62-total%62)%62]
}

// CreateVoucher generates a voucher key and voucher code from a private key
func (cv *CryptoVoucher) CreateVoucher(privateKey string) (string, string, error) {
	// Convert the private key to a Base62 encoded string
	compressedKey, err := cv.base62Encode(privateKey)
	if err != nil {
		return "", "", err
	}

	// Append the check character and split into a voucher key and voucher code
	voucher := compressedKey + string(cv.checkChar(compressedKey))
	voucherKey := voucher[:voucherKeyLength]
	voucherCode := voucher[voucherKeyLength:]
	return voucherKey, voucherCode, nil
}

// RestorePrivateKey reconstructs a private key from a voucher key and voucher code
func (cv *CryptoVoucher) RestorePrivateKey(voucherKey, voucherCode string) (string, error) {
	// Combine the voucher key and voucher code to reconstruct the full voucher
	voucher := voucherKey + voucherCode
	if len(voucher) != voucherLength {
		return "", fmt.Errorf("voucher must be %d characters long!", voucherLength)
	}

	// Decode the Base62 string back to the original private key
	compressedKey := voucher[:encodedLength]
	decodedKey, err := cv.base62Decode(compressedKey)
	if err != nil {
		return "", err
	}

	if cv.checkChar(compressedKey) != voucher[encodedLength] {
		return "", errors.New("invalid voucher check character!")
	}
	return decodedKey, nil
}
