// Package cryptovoucher encodes private keys into vouchers and restores them.
//
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
	// Characters used for Base62 encoding
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// 62^43 > 2^256, so every private key fits in 43 Base62 characters
	encodedLength = 43
	// Network code followed by the first 28 Base62 characters
	voucherKeyLength = 29
	// Remaining 15 Base62 characters followed by one check character
	voucherCodeLength = 16
)

// networks maps each network code (the first character of every voucher) to its network ID.
// Codes are permanent: an assigned code is never changed or reused, new networks are only appended.
var networks = map[byte]string{
	'1': "BITCOIN_BTC",
	'2': "ETHEREUM_ETH",
	'3': "ETHEREUM_USDT",
	'4': "ETHEREUM_USDC",
	'5': "TRON_TRX",
	'6': "TRON_USDT",
	'7': "BSC_BNB",
	'8': "BSC_USDT",
	'9': "BSC_USDC",
	'A': "POLYGON_POL",
	'B': "POLYGON_USDT",
	'C': "POLYGON_USDC",
	'D': "SOLANA_SOL",
	'E': "SOLANA_USDT",
	'F': "SOLANA_USDC",
	'G': "TON_TON",
	'H': "TON_USDT",
	'J': "ARBITRUM_ETH",
	'K': "ARBITRUM_USDT",
	'L': "ARBITRUM_USDC",
	'M': "OPTIMISM_ETH",
	'N': "OPTIMISM_USDT",
	'P': "OPTIMISM_USDC",
	'Q': "BASE_ETH",
	'R': "BASE_USDC",
	'S': "AVALANCHE_AVAX",
	'T': "AVALANCHE_USDT",
	'U': "AVALANCHE_USDC",
	'V': "LITECOIN_LTC",
	'W': "DOGECOIN_DOGE",
	'X': "BITCOINCASH_BCH",
	'Y': "XRPL_XRP",
	'a': "ETHEREUM_DAI",
	'b': "ETHEREUM_PYUSD",
}

// networkCodes maps each network ID back to its network code
var networkCodes = func() map[string]byte {
	codes := make(map[string]byte, len(networks))
	for code, network := range networks {
		codes[network] = code
	}
	return codes
}()

// secp256k1 curve order; valid private keys are in [1, N-1]
var secp256k1N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

// CryptoVoucher provides methods for encoding private keys into vouchers
// and restoring private keys from vouchers using Base62 encoding.
// The zero value is ready to use.
type CryptoVoucher struct{}

// NewCryptoVoucher initializes a new CryptoVoucher instance
func NewCryptoVoucher() *CryptoVoucher {
	return &CryptoVoucher{}
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
		encoded[i] = base62Chars[remainder.Int64()]
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
		index := strings.IndexByte(base62Chars, base62Input[i])
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

// checkChar computes the check character of a validated Base62 string: the sum of
// each digit times its 1-based position, modulo 61 (prime, so every position
// weight is invertible and any single swap changes the sum)
func (cv *CryptoVoucher) checkChar(base62Input string) byte {
	total := 0
	for i := 0; i < len(base62Input); i++ {
		total += (i + 1) * strings.IndexByte(base62Chars, base62Input[i])
	}
	return base62Chars[total%61]
}

// CreateVoucher generates a voucher key and voucher code from a private key and its network ID
func (cv *CryptoVoucher) CreateVoucher(privateKey, network string) (string, string, error) {
	code, ok := networkCodes[network]
	if !ok {
		return "", "", errors.New("unknown network!")
	}

	// Convert the private key to a Base62 encoded string
	compressedKey, err := cv.base62Encode(privateKey)
	if err != nil {
		return "", "", err
	}

	// Prefix the network code, append the check character and split into a voucher key and voucher code
	voucher := string(code) + compressedKey
	voucher += string(cv.checkChar(voucher))
	voucherKey := voucher[:voucherKeyLength]
	voucherCode := voucher[voucherKeyLength:]
	return voucherKey, voucherCode, nil
}

// RestorePrivateKey reconstructs a private key and its network ID from a voucher key and voucher code
func (cv *CryptoVoucher) RestorePrivateKey(voucherKey, voucherCode string) (string, string, error) {
	// Checking each part also catches the two parts entered in swapped order
	if len(voucherKey) != voucherKeyLength {
		return "", "", fmt.Errorf("voucher key must be %d characters long!", voucherKeyLength)
	}
	if len(voucherCode) != voucherCodeLength {
		return "", "", fmt.Errorf("voucher code must be %d characters long!", voucherCodeLength)
	}

	// Combine the voucher key and voucher code to reconstruct the full voucher
	voucher := voucherKey + voucherCode
	network, ok := networks[voucher[0]]
	if !ok {
		return "", "", errors.New("unknown network code!")
	}

	// Decode the Base62 string back to the original private key
	decodedKey, err := cv.base62Decode(voucher[1 : encodedLength+1])
	if err != nil {
		return "", "", err
	}

	// The check character covers the network code as well
	if cv.checkChar(voucher[:encodedLength+1]) != voucher[encodedLength+1] {
		return "", "", errors.New("invalid voucher check character!")
	}
	return decodedKey, network, nil
}
