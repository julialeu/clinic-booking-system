<?php

declare(strict_types=1);

namespace Clinic\Shared\Domain;

final readonly class PublishableMessage
{
    /** @param array<string, string> $headers */
    public function __construct(
        public string $topic,
        public string $key,
        public string $payload,
        public array $headers = [],
    ) {
    }
}

interface EventPublisher
{
    /** @param list<PublishableMessage> $messages */
    public function publish(array $messages): void;
}