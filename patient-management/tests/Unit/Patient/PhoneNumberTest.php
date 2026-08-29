<?php

declare(strict_types=1);

use Clinic\Patient\Domain\ValueObject\PhoneNumber;

test('accepts a valid international number', function (): void {
    $phone = PhoneNumber::fromString('+34600111222');

    expect($phone->value())->toBe('+34600111222');
});

test('normalises spaces and separators', function (string $input): void {
    expect(PhoneNumber::fromString($input)->value())->toBe('+34600111222');
})->with([
    '+34 600 111 222',
    '+34-600-111-222',
    '  +34600111222  ',
    '+34 (600) 111222',
]);

test('rejects numbers without an international prefix', function (): void {
    expect(fn () => PhoneNumber::fromString('600111222'))
        ->toThrow(InvalidArgumentException::class);
});

test('rejects malformed numbers', function (string $input): void {
    expect(fn () => PhoneNumber::fromString($input))
        ->toThrow(InvalidArgumentException::class);
})->with([
    '',
    '+',
    '+0600111222',
    '+3460011122233445566',
    'not a phone',
]);

test('compares by value', function (): void {
    $a = PhoneNumber::fromString('+34600111222');
    $b = PhoneNumber::fromString('+34 600 111 222');

    expect($a->equals($b))->toBeTrue();
});