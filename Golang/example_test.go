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

	// Create a voucher for USDT on TRON
	voucherKey, voucherCode, err := cryptoVoucher.CreateVoucher(privateKey, "TRON_USDT")
	if err != nil {
		fmt.Println("Error creating voucher:", err)
		return
	}

	fmt.Println("Voucher Key:", voucherKey)
	fmt.Println("Voucher Code:", voucherCode)

	// Restore the private key and network from the voucher
	restoredKey, network, err := cryptoVoucher.RestorePrivateKey(voucherKey, voucherCode)
	if err != nil {
		fmt.Println("Error restoring private key:", err)
		return
	}

	fmt.Println("Restored Private Key:", strings.ToUpper(restoredKey))
	fmt.Println("Network:", network)
	// Output:
	// Voucher Key: 62hvFlb6W2LlmFns9bG3NdCO6l85G
	// Voucher Code: VdTISeuq2iftf7zh
	// Restored Private Key: 0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B
	// Network: TRON_USDT
}
