<?php

declare(strict_types=1);

namespace Clinic\Shared\Infrastructure;

use Clinic\Shared\Domain\Clock;
use DateTimeImmutable;
use DateTimeZone;

final readonly class SystemClock implements Clock
{
    public function now(): DateTimeImmutable
    {
        return new DateTimeImmutable('now', new DateTimeZone('UTC'));
    }
}