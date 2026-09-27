# CryptoVoucher

CryptoVoucher turns a wallet private key into a short voucher that can be sold, printed or sent as text, and redeemed like a gift card. It is a small library for Go, Node.js, Python and PHP.

## The idea

Getting someone started with crypto usually means walking them through a wallet app, a backup phrase, networks and fees. A voucher removes that step. The issuer creates a fresh wallet, funds it and gives the buyer a 44-character code. The buyer does not need a wallet at all. To spend the voucher, they give the code to a merchant, who turns it back into the private key, checks the balance of the address and moves the funds to their own wallet.

The voucher is the private key itself, written in a shorter alphabet. Nothing is encrypted and no server is involved. Anyone who holds the full voucher controls the funds, the same way anyone holding a banknote can spend it. This is why redeeming has to follow a fixed order (see [Redeeming a voucher](#redeeming-a-voucher) and [The race window](#the-race-window)).

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
3. appends one check character,
4. splits the 44 characters into a voucher key and a voucher code.

| Part | Length | Content |
|---|---|---|
| Voucher key | 28 | First 28 Base62 characters |
| Voucher code | 16 | Last 15 Base62 characters, then the check character |

43 is the shortest Base62 length that fits every 256-bit value (62^43 is just above 2^256), so the voucher cannot get any shorter without losing information. The padding zeros have no effect on the value, and the decoder treats them like any other digit.

The check character works like the ISBN-10 check digit. Each of the 43 Base62 digits is multiplied by its position (1 to 43), the products are added up, and the sum modulo 61 picks the check character (`0` to `y`, so `z` never appears in that position). Because 61 is prime, this catches:

- every mistyped character, except `0` typed as `z` or the other way round,
- every swap of two characters, whether they are next to each other or further apart, again except `0` with `z`.

Any other kind of damage gets through about once in 61 tries. `restore_private_key` also checks the length of each part, so entering the voucher key and the voucher code in each other's fields is always caught. A voucher that fails any of these checks is rejected instead of silently decoding to some other valid key.

Both parts are needed to restore the key, and both must be kept secret. Do not print or display the voucher key as if it were a public card number. It holds about two thirds of the private key, and once the voucher address has sent any transaction (which puts its public key on chain), the voucher key alone is enough to recover the rest with modest hardware.

Vouchers are case sensitive: `a` and `A` are different characters.

## Sample data

This key is public. Do not send funds to its addresses.

| Field | Value |
|---|---|
| Private key | `0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B` |
| Voucher key | `2hvFlb6W2LlmFns9bG3NdCO6l85G` |
| Voucher code | `VdTISeuq2iftf7zs` |
| TRON address | `THpApTFkvxbvHThDKix1mwe7KDYxfVtRhP` |
| EVM address (Ethereum, BSC, Polygon) | `0x560b70C1F4Bd994037911858E29281909EcAAA14` |

Edge cases, handy for testing a port:

| Private key | Voucher key | Voucher code |
|---|---|---|
| `0000000000000000000000000000000000000000000000000000000000000001` | `0000000000000000000000000000` | `000000000000001h` |
| `fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140` (n - 1) | `yhjskwdA6OZ1AL1YmHWZWcETkvPc` | `IqwQly7v5TWNN68O` |

Inputs that must be rejected (messages from the Python version):

| Input | Result |
|---|---|
| `2hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISXuq2iftf7zs` (one character changed) | `Invalid voucher check character!` |
| `2hvblF6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7zs` (two characters swapped) | `Invalid voucher check character!` |
| `VdTISeuq2iftf7zs` + `2hvFlb6W2LlmFns9bG3NdCO6l85G` (parts in the wrong order) | `Voucher key must be 28 characters long!` |
| `2hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7z` (old format, no check character) | `Voucher code must be 16 characters long!` |
| Private key of 64 zeros | `Private key is out of secp256k1 range!` |

## Installation

**Go** 1.16 or newer:

```sh
go get github.com/EJ-Research/CryptoVoucher/Golang
```

**Node.js** 14 or newer. The package is not on npm yet, so install it from a checkout:

```sh
git clone https://github.com/EJ-Research/CryptoVoucher.git
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

None of these packages has been published to npm, PyPI or Packagist yet. Until they are, do not install a package with one of these names from a public registry; it would not come from this repository.

The Elixir and Ruby versions still exist but are deprecated, see [below](#elixir-and-ruby-are-deprecated).

## Usage

All four versions expose the same two operations. Private keys are accepted in upper or lower case, and restored keys always come back as 64 lowercase hex characters.

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

Call `restore_private_key` with both parts. If your form takes the voucher in a single field, split it after the 28th character. It fails when:

- the voucher key is not 28 characters or the voucher code is not 16,
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

Bitcoin can use the same key, but a single key has several address types (compressed or uncompressed legacy, nested SegWit, native SegWit, Taproot). The issuer has to tell the merchant which one was funded.

### 3. Check the balance

Always read balances from confirmed or finalized state, not from the latest block. A balance that only exists in a recent block can still disappear.

Common token contracts:

| Network | Token | Contract | Decimals |
|---|---|---|---|
| TRON | USDT (TRC-20) | `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` | 6 |
| Ethereum | USDT (ERC-20) | `0xdAC17F958D2ee523a2206206994597C13D831ec7` | 6 |
| BSC | USDT (BEP-20) | `0x55d398326f99059fF775485246999027B3197955` | 18 |
| Polygon | USDT | `0xc2132D05D31c914a87C6611C10748AEb04B58e8F` | 6 |

Check these against an official source (the token issuer or the network's block explorer) before going live. USDT on BSC uses 18 decimals, not 6.

#### TRON

TronGrid returns the account, including its TRX and TRC-20 balances:

```sh
curl -s "https://api.trongrid.io/v1/accounts/THpApTFkvxbvHThDKix1mwe7KDYxfVtRhP?only_confirmed=true"
```

`data[0].balance` is the TRX balance in sun (1 TRX = 1,000,000 sun) and is missing when it is zero. `data[0].trc20` is an array of one-entry objects, `[{ "<contract>": "<amount>" }]`.

An empty `data` array only means the account has not been activated. It can still hold tokens, because receiving TRC-20 tokens does not activate a TRON account. Before treating a voucher as empty, ask the token contract directly:

```sh
# balanceOf(address) on USDT, read from confirmed state; addresses in hex (0x41 prefix)
curl -s -X POST "https://api.trongrid.io/walletsolidity/triggerconstantcontract" -H "Content-Type: application/json" \
  -d '{"owner_address":"41560b70c1f4bd994037911858e29281909ecaaa14","contract_address":"41a614f803b6fd780986a42c78ec9c7f77e6ded13c","function_selector":"balanceOf(address)","parameter":"000000000000000000000000560b70c1f4bd994037911858e29281909ecaaa14"}'
```

`constant_result[0]` is the balance as a 32-byte hex number (6 decimals for USDT). `41a614f8...ded13c` is `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` in hex. For production traffic, get a TronGrid API key and send it in the `TRON-PRO-API-KEY` header.

With tronweb:

```js
const tronWeb = new TronWeb({ fullHost: "https://api.trongrid.io", privateKey });

const sun = await tronWeb.trx.getBalance(tronAddress); // confirmed balance
const usdt = await tronWeb.contract().at("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t");
const units = await usdt.balanceOf(tronAddress).call({ confirmed: true }); // 6 decimals
```

`trx.getBalance` already reads from the solidity node. For contract calls, `{ confirmed: true }` does the same; without it, the call reads the latest block.

#### Ethereum, BSC and other EVM chains

Use a JSON-RPC endpoint that supports the `finalized` block tag (Ethereum, BSC and Polygon nodes do):

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

Seeing a balance does not make the funds yours; only a confirmed transfer to your own wallet does. After checking the balance, send everything to your own address right away, wait for that transfer to become final, and only then hand over the product. The next section explains why.

How the transfer is paid for depends on what the voucher holds.

#### Native coin vouchers (TRX, ETH, BNB, POL)

The fee is paid from the voucher balance itself.

- On Ethereum, BSC and Polygon, a plain transfer to an address without contract code uses 21,000 gas. Send `balance - 21000 * maxFeePerGas` (or `21000 * gasPrice` for a legacy transaction). The actual fee is usually a bit lower and the difference stays on the voucher address as dust. On rollups such as Arbitrum, Optimism or Base the fee has an extra L1 part, so estimate it with the node instead.
- On TRON a TRX transfer uses bandwidth. It is normally covered by the free daily bandwidth every activated account gets; if not, a small amount of TRX is burned.

#### Token vouchers (USDT and similar)

The voucher address needs native coin to pay the fee (on TRON, delegated energy also works). Without it, the token transfer cannot be sent.

- On TRON, a USDT transfer needs energy. If the address has no energy of its own, the network burns TRX to pay for it. The amount depends on the current energy price and on whether the receiving address currently holds USDT: sending to an address with a zero USDT balance costs about twice as much. Set `feeLimit` high enough, because a transfer that runs out of energy fails and the fee is still burned.
- On Ethereum, BSC and Polygon, a token transfer needs ETH, BNB or POL for gas.

The simplest setup is for the issuer to add enough native coin when funding a token voucher. On TRON this matters twice: an address that has only received tokens is not activated and cannot send anything until it receives TRX. If a voucher arrives without gas, the merchant has to send a small amount of native coin to the voucher address (on TRON, delegating energy also works once the address is activated), wait for it to confirm, and then move the tokens right away. Gas sent this way is exposed just like the tokens, so send only what the transfer needs.

When the transfer is done, anything left on the voucher address (dust, unused gas) can be swept the same way or ignored.

## The race window

Every person or system that has seen the voucher can spend it:

- the issuer's system, which generated the key,
- the buyer,
- anyone who saw the voucher (a photo, a chat message, a receipt left on a counter),
- any other merchant the voucher was shown to.

The race window is the time between your balance check and the moment your sweep becomes final. If any other holder moves the funds during that time, your sweep fails or gets dropped, even though the balance looked fine a moment earlier.

Pending transactions do not close the window:

- On EVM chains, another holder can broadcast a transaction with the same nonce and a higher fee and replace yours before it is mined.
- On Bitcoin, replace-by-fee does the same thing.
- On TRON there is no fee bidding, but if two transactions spend the same balance, whichever lands in a block first wins.

No library can close this window, because any bearer key works this way. What you can do is keep it short and wait for finality before delivering:

1. Read the balance from confirmed or finalized state.
2. Sweep right away, not in a batch job later.
3. Deliver only after the sweep is final:

   | Network | When to treat the sweep as final |
   |---|---|
   | TRON | After 19 blocks (about one minute), when the block is solidified |
   | Ethereum | When the block is finalized, usually 13 to 19 minutes |
   | BSC | When the block is finalized, usually within a few seconds |
   | Polygon | When the block is finalized (`finalized` block tag) |
   | Bitcoin | 1 to 6 confirmations, depending on the amount |

4. Treat a failed or empty sweep as a voucher that has already been used.
5. Record the voucher address together with the sweep transaction ID. If there is a dispute, the chain is the source of truth.

## Elixir and Ruby are deprecated

The Elixir and Ruby versions are no longer maintained. They currently produce the same vouchers as the other four, so nothing breaks for existing users, but they will not get fixes or format changes. Use Go, Node.js, Python or PHP for new projects. The Ruby version prints a warning when it is loaded, and the Elixir functions are marked `@deprecated`.

Go, Node.js, Python and PHP cover nearly every place this library ends up: payment backends, shops, bots and internal tools. Six copies of the same logic are hard to keep in sync. Every format change had to be written, tested and released six times, and the copies drifted apart anyway. The library is one file of under 200 lines and the format is fully specified above, so anyone who needs another language can port it quickly, and AI coding tools can do most of that work today. The sample data and edge cases above are enough to check a port.

## Notes

- The library has been tested on TRON. Ethereum, BSC, Polygon and Bitcoin use the same key type, but test on your own setup before accepting real vouchers.
- Only secp256k1 keys are supported.
- Vouchers created before the check character was added (43 characters or fewer) are rejected by the current version.

## Contributing

Issues and pull requests are welcome. Any change to the voucher format has to land in all four maintained versions (Go, Node.js, Python and PHP) and has to keep the sample data above valid.

## License

[MIT](LICENSE)
