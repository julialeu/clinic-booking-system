<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\ValueObject;

use InvalidArgumentException;
use Ramsey\Uuid\Uuid;

final readonly class PatientId
{
    private function __construct(private string $value)
    {
    }

    public static function generate(): self
    {
        return new self(Uuid::uuid4()->toString());
    }

    public static function fromString(string $value): self
    {
        if (! Uuid::isValid($value)) {
            throw new InvalidArgumentException("Invalid patient id: {$value}");
        }

        return new self($value);
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