<?php

namespace App\Providers;

use Illuminate\Support\ServiceProvider;
use Clinic\Patient\Domain\Model\PatientRepository;
use Clinic\Patient\Infrastructure\Persistence\DoctrinePatientRepository;
use Clinic\Shared\Domain\Clock;
use Clinic\Shared\Infrastructure\SystemClock;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
         $this->app->bind(PatientRepository::class, DoctrinePatientRepository::class);
         $this->app->bind(Clock::class, SystemClock::class);
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
