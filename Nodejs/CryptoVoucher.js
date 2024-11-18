// Author: (EJ)
// Description: This code provides functionalities for encoding private keys into vouchers 
// and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
// across multiple programming languages and platforms.
//
// License: MIT
// Feel free to use, modify, or distribute this code under the terms of the MIT License.



// CryptoVoucher class to handle Base62 encoding/decoding and voucher creation
class CryptoVoucher {
    constructor() {
        // Characters used for Base62 encoding
        this.base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    }

    // Validates if the input is a hexadecimal string of the specified length
    validateHex(input, length) {
        if (input.length !== length) {
            return { valid: false, message: `Input must be ${length} characters long!` };
        }
        if (!/^[0-9a-fA-F]+$/.test(input)) {
            return { valid: false, message: "Input must be a valid hexadecimal string!" };
        }
        return { valid: true };
    }

    // Encodes a hexadecimal string into a Base62 string
    base62Encode(hexInput) {
        const validation = this.validateHex(hexInput, 64);
        if (!validation.valid) {
            return { success: false, message: validation.message };
        }

        // Convert hex string to a decimal number
        let decimalValue = BigInt(`0x${hexInput}`);
        let base62 = "";

        // Convert decimal to Base62
        while (decimalValue > 0) {
            const remainder = Number(decimalValue % 62n);
            base62 = this.base62Chars[remainder] + base62;
            decimalValue = decimalValue / 62n;
        }

        return { success: true, data: base62 };
    }

    // Decodes a Base62 string back into a hexadecimal string
    base62Decode(base62Input) {
        if (!/^[0-9A-Za-z]+$/.test(base62Input)) {
            return { success: false, message: "Invalid Base62 input. Only alphanumeric characters are allowed!" };
        }

        let decimalValue = 0n;

        // Convert Base62 to decimal
        for (let char of base62Input) {
            const index = this.base62Chars.indexOf(char);
            if (index === -1) {
                return { success: false, message: "Invalid character in Base62 input!" };
            }
            decimalValue = decimalValue * 62n + BigInt(index);
        }

        // Convert decimal to hex string and pad to 64 characters
        const hexOutput = decimalValue.toString(16).padStart(64, "0");
        return { success: true, data: hexOutput };
    }

    // Creates a voucher key and voucher code from a private key
    createVoucher(privateKey) {
        const encoded = this.base62Encode(privateKey);
        if (!encoded.success) {
            return { success: false, message: encoded.message };
        }

        const base62Encoded = encoded.data;
        if (base62Encoded.length < 28) {
            return { success: false, message: "Encoded key is too short!" };
        }

        const voucherKey = base62Encoded.slice(0, 28);
        const voucherCode = base62Encoded.slice(28);
        return { success: true, voucherKey, voucherCode };
    }

    // Restores a private key from a voucher key and voucher code
    restorePrivateKey(voucherKey, voucherCode) {
        const combined = voucherKey + voucherCode;
        const decoded = this.base62Decode(combined);
        if (!decoded.success) {
            return { success: false, message: decoded.message };
        }

        return { success: true, data: decoded.data };
    }
}

// Example usage
const cryptoVoucher = new CryptoVoucher();
const privateKey = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B";

// Create a voucher
const voucher = cryptoVoucher.createVoucher(privateKey);
if (voucher.success) {
    console.log("Voucher Key:", voucher.voucherKey);
    console.log("Voucher Code:", voucher.voucherCode);

    // Restore the private key
    const restored = cryptoVoucher.restorePrivateKey(voucher.voucherKey, voucher.voucherCode);
    if (restored.success) {
        console.log("Restored Private Key:", restored.data.toUpperCase());
        console.log(privateKey.toUpperCase() === restored.data.toUpperCase() ? "Success!" : "Failed!");
    } else {
        console.error("Error restoring private key:", restored.message);
    }
} else {
    console.error("Error creating voucher:", voucher.message);
}