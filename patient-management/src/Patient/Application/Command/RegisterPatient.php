<?php

declare(strict_types=1);

namespace Clinic\Patient\Application\Command;

final readonly class RegisterPatient
{
    public function __construct(
        public string $firstName,
        public string $firstSurname,
        public ?string $secondSurname,
        public string $phone,
        public ?string $email,
    ) {
    }
}