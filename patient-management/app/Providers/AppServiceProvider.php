<?php

namespace App\Providers;

use Illuminate\Support\ServiceProvider;
use Clinic\Patient\Domain\Model\PatientRepository;
use Clinic\Patient\Infrastructure\Persistence\DoctrinePatientRepository;
use Clinic\Shared\Domain\Clock;
use Clinic\Shared\Infrastructure\SystemClock;
use Clinic\Shared\Domain\EventStore;
use Clinic\Shared\Infrastructure\DoctrineEventStore;
use Clinic\Shared\Domain\EventPublisher;
use Clinic\Shared\Infrastructure\KafkaEventPublisher;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
         $this->app->bind(PatientRepository::class, DoctrinePatientRepository::class);
         $this->app->bind(Clock::class, SystemClock::class);
         $this->app->bind(EventStore::class, DoctrineEventStore::class);
         $this->app->bind(EventPublisher::class, function (): KafkaEventPublisher {
            return new KafkaEventPublisher(
                explode(',', env('KAFKA_BROKERS', 'localhost:9092'))
            );
        });
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
