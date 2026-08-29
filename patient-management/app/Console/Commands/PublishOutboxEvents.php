<?php

declare(strict_types=1);

namespace App\Console\Commands;

use Clinic\Shared\Domain\EventPublisher;
use Clinic\Shared\Domain\PublishableMessage;
use Doctrine\DBAL\Connection;
use Illuminate\Console\Command;
use Throwable;
use Doctrine\ORM\EntityManagerInterface;

final class PublishOutboxEvents extends Command
{
    protected $signature = 'outbox:publish {--batch=100}';

    protected $description = 'Publishes pending outbox events to Kafka';

    private const TOPIC = 'clinic.patients';

    private readonly Connection $connection;


    public function __construct(
        EntityManagerInterface $entityManager,
        private readonly EventPublisher $publisher,
    ) {
        parent::__construct();
        $this->connection = $entityManager->getConnection();

    }

    public function handle(): int
    {
        $batchSize = (int) $this->option('batch');

        $this->connection->beginTransaction();

        try {
            $events = $this->fetchPending($batchSize);

            if ($events === []) {
                $this->connection->rollBack();

                return self::SUCCESS;
            }

            $messages = array_map(
                fn (array $event): PublishableMessage => new PublishableMessage(
                    topic: self::TOPIC,
                    key: $event['aggregate_id'],
                    payload: $event['payload'],
                    headers: ['event_type' => $event['event_name']],
                ),
                $events,
            );

            $this->publisher->publish($messages);

            $ids = array_column($events, 'id');
            $this->markPublished($ids);

            $this->connection->commit();

            $this->info(sprintf('Published %d events', count($events)));

            return self::SUCCESS;
        } catch (Throwable $exception) {
            $this->connection->rollBack();
            $this->recordFailure($exception);
            $this->error('Publishing failed: ' . $exception->getMessage());

            return self::FAILURE;
        }
    }

    /** @return list<array<string, mixed>> */
    private function fetchPending(int $limit): array
    {
        return $this->connection->fetchAllAssociative(
            'SELECT id, aggregate_id, event_name, payload
             FROM outbox_events
             WHERE published_at IS NULL
             ORDER BY id
             LIMIT :limit
             FOR UPDATE SKIP LOCKED',
            ['limit' => $limit],
        );
    }

    /** @param list<int> $ids */
    private function markPublished(array $ids): void
    {
        $this->connection->executeStatement(
            'UPDATE outbox_events SET published_at = now() WHERE id = ANY(:ids)',
            ['ids' => '{' . implode(',', $ids) . '}'],
        );
    }

    private function recordFailure(Throwable $exception): void
    {
        $this->connection->executeStatement(
            'UPDATE outbox_events
             SET attempts = attempts + 1, last_error = :error
             WHERE published_at IS NULL',
            ['error' => $exception->getMessage()],
        );
    }
}