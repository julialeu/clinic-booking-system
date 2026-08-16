<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\ValueObject;

use InvalidArgumentException;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Embeddable]
final readonly class FullName
{
    private function __construct(
         #[ORM\Column(name: 'first_name', type: 'string', length: 80)]
        private string $firstName,

        #[ORM\Column(name: 'first_surname', type: 'string', length: 80)]
        private string $firstSurname,

        #[ORM\Column(name: 'second_surname', type: 'string', length: 80, nullable: true)]
        private ?string $secondSurname,
    ) {
    }

    public static function create(
        string $firstName,
        string $firstSurname,
        ?string $secondSurname = null,
    ): self {
        $firstName = trim($firstName);
        $firstSurname = trim($firstSurname);
        $secondSurname = $secondSurname !== null ? trim($secondSurname) : null;

        if ($firstName === '') {
            throw new InvalidArgumentException('First name cannot be empty');
        }

        if ($firstSurname === '') {
            throw new InvalidArgumentException('First surname cannot be empty');
        }

        if ($secondSurname === '') {
            $secondSurname = null;
        }

        return new self($firstName, $firstSurname, $secondSurname);
    }

    public function firstName(): string
    {
        return $this->firstName;
    }

    public function firstSurname(): string
    {
        return $this->firstSurname;
    }

    public function secondSurname(): ?string
    {
        return $this->secondSurname;
    }

    public function surnames(): string
    {
        return $this->secondSurname === null
            ? $this->firstSurname
            : "{$this->firstSurname} {$this->secondSurname}";
    }

    public function full(): string
    {
        return "{$this->firstName} {$this->surnames()}";
    }

    public function informal(): string
    {
        return $this->firstName;
    }

    public function equals(self $other): bool
    {
        return $this->firstName === $other->firstName
            && $this->firstSurname === $other->firstSurname
            && $this->secondSurname === $other->secondSurname;
    }
}