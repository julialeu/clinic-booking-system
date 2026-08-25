<?php

declare(strict_types=1);

namespace Clinic\Shared\Infrastructure;

use Clinic\Shared\Domain\DomainEvent;
use Clinic\Shared\Domain\EventStore;
use Doctrine\ORM\EntityManagerInterface;

final readonly class DoctrineEventStore implements EventStore
{
    public function __construct(private EntityManagerInterface $entityManager)
    {
    }

    public function append(array $events): void
    {
        if ($events === []) {
            return;
        }

        $connection = $this->entityManager->getConnection();

        foreach ($events as $event) {
            $connection->insert('outbox_events', [
                'aggregate_id' => $event->aggregateId(),
                'event_name' => $event->eventName(),
                'payload' => json_encode($event->payload(), JSON_THROW_ON_ERROR),
                'occurred_on' => $event->occurredOn()->format('Y-m-d H:i:sP'),
            ]);
        }
    }
}