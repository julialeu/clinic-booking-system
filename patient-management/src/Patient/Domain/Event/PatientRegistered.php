<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\Event;

use Clinic\Shared\Domain\DomainEvent;
use DateTimeImmutable;

final readonly class PatientRegistered implements DomainEvent
{
    public function __construct(
        private string $patientId,
        private string $firstName,
        private string $fullName,
        private string $phone,
        private DateTimeImmutable $occurredOn,
    ) {
    }

    public function aggregateId(): string
    {
        return $this->patientId;
    }

    public function eventName(): string
    {
        return 'patient.registered';
    }

    public function occurredOn(): DateTimeImmutable
    {
        return $this->occurredOn;
    }

    public function payload(): array
    {
        return [
            'patientId' => $this->patientId,
            'firstName' => $this->firstName,
            'fullName' => $this->fullName,
            'phone' => $this->phone,
            'occurredOn' => $this->occurredOn->format(DATE_ATOM),
        ];
    }
}