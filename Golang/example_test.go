package cryptovoucher_test

import (
	"fmt"
	"strings"

	cryptovoucher "github.com/EJ-Research/CryptoVoucher/Golang"
)

func Example() {
	// Example private key to encode and decode
	privateKey := "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"
	cryptoVoucher := cryptovoucher.NewCryptoVoucher()

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
	// Output:
	// Voucher Key: 2hvFlb6W2LlmFns9bG3NdCO6l85G
	// Voucher Code: VdTISeuq2iftf7zs
	// Restored Private Key: 0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B
}
