"use strict";

const assert = require("assert");
const fs = require("fs");
const path = require("path");
const { CryptoVoucher } = require("./CryptoVoucher");

const vectors = JSON.parse(fs.readFileSync(path.join(__dirname, "..", "vectors.json"), "utf8"));
const cv = new CryptoVoucher();

function assertError(result, expected, label) {
    assert.strictEqual(result.success, false, label);
    assert.strictEqual(result.message.toLowerCase(), expected.toLowerCase(), label);
}

for (const v of vectors.valid) {
    assert.deepStrictEqual(cv.createVoucher(v.private_key, v.network),
        { success: true, voucherKey: v.voucher_key, voucherCode: v.voucher_code });
    assert.deepStrictEqual(cv.restorePrivateKey(v.voucher_key, v.voucher_code),
        { success: true, data: v.restored_key, network: v.network });
}
for (const v of vectors.create_errors) {
    assertError(cv.createVoucher(v.private_key, v.network), v.error,
        `createVoucher(${JSON.stringify(v.private_key)}, ${JSON.stringify(v.network)})`);
}
for (const v of vectors.restore_errors) {
    assertError(cv.restorePrivateKey(v.voucher_key, v.voucher_code), v.error, v.note);
}

console.log(`${vectors.valid.length + vectors.create_errors.length + vectors.restore_errors.length} vectors passed`);
