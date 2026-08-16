<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\ValueObject;

use InvalidArgumentException;

final readonly class PhoneNumber
{
    private function __construct(private string $value)
    {
    }

    public static function fromString(string $value): self
    {
        $normalised = preg_replace('/[\s\-\(\)]/', '', trim($value)) ?? '';

        if (preg_match('/^\+[1-9]\d{7,14}$/', $normalised) !== 1) {
            throw new InvalidArgumentException("Invalid phone number: {$value}");
        }

        return new self($normalised);
    }

    public function value(): string
    {
        return $this->value;
    }

    public function equals(self $other): bool
    {
        return $this->value === $other->value;
    }
}