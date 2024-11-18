# Crypto Voucher Library

## Overview

The **Crypto Voucher Library** is a multi-language implementation designed to simplify the use of cryptocurrency for everyday users. The project is implemented in **Elixir**, **Go**, **Node.js**, **PHP**, **Python**, and **Ruby**, making it accessible to developers working with diverse platforms and tools

### Purpose

One of the significant barriers to adopting cryptocurrency is the complexity of managing wallets, securing private keys, and understanding how to use them, This library addresses that challenge by introducing a simple voucher system:

1. **For Buyers:**
   - A user buys cryptocurrency worth a specific amount (e.g., $10).
   - The system generates a wallet, transfers the cryptocurrency into it, and creates a voucher based on the private key of the wallet!
   - The buyer receives the voucher

2. **For Sellers (Merchants):**
   - A merchant accepting the voucher can use this library to decode the voucher back into the wallet's private key!
   - The merchant checks the wallet's balance and processes the transaction
   - Once verified, the product or service is provided to the buyer

This system abstracts away the complexity of managing wallets while ensuring security and usability for general users

---

## Features

- **Multi-language Implementation:** Support for Elixir, Go, Node.js, PHP, Python, and Ruby
- **Base62 Encoding/Decoding:** Efficient encoding of private keys into vouchers and back
- **Simple Integration:** Easy-to-use APIs for integrating with any application
- **Security-focused Design:** Keeps private keys secure while ensuring straightforward usage

---

## How It Works

### 1. Creating a Voucher:

- A user initiates a purchase (e.g., $10 worth of cryptocurrency)
- The system:
  1. Creates a wallet for the user
  2. Transfers the cryptocurrency into the wallet
  3. Encodes the wallet's private key into a voucher (key + code)
- The user receives the voucher, which can be used for transactions

### 2. Redeeming a Voucher:

- The merchant:
  1. Decodes the voucher back into the private key using this library
  2. Checks the wallet's balance
  3. Processes the transaction and provides the service/product to the buyer

---

## Installation

Each language implementation has its specific installation and setup instructions! Below is a summary for each:

### Elixir
- Add the module to your project
- Ensure you have **Elixir 1.12 or higher** installed

### Go
- Use `go get` to add the package.
- Requires **Go 1.16 or higher**

### Node.js
- Requires **Node.js 14 or higher**

### PHP
- Include the library in your project manually
- Requires **PHP 7.4 or higher** with the BCMath extension

### Python
- Requires **Python 3.8 or higher**

### Ruby
- Requires **Ruby 2.7 or higher**

---

## Usage

### Example Workflow

#### Buyer:
1. Call the `create_voucher` function to generate a voucher for a new wallet
2. Receive the `voucher_key` and `voucher_code`

#### Merchant:
1. Use the `restore_private_key` function to decode the voucher into the wallet's private key
2. Check the balance and process the transaction


### Inputs and Outputs
- **Input:** A 64-character hexadecimal private key
- **Output:**
  - `voucher_key` (28 characters)
  - `voucher_code` (remaining Base62 encoded characters)

---



## Notes

This library has been tested with the TRON blockchain and works seamlessly! If you intend to use it with other blockchain networks, ensure thorough testing to verify compatibility and correctness



## Contributing

Contributions are welcome! Please follow the steps below:

1. Fork the repository
2. Implement your changes
3. Submit a pull request with a detailed description of the changes

---

## License

This project is licensed under the **MIT License** See the `LICENSE` file for details

---

## Contact

For questions, issues, or suggestions, feel free to open an issue on GitHub or contact the author via email

