package cryptovoucher_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	cryptovoucher "github.com/EJ-Research/CryptoVoucher/Golang"
)

type vectorFile struct {
	Valid []struct {
		PrivateKey  string `json:"private_key"`
		Network     string `json:"network"`
		VoucherKey  string `json:"voucher_key"`
		VoucherCode string `json:"voucher_code"`
		RestoredKey string `json:"restored_key"`
	} `json:"valid"`
	CreateErrors []struct {
		PrivateKey string `json:"private_key"`
		Network    string `json:"network"`
		Error      string `json:"error"`
	} `json:"create_errors"`
	RestoreErrors []struct {
		VoucherKey  string `json:"voucher_key"`
		VoucherCode string `json:"voucher_code"`
		Error       string `json:"error"`
		Note        string `json:"note"`
	} `json:"restore_errors"`
}

func TestVectors(t *testing.T) {
	data, err := os.ReadFile("../vectors.json")
	if os.IsNotExist(err) {
		// vectors.json sits outside the Go module, so it is missing from downloaded copies
		t.Skip("../vectors.json not found")
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors vectorFile
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}

	cv := cryptovoucher.NewCryptoVoucher()
	for _, v := range vectors.Valid {
		key, code, err := cv.CreateVoucher(v.PrivateKey, v.Network)
		if err != nil || key != v.VoucherKey || code != v.VoucherCode {
			t.Errorf("CreateVoucher(%q, %q) = %q, %q, %v; want %q, %q", v.PrivateKey, v.Network, key, code, err, v.VoucherKey, v.VoucherCode)
		}
		restored, network, err := cv.RestorePrivateKey(v.VoucherKey, v.VoucherCode)
		if err != nil || restored != v.RestoredKey || network != v.Network {
			t.Errorf("RestorePrivateKey(%q, %q) = %q, %q, %v; want %q, %q", v.VoucherKey, v.VoucherCode, restored, network, err, v.RestoredKey, v.Network)
		}
	}
	for _, v := range vectors.CreateErrors {
		if _, _, err := cv.CreateVoucher(v.PrivateKey, v.Network); err == nil || !strings.EqualFold(err.Error(), v.Error) {
			t.Errorf("CreateVoucher(%q, %q) error = %v; want %q", v.PrivateKey, v.Network, err, v.Error)
		}
	}
	for _, v := range vectors.RestoreErrors {
		if _, _, err := cv.RestorePrivateKey(v.VoucherKey, v.VoucherCode); err == nil || !strings.EqualFold(err.Error(), v.Error) {
			t.Errorf("%s: RestorePrivateKey(%q, %q) error = %v; want %q", v.Note, v.VoucherKey, v.VoucherCode, err, v.Error)
		}
	}
}
