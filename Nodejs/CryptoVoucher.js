// Author: (EJ)
// Description: This code provides functionalities for encoding private keys into vouchers 
// and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
// across multiple programming languages and platforms.
//
// License: MIT
// Feel free to use, modify, or distribute this code under the terms of the MIT License.



// secp256k1 curve order; valid private keys are in [1, N-1]
const SECP256K1_N = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141n;
// 62^43 > 2^256, so every private key fits in 43 Base62 characters
const ENCODED_LENGTH = 43;
// Network code followed by the first 28 Base62 characters
const VOUCHER_KEY_LENGTH = 29;
// Remaining 15 Base62 characters followed by one check character
const VOUCHER_CODE_LENGTH = 16;

// Network code (first character of every voucher) -> network ID.
// Codes are permanent: an assigned code is never changed or reused, new networks are only appended.
const NETWORKS = new Map([
    ["1", "BITCOIN_BTC"],
    ["2", "ETHEREUM_ETH"],
    ["3", "ETHEREUM_USDT"],
    ["4", "ETHEREUM_USDC"],
    ["5", "TRON_TRX"],
    ["6", "TRON_USDT"],
    ["7", "BSC_BNB"],
    ["8", "BSC_USDT"],
    ["9", "BSC_USDC"],
    ["A", "POLYGON_POL"],
    ["B", "POLYGON_USDT"],
    ["C", "POLYGON_USDC"],
    ["D", "SOLANA_SOL"],
    ["E", "SOLANA_USDT"],
    ["F", "SOLANA_USDC"],
    ["G", "TON_TON"],
    ["H", "TON_USDT"],
    ["J", "ARBITRUM_ETH"],
    ["K", "ARBITRUM_USDT"],
    ["L", "ARBITRUM_USDC"],
    ["M", "OPTIMISM_ETH"],
    ["N", "OPTIMISM_USDT"],
    ["P", "OPTIMISM_USDC"],
    ["Q", "BASE_ETH"],
    ["R", "BASE_USDC"],
    ["S", "AVALANCHE_AVAX"],
    ["T", "AVALANCHE_USDT"],
    ["U", "AVALANCHE_USDC"],
    ["V", "LITECOIN_LTC"],
    ["W", "DOGECOIN_DOGE"],
    ["X", "BITCOINCASH_BCH"],
    ["Y", "XRPL_XRP"],
    ["a", "ETHEREUM_DAI"],
    ["b", "ETHEREUM_PYUSD"],
]);
const NETWORK_CODES = new Map([...NETWORKS].map(([code, network]) => [network, code]));

// CryptoVoucher class to handle Base62 encoding/decoding and voucher creation
class CryptoVoucher {
    constructor() {
        // Characters used for Base62 encoding
        this.base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    }

    // Validates if the input is a hexadecimal string of the specified length
    validateHex(input, length) {
        if (typeof input !== "string") {
            return { valid: false, message: "Input must be a string!" };
        }
        if (input.length !== length) {
            return { valid: false, message: `Input must be ${length} characters long!` };
        }
        if (!/^[0-9a-fA-F]+$/.test(input)) {
            return { valid: false, message: "Input must be a valid hexadecimal string!" };
        }
        return { valid: true };
    }

    // Encodes a hexadecimal private key into a fixed-length Base62 string
    base62Encode(hexInput) {
        const validation = this.validateHex(hexInput, 64);
        if (!validation.valid) {
            return { success: false, message: validation.message };
        }

        // Convert hex string to a decimal number
        let decimalValue = BigInt(`0x${hexInput}`);
        if (decimalValue <= 0n || decimalValue >= SECP256K1_N) {
            return { success: false, message: "Private key is out of secp256k1 range!" };
        }

        let base62 = "";

        // Convert decimal to Base62
        while (decimalValue > 0) {
            const remainder = Number(decimalValue % 62n);
            base62 = this.base62Chars[remainder] + base62;
            decimalValue = decimalValue / 62n;
        }

        // Leading zeros keep the length fixed and do not change the decoded value
        return { success: true, data: base62.padStart(ENCODED_LENGTH, "0") };
    }

    // Decodes a Base62 string back into a hexadecimal private key
    base62Decode(base62Input) {
        if (typeof base62Input !== "string") {
            return { success: false, message: "Base62 input must be a string!" };
        }
        if (base62Input.length !== ENCODED_LENGTH) {
            return { success: false, message: `Base62 input must be ${ENCODED_LENGTH} characters long!` };
        }
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

        if (decimalValue <= 0n || decimalValue >= SECP256K1_N) {
            return { success: false, message: "Private key is out of secp256k1 range!" };
        }

        // Convert decimal to hex string and pad to 64 characters
        const hexOutput = decimalValue.toString(16).padStart(64, "0");
        return { success: true, data: hexOutput };
    }

    // Computes the check character of a validated Base62 string: the sum of each digit
    // times its 1-based position, modulo 61 (prime, so every position weight is
    // invertible and any single swap changes the sum)
    checkChar(base62Input) {
        let total = 0;
        for (let i = 0; i < base62Input.length; i++) {
            total += (i + 1) * this.base62Chars.indexOf(base62Input[i]);
        }
        return this.base62Chars[total % 61];
    }

    // Creates a voucher key and voucher code from a private key and its network ID
    createVoucher(privateKey, network) {
        const code = NETWORK_CODES.get(network);
        if (code === undefined) {
            return { success: false, message: "Unknown network!" };
        }

        const encoded = this.base62Encode(privateKey);
        if (!encoded.success) {
            return { success: false, message: encoded.message };
        }

        let voucher = code + encoded.data;
        voucher += this.checkChar(voucher);
        const voucherKey = voucher.slice(0, VOUCHER_KEY_LENGTH);
        const voucherCode = voucher.slice(VOUCHER_KEY_LENGTH);
        return { success: true, voucherKey, voucherCode };
    }

    // Restores a private key and its network ID from a voucher key and voucher code
    restorePrivateKey(voucherKey, voucherCode) {
        if (typeof voucherKey !== "string" || typeof voucherCode !== "string") {
            return { success: false, message: "Voucher key and voucher code must be strings!" };
        }
        // Checking each part also catches the two parts entered in swapped order
        if (voucherKey.length !== VOUCHER_KEY_LENGTH) {
            return { success: false, message: `Voucher key must be ${VOUCHER_KEY_LENGTH} characters long!` };
        }
        if (voucherCode.length !== VOUCHER_CODE_LENGTH) {
            return { success: false, message: `Voucher code must be ${VOUCHER_CODE_LENGTH} characters long!` };
        }

        const voucher = voucherKey + voucherCode;
        const network = NETWORKS.get(voucher[0]);
        if (network === undefined) {
            return { success: false, message: "Unknown network code!" };
        }

        const decoded = this.base62Decode(voucher.slice(1, ENCODED_LENGTH + 1));
        if (!decoded.success) {
            return { success: false, message: decoded.message };
        }

        // The check character covers the network code as well
        if (this.checkChar(voucher.slice(0, ENCODED_LENGTH + 1)) !== voucher[ENCODED_LENGTH + 1]) {
            return { success: false, message: "Invalid voucher check character!" };
        }

        return { success: true, data: decoded.data, network };
    }
}

module.exports = { CryptoVoucher };

// Example usage (runs only when this file is executed directly)
if (require.main === module) {
    const cryptoVoucher = new CryptoVoucher();
    const privateKey = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B";

    // Create a voucher for USDT on TRON
    const voucher = cryptoVoucher.createVoucher(privateKey, "TRON_USDT");
    if (voucher.success) {
        console.log("Voucher Key:", voucher.voucherKey);
        console.log("Voucher Code:", voucher.voucherCode);

        // Restore the private key and network
        const restored = cryptoVoucher.restorePrivateKey(voucher.voucherKey, voucher.voucherCode);
        if (restored.success) {
            console.log("Restored Private Key:", restored.data.toUpperCase());
            console.log("Network:", restored.network);
            console.log(privateKey.toUpperCase() === restored.data.toUpperCase() ? "Success!" : "Failed!");
        } else {
            console.error("Error restoring private key:", restored.message);
        }
    } else {
        console.error("Error creating voucher:", voucher.message);
    }
}
