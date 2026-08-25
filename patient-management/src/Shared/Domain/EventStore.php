<?php

declare(strict_types=1);

namespace Clinic\Shared\Domain;

interface EventStore
{
    /** @param list<DomainEvent> $events */
    public function append(array $events): void;
}