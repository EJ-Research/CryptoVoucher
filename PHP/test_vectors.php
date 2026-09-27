<?php

require __DIR__ . '/CryptoVoucher.php';

$vectors = json_decode(file_get_contents(__DIR__ . '/../vectors.json'), true, 512, JSON_THROW_ON_ERROR);
$cv = new CryptoVoucher();
$failures = [];

foreach ($vectors['valid'] as $v) {
    $created = $cv->createVoucher($v['private_key'], $v['network']);
    if ($created !== ['status' => 'success', 'voucher_key' => $v['voucher_key'], 'voucher_code' => $v['voucher_code']]) {
        $failures[] = "createVoucher({$v['private_key']}, {$v['network']}) = " . json_encode($created);
    }
    $restored = $cv->restorePrivateKey($v['voucher_key'], $v['voucher_code']);
    if ($restored !== ['status' => 'success', 'data' => $v['restored_key'], 'network' => $v['network']]) {
        $failures[] = "restorePrivateKey({$v['voucher_key']}, {$v['voucher_code']}) = " . json_encode($restored);
    }
}
foreach ($vectors['create_errors'] as $v) {
    $result = $cv->createVoucher($v['private_key'], $v['network']);
    if ($result['status'] !== 'error' || strcasecmp($result['message'], $v['error']) !== 0) {
        $failures[] = 'createVoucher(' . json_encode($v['private_key']) . ', ' . json_encode($v['network']) . ') = ' . json_encode($result) . ", want {$v['error']}";
    }
}
foreach ($vectors['restore_errors'] as $v) {
    $result = $cv->restorePrivateKey($v['voucher_key'], $v['voucher_code']);
    if ($result['status'] !== 'error' || strcasecmp($result['message'], $v['error']) !== 0) {
        $failures[] = "{$v['note']}: " . json_encode($result) . ", want {$v['error']}";
    }
}

if ($failures) {
    fwrite(STDERR, implode(PHP_EOL, $failures) . PHP_EOL);
    exit(1);
}
echo (count($vectors['valid']) + count($vectors['create_errors']) + count($vectors['restore_errors'])) . ' vectors passed' . PHP_EOL;
