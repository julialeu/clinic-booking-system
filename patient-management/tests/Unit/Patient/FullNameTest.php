<?php

declare(strict_types=1);

use Clinic\Patient\Domain\ValueObject\FullName;

test('builds a name with two surnames', function (): void {
    $name = FullName::create('María', 'García', 'López');

    expect($name->full())->toBe('María García López')
        ->and($name->surnames())->toBe('García López')
        ->and($name->informal())->toBe('María');
});

test('second surname is optional', function (): void {
    $name = FullName::create('Ana', 'Ruiz');

    expect($name->full())->toBe('Ana Ruiz')
        ->and($name->secondSurname())->toBeNull();
});

test('treats an empty second surname as absent', function (): void {
    $withNull = FullName::create('Ana', 'Ruiz', null);
    $withEmpty = FullName::create('Ana', 'Ruiz', '');

    expect($withEmpty->equals($withNull))->toBeTrue();
});

test('trims surrounding whitespace', function (): void {
    $name = FullName::create('  Ana  ', '  Ruiz  ');

    expect($name->full())->toBe('Ana Ruiz');
});

test('requires a first name', function (): void {
    expect(fn () => FullName::create('', 'Ruiz'))
        ->toThrow(InvalidArgumentException::class);
});

test('requires a first surname', function (): void {
    expect(fn () => FullName::create('Ana', '   '))
        ->toThrow(InvalidArgumentException::class);
});