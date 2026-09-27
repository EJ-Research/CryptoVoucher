# CryptoVoucher

CryptoVoucher turns a wallet private key into a short voucher that can be sold, printed, sent in a chat and redeemed like a gift card. It is a small library for Go, Node.js, Python and PHP.

## The idea

For most people, buying crypto is still harder than it should be. They need a wallet app, a backup phrase, the right network and some idea of what a fee is. Many give up before the first transaction.

A voucher skips all of that. The seller creates a fresh wallet, funds it, and gives the buyer a 44 character code. The buyer never has to know a wallet exists. To spend it, they hand the code to a merchant, who turns it back into the private key, checks what is on the address and moves the funds to their own wallet.

The voucher is the private key written in a shorter alphabet. Nothing is encrypted and there is no server in between. Whoever holds the full voucher controls the funds, just like whoever holds a banknote can spend it. That is what keeps the system simple, and it is also why a voucher has to be redeemed in a specific order (see [Redeeming a voucher](#redeeming-a-voucher) and [The race window](#the-race-window)).

The whole flow:

1. The issuer generates a new private key and funds its address.
2. The issuer calls `create_voucher` and gives the voucher to the buyer.
3. The buyer passes the voucher to a merchant.
4. The merchant calls `restore_private_key`, derives the address and checks the balance.
5. The merchant moves the full balance to their own wallet, waits until that transfer is final, and only then delivers.

## Voucher format

A private key is a 256-bit number, normally written as 64 hex characters. To build a voucher, the library:

1. checks that the key is a valid secp256k1 private key (`1 <= key < n`),
2. writes the number in Base62 (`0-9`, `A-Z`, `a-z`) and left-pads it with `0` to 43 characters,
3. appends one Luhn mod 62 check character,
4. splits the 44 characters into a voucher key and a voucher code.

| Part | Length | Content |
|---|---|---|
| Voucher key | 28 | First 28 Base62 characters |
| Voucher code | 16 | Last 15 Base62 characters, then the check character |

43 is the shortest Base62 length that fits every 256-bit value (62^43 is just above 2^256), so the voucher cannot get any shorter without losing information. The padding zeros have no effect on the value, and the decoder treats them like any other digit.

The check character catches every single mistyped character and almost every swap of two neighbouring characters (the only swap it misses is `0` with `z`). A voucher that fails the check is rejected instead of silently decoding to some other valid key.

The two parts follow the usual gift card layout of a card number plus a PIN. Both parts are needed to restore the key, and both have to stay secret, since the voucher key alone reveals most of the private key.

Vouchers are case sensitive: `a` and `A` are different characters.

## Sample data

This key is public. Do not send funds to its addresses.

| Field | Value |
|---|---|
| Private key | `0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B` |
| Voucher key | `2hvFlb6W2LlmFns9bG3NdCO6l85G` |
| Voucher code | `VdTISeuq2iftf7zY` |
| TRON address | `THpApTFkvxbvHThDKix1mwe7KDYxfVtRhP` |
| EVM address (Ethereum, BSC, Polygon) | `0x560b70C1F4Bd994037911858E29281909EcAAA14` |

Edge cases, handy for testing a port:

| Private key | Voucher key | Voucher code |
|---|---|---|
| `0000000000000000000000000000000000000000000000000000000000000001` | `0000000000000000000000000000` | `000000000000001y` |
| `fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140` (n - 1) | `yhjskwdA6OZ1AL1YmHWZWcETkvPc` | `IqwQly7v5TWNN68m` |

Inputs that must be rejected (messages from the Python version):

| Input | Result |
|---|---|
| `2hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISXuq2iftf7zY` (one character changed) | `Invalid voucher check character!` |
| `2hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7z` (old format, no check character) | `Voucher must be 44 characters long!` |
| Private key `000...000` | `Private key is out of secp256k1 range!` |

## Installation

**Go** 1.16 or newer:

```sh
go get github.com/EJ-Research/CryptoVoucher/Golang
```

**Node.js** 14 or newer. The package is not on npm yet, so install it from a checkout:

```sh
npm install ./CryptoVoucher/Nodejs
```

**Python** 3.8 or newer:

```sh
pip install "git+https://github.com/EJ-Research/CryptoVoucher.git#subdirectory=Python"
```

**PHP** 7.4 or newer with the BCMath extension. Add the repository to `composer.json`:

```json
{
    "repositories": [{ "type": "vcs", "url": "https://github.com/EJ-Research/CryptoVoucher" }],
    "require": { "ej-research/cryptovoucher": "dev-main" }
}
```

The Elixir and Ruby versions still exist but are deprecated, see [below](#elixir-and-ruby-are-deprecated).

## Usage

All four versions expose the same two operations. Private keys are accepted in upper or lower case, and restored keys always come back as 64 lower case hex characters.

**Go**

```go
import cryptovoucher "github.com/EJ-Research/CryptoVoucher/Golang"

cv := cryptovoucher.NewCryptoVoucher()

voucherKey, voucherCode, err := cv.CreateVoucher(privateKey)
if err != nil {
	// invalid private key
}

restoredKey, err := cv.RestorePrivateKey(voucherKey, voucherCode)
if err != nil {
	// wrong length, bad character, failed check or out of range
}
```

**Node.js**

```js
const { CryptoVoucher } = require("crypto-voucher");

const cv = new CryptoVoucher();

const voucher = cv.createVoucher(privateKey);
// { success: true, voucherKey: "...", voucherCode: "..." }
// { success: false, message: "..." }

const restored = cv.restorePrivateKey(voucher.voucherKey, voucher.voucherCode);
// { success: true, data: "0b6bf6..." }
```

**Python**

```python
from CryptoVoucher import CryptoVoucher

cv = CryptoVoucher()

voucher_key, voucher_code, error = cv.create_voucher(private_key)
restored_key, error = cv.restore_private_key(voucher_key, voucher_code)
```

`error` is `None` on success and a message otherwise.

**PHP**

```php
require 'vendor/autoload.php';

$cv = new CryptoVoucher();

$voucher = $cv->createVoucher($privateKey);
// ['status' => 'success', 'voucher_key' => '...', 'voucher_code' => '...']

$restored = $cv->restorePrivateKey($voucher['voucher_key'], $voucher['voucher_code']);
// ['status' => 'success', 'data' => '0b6bf6...']
// ['status' => 'error', 'message' => '...']
```

Running the Node.js, Python or PHP file directly prints the sample voucher above, and `go test` runs the same example in Go. Importing the library has no side effects.

## Redeeming a voucher

### 1. Decode it

Call `restore_private_key` with both parts. It fails when:

- the combined length is not 44,
- a character is outside `0-9A-Za-z`,
- the check character does not match, which almost always means a typo,
- the decoded number is not a valid secp256k1 key.

On failure, ask the customer to type the voucher again. Never try to "fix" it by guessing characters.

### 2. Get the address

A single private key maps to a different address on each network family. Any secp256k1 wallet library can derive them. In Node.js:

```js
const { Wallet } = require("ethers");      // ethers v6
const { TronWeb } = require("tronweb");    // tronweb v6

const evmAddress = new Wallet("0x" + privateKey).address;      // Ethereum, BSC, Polygon, ...
const tronAddress = TronWeb.address.fromPrivateKey(privateKey); // TRON
```

In Python:

```python
from eth_account import Account
from tronpy.keys import PrivateKey

evm_address = Account.from_key("0x" + private_key).address
tron_address = PrivateKey(bytes.fromhex(private_key)).public_key.to_base58check_address()
```

A TRON address holds the same 20 bytes as the EVM address, prefixed with `0x41` and written in Base58Check. For the sample key, `0x560b70C1...EcAAA14` on Ethereum is `THpApTFk...fVtRhP` on TRON.

Bitcoin can use the same key, but a single key has several address types (legacy, SegWit, Taproot). The issuer has to tell the merchant which one was funded.

### 3. Check the balance

Always read the balance from confirmed (or finalized) state rather than from the latest block. A balance that only exists in a pending block can still disappear.

Common token contracts:

| Network | Token | Contract | Decimals |
|---|---|---|---|
| TRON | USDT (TRC-20) | `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` | 6 |
| Ethereum | USDT (ERC-20) | `0xdAC17F958D2ee523a2206206994597C13D831ec7` | 6 |
| BSC | USDT (BEP-20) | `0x55d398326f99059fF775485246999027B3197955` | 18 |
| Polygon | USDT | `0xc2132D05D31c914a87C6611C10748AEb04B58e8F` | 6 |

Double check these against the token issuer's own documentation before going live. Note that USDT on BSC uses 18 decimals, not 6.

#### TRON

TronGrid returns TRX and TRC-20 balances in one request:

```sh
curl -s "https://api.trongrid.io/v1/accounts/THpApTFkvxbvHThDKix1mwe7KDYxfVtRhP?only_confirmed=true"
```

`data[0].balance` is the TRX balance in sun (1 TRX = 1,000,000 sun), and `data[0].trc20` lists token balances as `{ "<contract>": "<amount>" }`. An empty `data` array means the address has never been activated, so there is nothing on it. For production traffic, get a TronGrid API key and send it in the `TRON-PRO-API-KEY` header.

With tronweb:

```js
const tronWeb = new TronWeb({ fullHost: "https://api.trongrid.io", privateKey });

const sun = await tronWeb.trx.getBalance(tronAddress); // confirmed balance
const usdt = await tronWeb.contract().at("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t");
const units = await usdt.balanceOf(tronAddress).call({ confirmed: true }); // 6 decimals
```

`trx.getBalance` already reads from the solidity node. For contract calls, `{ confirmed: true }` does the same; without it, the call reads the latest block.

#### Ethereum, BSC and other EVM chains

Any JSON-RPC endpoint works. Use the `finalized` block tag:

```sh
# native coin (ETH, BNB, POL), result in wei as hex
curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0x560b70C1F4Bd994037911858E29281909EcAAA14","finalized"]}'

# token balance: balanceOf(address), selector 0x70a08231 + address padded to 32 bytes
curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x70a08231000000000000000000000000560b70c1f4bd994037911858e29281909ecaaa14"},"finalized"]}'
```

With ethers:

```js
const { JsonRpcProvider, Contract } = require("ethers");

const provider = new JsonRpcProvider(RPC_URL);
const wei = await provider.getBalance(evmAddress, "finalized");

const usdt = new Contract(USDT_ADDRESS, ["function balanceOf(address) view returns (uint256)"], provider);
const units = await usdt.balanceOf(evmAddress, { blockTag: "finalized" });
```

### 4. Move the funds before you deliver

A balance check proves nothing on its own; the funds are yours only after they reach your wallet. So after checking the balance, send everything to your own address right away, wait for that transfer to become final, and only then hand over the product. The next section explains why.

How the transfer is paid for depends on what the voucher holds.

**Native coin (TRX, ETH, BNB).** The fee is paid from the voucher balance itself.

- On EVM chains a plain transfer uses 21,000 gas. Send `balance - 21000 * maxFeePerGas`. The real fee is usually a bit lower, and the difference stays on the voucher address as dust.
- On TRON a TRX transfer uses bandwidth. It is normally covered by the free daily bandwidth every activated account gets; if not, a small amount of TRX is burned.

**Tokens (USDT and similar).** The voucher address needs some native coin to pay the fee, otherwise the token transfer cannot be sent at all.

- On TRON, a USDT transfer needs energy. Without staked energy, the network burns TRX to cover it. The cost depends on the current energy price and on whether the receiving address has ever held USDT (the first transfer to a new address costs about twice as much). Set `feeLimit` high enough; a transfer that runs out of energy fails and still burns its fee.
- On Ethereum, BSC and Polygon, a token transfer needs ETH, BNB or POL for gas.

The easiest setup is for the issuer to add enough native coin when charging a token voucher. On TRON, sending TRX first also activates the new address. If a voucher arrives without gas, the merchant has to send a small amount of native coin to the voucher address, wait for it to confirm, and then move the tokens right away. Any gas sent this way is just as exposed as the tokens themselves, so send only what the transfer needs.

When the transfer is done, anything left on the voucher address (dust, unused gas) can be swept the same way or ignored.

## The race window

Every person or system that has seen the voucher can spend it:

- the issuer's system, which generated the key,
- the buyer,
- anyone who saw the voucher (a photo, a chat message, a receipt left on a counter),
- any other merchant the voucher was shown to.

The race window is the time between your balance check and the moment your sweep becomes final. If any other holder moves the funds during that time, your sweep fails or confirms with nothing to send, even though the balance looked fine a moment earlier.

Pending transactions do not close the window:

- On EVM chains, another holder can broadcast a transaction with the same nonce and a higher fee and replace yours before it is mined.
- On Bitcoin, replace-by-fee does the same thing.
- On TRON there is no fee bidding, but if two transactions spend the same balance, whichever lands in a block first wins.

The library cannot close this window. That is simply how a bearer key works. What you can do is keep the window short and wait for finality before delivering:

1. Read the balance from confirmed or finalized state.
2. Sweep right away, not in a batch job later.
3. Deliver only after the sweep is final:

   | Network | When to treat the sweep as final |
   |---|---|
   | TRON | After 19 blocks (about one minute), when the block is solidified |
   | Ethereum | When the block is finalized, about 13 minutes |
   | BSC | When the block is finalized, a few seconds with fast finality |
   | Bitcoin | 1 to 6 confirmations, depending on the amount |

4. Treat a failed or empty sweep as a voucher that has already been used.
5. Record the voucher address together with the sweep transaction ID. If there is a dispute, the chain is the source of truth.

## Elixir and Ruby are deprecated

The Elixir and Ruby versions are no longer maintained. They produce exactly the same vouchers as the other four versions today, so existing users are not affected, but they will not receive fixes or format changes. New projects should use Go, Node.js, Python or PHP. The Ruby version prints a warning when it is loaded, and the Elixir functions are marked with `@deprecated`.

The reasons are practical. Go, Node.js, Python and PHP cover almost every place this library actually ends up: payment backends, shops, bots and internal tools. Keeping six copies of the same logic in sync costs more than it returns, because every change to the format has to be written, tested and released six times, and the versions tend to drift apart. The whole library is a single class of under 200 lines and the format is fully described above, so porting it to another language is a small job today, and an AI coding assistant can do most of it. Use the sample data and edge cases above to check the result.

## Notes

- The library has been tested on TRON. Ethereum, BSC and Bitcoin use the same key type, but test on your own setup before accepting real vouchers.
- Only secp256k1 keys are supported.
- Vouchers created before the check character was added (40 to 43 characters) are rejected by the current version.

## Contributing

Issues and pull requests are welcome. Any change to the voucher format has to land in all four maintained versions (Go, Node.js, Python and PHP) and has to keep the sample data above valid.

## License

[MIT](LISENCE)
