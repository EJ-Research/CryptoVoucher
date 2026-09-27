# crypto-voucher

Turns a private key and its network into a 45-character voucher and back. This is the Node.js version of [CryptoVoucher](https://github.com/EJ-Research/CryptoVoucher); the voucher format, the network code table and the guide for redeeming vouchers are in the main README.

## Install

```sh
npm install crypto-voucher
```

Node.js 14 or newer, no dependencies.

## Use

```js
const { CryptoVoucher } = require("crypto-voucher");

const cv = new CryptoVoucher();

const voucher = cv.createVoucher(privateKey, "TRON_USDT");
// { success: true, voucherKey: "...", voucherCode: "..." }
// { success: false, message: "..." }

const restored = cv.restorePrivateKey(voucher.voucherKey, voucher.voucherCode);
// { success: true, data: "0b6bf6...", network: "TRON_USDT" }
```

Whoever holds a voucher controls its funds, so read the redeem guide in the main README before accepting vouchers.

## License

MIT
