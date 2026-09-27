# cryptovoucher

Turns a private key and its network into a 45-character voucher and back. This is the Python version of [CryptoVoucher](https://github.com/EJ-Research/CryptoVoucher); the voucher format, the network code table and the guide for redeeming vouchers are in the main README.

## Install

```sh
pip install cryptovoucher
```

Python 3.8 or newer, no dependencies.

## Use

```python
from CryptoVoucher import CryptoVoucher

cv = CryptoVoucher()

voucher_key, voucher_code, error = cv.create_voucher(private_key, "TRON_USDT")
restored_key, network, error = cv.restore_private_key(voucher_key, voucher_code)
```

`error` is `None` on success and a message otherwise. Whoever holds a voucher controls its funds, so read the redeem guide in the main README before accepting vouchers.

## License

MIT
