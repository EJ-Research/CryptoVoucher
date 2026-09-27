"""Keeps every copy of the network table in sync with networks.json.

    python tools/sync_registry.py                        rewrite the generated tables
    python tools/sync_registry.py --check                fail if any table is out of date or a network has no vector
    python tools/sync_registry.py --check --previous F   also fail if a code in F was changed or removed

networks.json is the only place where network codes are edited. The tables in the four
implementations and in README.md sit between "BEGIN NETWORKS" and "END NETWORKS" markers
and are generated from it.
"""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
KEYS = {"secp256k1", "ed25519-seed"}
ADDRESSES = {
    "evm", "tron", "bitcoin-p2wpkh", "litecoin-p2wpkh", "dogecoin-p2pkh",
    "bitcoincash-cashaddr", "xrpl-classic", "solana", "ton-w5",
}

TARGETS = {
    "Python/CryptoVoucher.py": lambda n: '    "%s": "%s",' % (n["code"], n["id"]),
    "Golang/cryptovoucher.go": lambda n: "\t'%s': \"%s\"," % (n["code"], n["id"]),
    "Nodejs/CryptoVoucher.js": lambda n: '    ["%s", "%s"],' % (n["code"], n["id"]),
    "PHP/CryptoVoucher.php": lambda n: "        '%s' => '%s'," % (n["code"], n["id"]),
}


def fail(message):
    sys.exit("networks.json: " + message)


def validate(registry):
    reserved = set(registry["never_assigned"]) | set(registry["left_out"])
    codes, ids = set(), set()
    for n in registry["networks"]:
        label = n.get("id", "?")
        code = n["code"]
        if len(code) != 1 or code not in ALPHABET or code in reserved:
            fail("%s: code %r is not a usable Base62 character" % (label, code))
        if code in codes or n["id"] in ids:
            fail("%s: duplicate code or ID" % label)
        codes.add(code)
        ids.add(n["id"])
        if not re.fullmatch(r"[A-Z0-9]+_[A-Z0-9]+", n["id"]):
            fail("%s: ID must look like NETWORK_ASSET" % label)
        if not isinstance(n["decimals"], int) or not 0 <= n["decimals"] <= 36:
            fail("%s: bad decimals" % label)
        if n["key"] not in KEYS or n["address"] not in ADDRESSES:
            fail("%s: unknown key type or address format" % label)
        if not re.fullmatch(r"[-a-z0-9]{3,8}:[-_a-zA-Z0-9]{1,32}", n["caip2"]):
            fail("%s: bad CAIP-2 chain ID" % label)
        if n["contract"] is None:
            if not isinstance(n["slip44"], int) or n["caip19"] != "%s/slip44:%d" % (n["caip2"], n["slip44"]):
                fail("%s: a native asset needs slip44 and a matching CAIP-19 ID" % label)
        else:
            if n["slip44"] is not None:
                fail("%s: a token has no slip44 coin type" % label)
            caip19 = n["caip19"]
            if caip19 is not None and not (caip19.startswith(n["caip2"] + "/") and caip19.endswith(":" + n["contract"])):
                fail("%s: CAIP-19 ID must be the chain ID plus the contract" % label)


def check_permanence(registry, previous_path):
    previous = json.loads(Path(previous_path).read_text(encoding="utf-8"))
    current = {n["code"]: n["id"] for n in registry["networks"]}
    for n in previous["networks"]:
        if current.get(n["code"]) != n["id"]:
            fail("code %r (%s) was changed or removed; assigned codes are permanent" % (n["code"], n["id"]))


def check_vectors(registry):
    vectors = json.loads((ROOT / "vectors.json").read_text(encoding="utf-8"))
    covered = {v["network"] for v in vectors["valid"]}
    missing = [n["id"] for n in registry["networks"] if n["id"] not in covered]
    if missing:
        sys.exit("vectors.json: no valid vector for " + ", ".join(missing))


def readme_table(registry):
    lines = ["| Code | Network ID | Network | Asset | Token contract | Decimals |", "|---|---|---|---|---|---|"]
    for n in registry["networks"]:
        asset = n["asset"] + (" (%s)" % n["variant"] if n.get("variant") else "")
        contract = "`%s`" % n["contract"] if n["contract"] else "native"
        lines.append("| `%s` | `%s` | %s | %s | %s | %d |" % (n["code"], n["id"], n["network"], asset, contract, n["decimals"]))
    return [""] + lines + [""]


def replace_block(text, lines, path):
    source = text.split("\n")
    begin = [i for i, line in enumerate(source) if "BEGIN NETWORKS" in line]
    end = [i for i, line in enumerate(source) if "END NETWORKS" in line]
    if len(begin) != 1 or len(end) != 1 or begin[0] > end[0]:
        sys.exit("%s: expected one BEGIN NETWORKS / END NETWORKS pair" % path)
    return "\n".join(source[:begin[0] + 1] + lines + source[end[0]:])


def main(argv):
    check = "--check" in argv
    registry = json.loads((ROOT / "networks.json").read_text(encoding="utf-8"))
    validate(registry)
    if "--previous" in argv:
        check_permanence(registry, argv[argv.index("--previous") + 1])
    # Only in --check mode: a new network's vector is made with the regenerated tables
    if check:
        check_vectors(registry)

    outputs = {path: [render(n) for n in registry["networks"]] for path, render in TARGETS.items()}
    outputs["README.md"] = readme_table(registry)

    stale = []
    for path, lines in outputs.items():
        file = ROOT / path
        text = file.read_text(encoding="utf-8")
        updated = replace_block(text, lines, path)
        if updated != text:
            stale.append(path)
            if not check:
                file.write_text(updated, encoding="utf-8")
    if check and stale:
        sys.exit("out of date, run python tools/sync_registry.py: " + ", ".join(stale))
    print("%d networks; %s" % (len(registry["networks"]), "updated " + ", ".join(stale) if stale else "all tables in sync"))


if __name__ == "__main__":
    main(sys.argv[1:])
