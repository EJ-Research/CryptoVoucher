# CryptoVoucher

CryptoVoucher turns a wallet private key into a short voucher that can be sold, printed or sent as text, and redeemed like a gift card. The voucher also says which network and coin it holds, for example USDT on TRON or BTC on Bitcoin. It is a small library for Go, Node.js, Python and PHP.

## The idea

Getting someone started with crypto usually means walking them through a wallet app, a backup phrase, networks and fees. A voucher removes that step. The issuer creates a fresh wallet, funds it and gives the buyer a 45-character code. The buyer does not need a wallet at all. To spend the voucher, they give the code to a merchant, who turns it back into the private key, checks the balance of the address and moves the funds to their own wallet.

The voucher is the private key itself, written in a shorter alphabet, plus a network code and a check character. Nothing is encrypted and no server is involved. Anyone who holds the full voucher controls the funds, the same way anyone holding a banknote can spend it. This is why redeeming has to follow a fixed order (see [Redeeming a voucher](#redeeming-a-voucher) and [The race window](#the-race-window)).

One key has an address on many networks, and all EVM chains even share the same address. The network code tells the merchant which chain, which coin and which address type to check.

The whole flow:

1. The issuer generates a new private key and funds its address on one network.
2. The issuer calls `create_voucher` with the key and the network ID, and gives the voucher to the buyer.
3. The buyer passes the voucher to a merchant.
4. The merchant calls `restore_private_key`, which returns the key and the network ID, derives the address for that network and checks the balance.
5. The merchant moves the full balance to their own wallet, waits until that transfer is final, and only then delivers.

## Voucher format

A private key is a 256-bit number, normally written as 64 hex characters. To build a voucher, the library:

1. checks that the key is in the range `1 <= key < n`, where `n` is the order of the secp256k1 curve,
2. writes the network code (one character, see [Network codes](#network-codes)),
3. writes the key in Base62 (`0-9`, `A-Z`, `a-z`), left-padded with `0` to 43 characters,
4. appends one check character,
5. splits the 45 characters into a voucher key and a voucher code.

| Part | Length | Content |
|---|---|---|
| Voucher key | 29 | Network code, then the first 28 Base62 characters |
| Voucher code | 16 | Last 15 Base62 characters, then the check character |

43 is the shortest Base62 length that fits every 256-bit value (62^43 is just above 2^256), so the key part cannot get any shorter without losing information. The padding zeros have no effect on the value, and the decoder treats them like any other digit.

The check character works like the ISBN-10 check digit. The Base62 value of each of the first 44 characters (`0-9` = 0 to 9, `A-Z` = 10 to 35, `a-z` = 36 to 61), that is the network code and the 43 key digits, is multiplied by its position (1 to 44). The products are added up, and the sum modulo 61 picks the check character (`0` to `y`, so `z` never appears in that position). Because 61 is prime, this catches:

- every mistyped character, except `0` typed as `z` or the other way round,
- every swap of two characters, whether they are next to each other or further apart, again except `0` with `z`.

A typo in the network code is always caught, because `0` and `z` are never used as network codes. Other kinds of damage get through about once in 61 tries. The one fixed blind spot is a doubled character in the first two places of the voucher code (positions 30 and 31) that turns into a different doubled character, such as `44` typed as `77`.

`restore_private_key` also checks the length of each part, so entering the voucher key and the voucher code in each other's fields is always caught. If your form takes the whole voucher in a single field, a reversed paste can only be caught by the check character. A voucher that fails any of these checks is rejected instead of silently decoding to some other valid key.

Both parts are needed to restore the key, and both must be kept secret. Do not print or display the voucher key as if it were a public card number. It holds about two thirds of the private key, and on secp256k1 networks, once the voucher address has sent any transaction (which puts its public key on chain), the voucher key alone is enough to recover the rest with modest hardware.

Vouchers are case sensitive: `a` and `A` are different characters.

## Network codes

The first character of every voucher is one of these codes. The network ID is what `create_voucher` takes and `restore_private_key` returns.

| Code | Network ID | Network | Asset | Token contract | Decimals |
|---|---|---|---|---|---|
| `1` | `BITCOIN_BTC` | Bitcoin | BTC | native | 8 |
| `2` | `ETHEREUM_ETH` | Ethereum | ETH | native | 18 |
| `3` | `ETHEREUM_USDT` | Ethereum | USDT | `0xdAC17F958D2ee523a2206206994597C13D831ec7` | 6 |
| `4` | `ETHEREUM_USDC` | Ethereum | USDC | `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` | 6 |
| `5` | `TRON_TRX` | TRON | TRX | native | 6 |
| `6` | `TRON_USDT` | TRON | USDT | `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` | 6 |
| `7` | `BSC_BNB` | BNB Smart Chain | BNB | native | 18 |
| `8` | `BSC_USDT` | BNB Smart Chain | USDT (Binance-Peg) | `0x55d398326f99059fF775485246999027B3197955` | 18 |
| `9` | `BSC_USDC` | BNB Smart Chain | USDC (Binance-Peg) | `0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d` | 18 |
| `A` | `POLYGON_POL` | Polygon PoS | POL | native | 18 |
| `B` | `POLYGON_USDT` | Polygon PoS | USDT | `0xc2132D05D31c914a87C6611C10748AEb04B58e8F` | 6 |
| `C` | `POLYGON_USDC` | Polygon PoS | USDC | `0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359` | 6 |
| `D` | `SOLANA_SOL` | Solana | SOL | native | 9 |
| `E` | `SOLANA_USDT` | Solana | USDT | `Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB` | 6 |
| `F` | `SOLANA_USDC` | Solana | USDC | `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v` | 6 |
| `G` | `TON_TON` | TON | TON | native | 9 |
| `H` | `TON_USDT` | TON | USDT | `EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs` | 6 |
| `J` | `ARBITRUM_ETH` | Arbitrum One | ETH | native | 18 |
| `K` | `ARBITRUM_USDT` | Arbitrum One | USDT (USDT0) | `0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9` | 6 |
| `L` | `ARBITRUM_USDC` | Arbitrum One | USDC | `0xaf88d065e77c8cC2239327C5EDb3A432268e5831` | 6 |
| `M` | `OPTIMISM_ETH` | OP Mainnet | ETH | native | 18 |
| `N` | `OPTIMISM_USDT` | OP Mainnet | USDT (bridged) | `0x94b008aA00579c1307B0EF2c499aD98a8ce58e58` | 6 |
| `P` | `OPTIMISM_USDC` | OP Mainnet | USDC | `0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85` | 6 |
| `Q` | `BASE_ETH` | Base | ETH | native | 18 |
| `R` | `BASE_USDC` | Base | USDC | `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` | 6 |
| `S` | `AVALANCHE_AVAX` | Avalanche C-Chain | AVAX | native | 18 |
| `T` | `AVALANCHE_USDT` | Avalanche C-Chain | USDT | `0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7` | 6 |
| `U` | `AVALANCHE_USDC` | Avalanche C-Chain | USDC | `0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E` | 6 |
| `V` | `LITECOIN_LTC` | Litecoin | LTC | native | 8 |
| `W` | `DOGECOIN_DOGE` | Dogecoin | DOGE | native | 8 |
| `X` | `BITCOINCASH_BCH` | Bitcoin Cash | BCH | native | 8 |
| `Y` | `XRPL_XRP` | XRP Ledger | XRP | native | 6 |
| `a` | `ETHEREUM_DAI` | Ethereum | DAI | `0x6B175474E89094C44Da98b954EedeAC495271d0F` | 18 |
| `b` | `ETHEREUM_PYUSD` | Ethereum | PYUSD | `0x6c3ea9036406852006290770BEdFcAbA0e23A0e8` | 6 |

A voucher holds exactly the asset its code names, at exactly the contract listed. Other versions of the same coin are different tokens and do not count, for example USDC.e on Polygon or Arbitrum, USDT.e on Avalanche, or USDT0 on OP Mainnet (`0x01bFF41798a0BcF287b996046Ca68b395DbC1071`). Every token contract above was cross-checked in at least two independent token registries, but still check them against the token issuer or the network's block explorer before going live. USDT and USDC on BNB Smart Chain use 18 decimals, not 6.

How the network turns the voucher's number into an address:

| Networks | Key | Address |
|---|---|---|
| Ethereum (chain ID 1), BNB Smart Chain (56), Polygon PoS (137), Arbitrum One (42161), OP Mainnet (10), Base (8453), Avalanche C-Chain (43114) | secp256k1 private key | EVM address `0x...`, the same on every EVM chain |
| TRON | secp256k1 private key | The EVM address bytes with a `0x41` prefix, in Base58Check: `T...` |
| Bitcoin | secp256k1, compressed public key | Native SegWit P2WPKH: `bc1q...` |
| Litecoin | secp256k1, compressed public key | Native SegWit P2WPKH: `ltc1q...` |
| Dogecoin | secp256k1, compressed public key | P2PKH: `D...` |
| Bitcoin Cash | secp256k1, compressed public key | P2PKH in CashAddr: `bitcoincash:q...` |
| XRP Ledger | secp256k1, compressed public key | Classic address: `r...` |
| Solana | Ed25519 seed | Base58 public key; tokens sit in the associated token account |
| TON | Ed25519 seed | W5 wallet (v5r1) on workchain 0 with the default wallet ID 2147483409 (`0x7FFFFF11`), non-bounceable form `UQ...` |

On Solana and TON the 32 bytes are used as the Ed25519 seed. The voucher still requires `1 <= key < n`; a random seed falls outside that range with a probability of about 2^-128, and if it ever happens the issuer generates another one.

The table follows three rules:

- An assigned code is permanent. It is never changed or reused, and new networks only get codes that have never been used.
- `0` and `z` are never assigned, since the check character cannot tell them apart. `I`, `O`, `l` and `o` are left out because they are easy to misread.
- The library rejects any code that is not in the table. There is no "other" or "unspecified" code.

## Sample data

This key is public. Do not send funds to its addresses.

| Field | Value |
|---|---|
| Private key | `0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B` |
| Network | `TRON_USDT` |
| Voucher key | `62hvFlb6W2LlmFns9bG3NdCO6l85G` |
| Voucher code | `VdTISeuq2iftf7zh` |

Addresses of the same key on each network family:

| Network | Address |
|---|---|
| EVM chains | `0x560b70C1F4Bd994037911858E29281909EcAAA14` |
| TRON | `THpApTFkvxbvHThDKix1mwe7KDYxfVtRhP` |
| Bitcoin | `bc1qk5m5qsegkvgvl8a0s85wqm4fedgv7ezcj44fyv` |
| Litecoin | `ltc1qk5m5qsegkvgvl8a0s85wqm4fedgv7ezckf0duu` |
| Dogecoin | `DMfH2tUKF29NiYrLh3J9m8v8M3x7w9xxvJ` |
| Bitcoin Cash | `bitcoincash:qz6nwszr9ze3pnul47q73crw4894pnmytq5jp48w4h` |
| XRP Ledger | `rHXBVdXCAcEaBYCjxTJbD4kXTvDFcDtg9j` |
| Solana | `Fru5TeY998yrMjQPiroZQvwH62196kTFZDYtaB35YQ1C` |
| Solana, USDC token account | `9qSJDRQgrhsUXtDpQJ9UBpt9jho6AK5ETu9uXsL9ftBY` |
| Solana, USDT token account | `6cfJYfWqMMy3GTyT3t4bKST4hZFp9qV5SiwfKw5ZBWMu` |
| TON | `UQDnNYjgOR9KMfFq_2z7ugFLOmsrIOdbSj_wkIE1MJCR0eb9` |

Edge cases, handy for testing a port:

| Private key | Network | Voucher key | Voucher code |
|---|---|---|---|
| `0000000000000000000000000000000000000000000000000000000000000001` | `BITCOIN_BTC` | `10000000000000000000000000000` | `000000000000001j` |
| `fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140` (n - 1) | `ETHEREUM_PYUSD` | `byhjskwdA6OZ1AL1YmHWZWcETkvPc` | `IqwQly7v5TWNN687` |

Inputs that must be rejected (messages from the Python version):

| Input | Result |
|---|---|
| `62hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISXuq2iftf7zh` (one character changed) | `Invalid voucher check character!` |
| `62hvblF6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7zh` (two characters swapped) | `Invalid voucher check character!` |
| `52hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7zh` (network code mistyped) | `Invalid voucher check character!` |
| `02hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7zh` (code not in the table) | `Unknown network code!` |
| `VdTISeuq2iftf7zh` + `62hvFlb6W2LlmFns9bG3NdCO6l85G` (parts in the wrong order) | `Voucher key must be 29 characters long!` |
| `2hvFlb6W2LlmFns9bG3NdCO6l85G` + `VdTISeuq2iftf7z` (first release format, no network code) | `Voucher key must be 29 characters long!` |
| Network ID `TRON` when creating a voucher | `Unknown network!` |
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

All four versions expose the same two operations. Private keys are accepted in upper or lower case, and restored keys always come back as 64 lowercase hex characters. Network IDs are exactly the ones in the [table](#network-codes).

**Go**

```go
import cryptovoucher "github.com/EJ-Research/CryptoVoucher/Golang"

cv := cryptovoucher.NewCryptoVoucher()

voucherKey, voucherCode, err := cv.CreateVoucher(privateKey, "TRON_USDT")
if err != nil {
	// invalid private key or unknown network
}

restoredKey, network, err := cv.RestorePrivateKey(voucherKey, voucherCode)
if err != nil {
	// wrong length, unknown network code, bad character, failed check or out of range
}
```

**Node.js**

```js
const { CryptoVoucher } = require("crypto-voucher");

const cv = new CryptoVoucher();

const voucher = cv.createVoucher(privateKey, "TRON_USDT");
// { success: true, voucherKey: "...", voucherCode: "..." }
// { success: false, message: "..." }

const restored = cv.restorePrivateKey(voucher.voucherKey, voucher.voucherCode);
// { success: true, data: "0b6bf6...", network: "TRON_USDT" }
```

**Python**

```python
from CryptoVoucher import CryptoVoucher

cv = CryptoVoucher()

voucher_key, voucher_code, error = cv.create_voucher(private_key, "TRON_USDT")
restored_key, network, error = cv.restore_private_key(voucher_key, voucher_code)
```

`error` is `None` on success and a message otherwise.

**PHP**

```php
require 'vendor/autoload.php';

$cv = new CryptoVoucher();

$voucher = $cv->createVoucher($privateKey, 'TRON_USDT');
// ['status' => 'success', 'voucher_key' => '...', 'voucher_code' => '...']

$restored = $cv->restorePrivateKey($voucher['voucher_key'], $voucher['voucher_code']);
// ['status' => 'success', 'data' => '0b6bf6...', 'network' => 'TRON_USDT']
// ['status' => 'error', 'message' => '...']
```

Running the Node.js, Python or PHP file directly prints the sample voucher above, and `go test` runs the same example in Go. Importing the library has no side effects.

## Redeeming a voucher

### 1. Decode it

Call `restore_private_key` with both parts. If your form takes the voucher in a single field, split it after the 29th character. It fails when:

- the voucher key is not 29 characters or the voucher code is not 16,
- the first character is not a code from the network table,
- a character is outside `0-9A-Za-z`,
- the check character does not match, which almost always means a typo,
- the decoded number is outside `1 <= key < n`.

On failure, ask the customer to type the voucher again. Never try to "fix" it by guessing characters. Show the decoded network (for example "USDT on TRON") before going on, so the customer can confirm it.

### 2. Get the address

Use the network ID to pick the address rule from the [table above](#network-codes). For EVM chains and TRON, in Node.js:

```js
const { Wallet } = require("ethers");      // ethers v6
const { TronWeb } = require("tronweb");    // tronweb v6

const evmAddress = new Wallet("0x" + privateKey).address;      // Ethereum, BSC, Polygon, Arbitrum, OP, Base, Avalanche
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

The other networks, in Node.js:

```js
const bitcoin = require("bitcoinjs-lib");
const ecc = require("tiny-secp256k1");
const { ECPairFactory } = require("ecpair");
const xrpl = require("xrpl");
const { Keypair, PublicKey } = require("@solana/web3.js");
const { getAssociatedTokenAddressSync } = require("@solana/spl-token");
const { keyPairFromSeed } = require("@ton/crypto");
const { WalletContractV5R1 } = require("@ton/ton");

const secret = Buffer.from(privateKey, "hex");

// Bitcoin family and XRP Ledger: compressed secp256k1 public key
const pubkey = Buffer.from(ECPairFactory(ecc).fromPrivateKey(secret, { compressed: true }).publicKey);
const btcAddress = bitcoin.payments.p2wpkh({ pubkey, network: bitcoin.networks.bitcoin }).address;
const xrpAddress = xrpl.deriveAddress(pubkey.toString("hex").toUpperCase());

// Solana: the owner address, and the token account that holds USDC
const owner = Keypair.fromSeed(secret).publicKey;
const usdcAccount = getAssociatedTokenAddressSync(new PublicKey("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"), owner);

// TON: W5 wallet on workchain 0
const tonKeys = keyPairFromSeed(secret);
const tonAddress = WalletContractV5R1.create({ workchain: 0, publicKey: tonKeys.publicKey }).address.toString({ bounceable: false });
```

Litecoin uses the same P2WPKH script with the `ltc` bech32 prefix, Dogecoin uses P2PKH with version byte `0x1e`, and Bitcoin Cash uses P2PKH written in CashAddr. Compare your results with the sample addresses above before going live.

### 3. Check the balance

Always read balances from confirmed or finalized state, not from the latest block. A balance that only exists in a recent block can still disappear. Token contracts and decimals are in the [network table](#network-codes).

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

#### EVM chains

Use a JSON-RPC endpoint for the chain the voucher names. Nodes of all seven EVM chains in the table support the `finalized` block tag; on Avalanche it returns the last accepted block.

```sh
# native coin (ETH, BNB, POL, AVAX), result in wei as hex
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

#### Other networks

- **Solana:** read SOL with `getBalance` and tokens with `getTokenAccountBalance` on the associated token account, both with the `finalized` commitment. If the associated token account does not exist, the call returns an error; treat that as zero. Tokens can also sit in other token accounts of the same owner, so `getTokenAccountsByOwner` with a mint filter finds everything.
- **TON:** read TON from the wallet address. For USDT, ask the USDT contract for the owner's jetton wallet with `get_wallet_address`, then read that wallet's balance with `get_wallet_data`. A jetton wallet that is not deployed yet holds nothing.
- **Bitcoin, Litecoin, Dogecoin, Bitcoin Cash:** add up the confirmed unspent outputs of the address, using your own node, an Electrum server or an indexer.
- **XRP Ledger:** call `account_info` against the `validated` ledger. The balance is in drops (1 XRP = 1,000,000 drops), and part of it is locked as the account reserve.

### 4. Move the funds before you deliver

Seeing a balance does not make the funds yours; only a confirmed transfer to your own wallet does. After checking the balance, send everything to your own address at once, wait for that transfer to become final, and only then hand over the product. The next section explains why.

How the transfer is paid for depends on what the voucher holds.

#### Native coin vouchers

The fee is paid from the voucher balance itself.

- On Ethereum, BSC and Polygon, a plain transfer to an address without contract code uses 21,000 gas. Send `balance - 21000 * maxFeePerGas` (or `21000 * gasPrice` for a legacy transaction). The actual fee is usually a bit lower and the difference stays on the voucher address as dust. On Avalanche the same rule applies. Rollups add an L1 part to the fee: on Arbitrum, `eth_estimateGas` already includes it, while on OP Mainnet and Base you also have to subtract the L1 data fee, which `getL1Fee` on the GasPriceOracle contract (`0x420000000000000000000000000000000000000F`) returns.
- On TRON a TRX transfer uses bandwidth. It is normally covered by the free daily bandwidth every activated account gets; if not, a small amount of TRX is burned.
- On Bitcoin, Litecoin, Dogecoin and Bitcoin Cash, spend every unspent output in one transaction and take the fee out of the amount.
- On Solana the fee comes out of the SOL balance. Send the balance minus the fee exactly; leaving a remainder below the rent-exempt minimum makes the transfer fail.
- On TON the W5 wallet contract is deployed by its first outgoing transfer, which the balance pays for. Send with mode 130 (128 to carry the whole remaining balance, plus 2, which W5 requires on every action of an external message). `@ton/ton` adds the 2 by itself; hand-built messages must include it, or nothing moves.
- On the XRP Ledger only the balance above the account reserve can be sent. Most of the reserve can be recovered by deleting the account, which is only allowed once the account is old enough.

#### Token vouchers

The voucher address needs native coin to pay the fee. The only exception is TON, where a W5 wallet can pay a gasless relay in USDT instead.

- On TRON, a USDT transfer needs energy. If the address has no energy of its own, the network burns TRX to pay for it. The amount depends on the current energy price and on whether the receiving address currently holds USDT: sending to an address with a zero USDT balance costs about twice as much. Set `feeLimit` high enough, because a transfer that runs out of energy fails and the fee is still burned. Delegated energy also works.
- On the EVM chains, a token transfer needs ETH, BNB, POL or AVAX for gas.
- On Solana, the transfer needs SOL for the fee, and if your own token account for that token does not exist yet, whoever creates it pays its rent deposit.
- On TON, the transfer needs TON for gas. W5 wallets can also send USDT through gasless relay services that take their fee in USDT.

The simplest setup is for the issuer to add enough native coin when funding a token voucher. On TRON it is required: an address that has only received tokens is not activated and cannot send anything until it receives TRX. If a voucher arrives without gas, the merchant has to send a small amount of native coin to the voucher address, wait for it to confirm, and then move the tokens immediately. Gas sent this way is exposed just like the tokens, so send only what the transfer needs.

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
- On TRON and the other networks, if two transactions spend the same balance, whichever lands in a block first wins.

No library can close this window, because any bearer key works this way. Keep it short and wait for finality before delivering:

1. Read the balance from confirmed or finalized state.
2. Sweep immediately, not in a batch job later.
3. Deliver only after the sweep is final:

   | Network | When to treat the sweep as final |
   |---|---|
   | TRON | After 19 blocks (about one minute), when the block is solidified |
   | Ethereum | When the block is finalized, usually 13 to 19 minutes |
   | BNB Smart Chain | When the block is finalized, usually within a few seconds |
   | Polygon PoS | When the block is finalized (`finalized` block tag) |
   | Arbitrum One, OP Mainnet, Base | When the block is finalized (`finalized` block tag), which follows Ethereum finality |
   | Avalanche C-Chain | When the block is accepted, usually within a few seconds |
   | Solana | At the `finalized` commitment, usually under a minute |
   | TON | When the transaction is in a masterchain-confirmed block, usually within seconds |
   | XRP Ledger | When the transaction is in a validated ledger, usually within seconds |
   | Bitcoin, Bitcoin Cash | 1 to 6 confirmations, depending on the amount |
   | Litecoin, Dogecoin | More confirmations than on Bitcoin, since blocks come faster (2.5 minutes and 1 minute) |

4. Treat a failed or empty sweep as a voucher that has already been used.
5. Record the voucher address together with the sweep transaction ID. If there is a dispute, the chain is the source of truth.

## Elixir and Ruby are deprecated

The Elixir and Ruby versions are no longer maintained and will not be updated. They were frozen at an interim 44-character format without a network code. No release ever used that format, and the maintained versions do not accept it, so do not use them for new vouchers. The Ruby version prints a warning when it is loaded, and the Elixir functions are marked `@deprecated`.

Go, Node.js, Python and PHP are what most payment backends and shops are written in today. Six copies of the same logic are hard to keep in sync: every format change had to be written and tested six times, and the copies drifted apart anyway. The library is one file of 220 to 250 lines and the format is fully specified above, so porting it to another language is a small job, and AI coding tools can do most of it. The sample data and edge cases above are enough to check a port.

## Notes

- The voucher flow has been tested on TRON. The other networks use standard key and address rules, but test on your own setup before accepting real vouchers.
- The voucher always carries a number in `1 <= key < n`. secp256k1 networks use it as the private key, Solana and TON as the Ed25519 seed. Networks with other key schemes, such as Cardano, Monero or Polkadot, are not supported.
- Vouchers without a network code are rejected by the current version. That includes vouchers from the first release, which used plain Base62 with no padding and no check character (a 28-character key plus up to 15 characters). To decode one of those, use the code from the first release (commit `fea1e8f`).

## Contributing

Issues and pull requests are welcome. Any change to the voucher format or the network table has to land in all four maintained versions (Go, Node.js, Python and PHP) and has to keep the sample data above valid. New networks get new codes; existing codes never change.

## License

[MIT](LICENSE)
