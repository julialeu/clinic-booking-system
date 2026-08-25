<?php

declare(strict_types=1);

namespace Clinic\Shared\Domain;

use DateTimeImmutable;

interface DomainEvent
{
    public function aggregateId(): string;

    public function eventName(): string;

    public function occurredOn(): DateTimeImmutable;

    /** @return array<string, mixed> */
    public function payload(): array;
}