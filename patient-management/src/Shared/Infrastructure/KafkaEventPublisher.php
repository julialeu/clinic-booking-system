<?php

declare(strict_types=1);

namespace Clinic\Shared\Infrastructure;

use Clinic\Shared\Domain\EventPublisher;
use longlang\phpkafka\Producer\Producer;
use longlang\phpkafka\Producer\ProducerConfig;

final class KafkaEventPublisher implements EventPublisher
{
    private ?Producer $producer = null;

    /** @param list<string> $brokers */
    public function __construct(private readonly array $brokers)
    {
    }

    public function publish(array $messages): void
    {
        if ($messages === []) {
            return;
        }

        $producer = $this->producer();

        foreach ($messages as $message) {
            $producer->send(
                $message->topic,
                $message->payload,
                $message->key,
                $message->headers,
            );
        }
    }

    private function producer(): Producer
    {
        if ($this->producer === null) {
            $config = new ProducerConfig();
            $config->setBootstrapServer(implode(',', $this->brokers));
            $config->setUpdateBrokers(true);
            $config->setAcks(-1);

            $this->producer = new Producer($config);
        }

        return $this->producer;
    }
}