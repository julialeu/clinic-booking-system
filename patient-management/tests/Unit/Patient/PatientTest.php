<?php

declare(strict_types=1);

use Clinic\Patient\Domain\Event\PatientContactChanged;
use Clinic\Patient\Domain\Event\PatientRegistered;
use Clinic\Patient\Domain\Model\Patient;
use Clinic\Patient\Domain\ValueObject\FullName;
use Clinic\Patient\Domain\ValueObject\PhoneNumber;

function registerTestPatient(?DateTimeImmutable $now = null): Patient
{
    return Patient::register(
        FullName::create('María', 'García', 'López'),
        PhoneNumber::fromString('+34600111222'),
        $now ?? new DateTimeImmutable('2026-09-01 10:00:00'),
    );
}

test('a new patient is active', function (): void {
    expect(registerTestPatient()->isActive())->toBeTrue();
});

test('registering emits a domain event', function (): void {
    $patient = registerTestPatient();

    $events = $patient->pullDomainEvents();

    expect($events)->toHaveCount(1)
        ->and($events[0])->toBeInstanceOf(PatientRegistered::class)
        ->and($events[0]->payload()['firstName'])->toBe('María');
});

test('pulling events clears them', function (): void {
    $patient = registerTestPatient();

    $patient->pullDomainEvents();

    expect($patient->pullDomainEvents())->toBeEmpty();
});

test('changing the phone emits a contact event', function (): void {
    $patient = registerTestPatient();
    $patient->pullDomainEvents();

    $patient->changePhone(
        PhoneNumber::fromString('+34699888777'),
        new DateTimeImmutable('2026-09-02 12:00:00'),
    );

    $events = $patient->pullDomainEvents();

    expect($events)->toHaveCount(1)
        ->and($events[0])->toBeInstanceOf(PatientContactChanged::class)
        ->and($events[0]->payload()['phone'])->toBe('+34699888777');
});

test('records clinical history', function (): void {
    $patient = registerTestPatient();

    $record = $patient->addClinicalRecord(
        'Dolor lumbar',
        'Masaje descontracturante',
        new DateTimeImmutable('2026-09-05 09:00:00'),
    );

    expect($patient->clinicalRecords())->toHaveCount(1)
        ->and($record->consultationReason())->toBe('Dolor lumbar')
        ->and($record->evolution())->toBeNull();
});

test('rejects clinical records with an empty reason', function (): void {
    $patient = registerTestPatient();

    expect(fn () => $patient->addClinicalRecord('   ', 'Tratamiento', new DateTimeImmutable()))
        ->toThrow(InvalidArgumentException::class);
});

test('a discharged patient cannot receive clinical records', function (): void {
    $patient = registerTestPatient();
    $patient->discharge(new DateTimeImmutable('2026-09-10 10:00:00'));

    expect(fn () => $patient->addClinicalRecord('Dolor', 'Tratamiento', new DateTimeImmutable()))
        ->toThrow(DomainException::class);
});

test('cannot discharge twice', function (): void {
    $patient = registerTestPatient();
    $patient->discharge(new DateTimeImmutable('2026-09-10 10:00:00'));

    expect(fn () => $patient->discharge(new DateTimeImmutable('2026-09-11 10:00:00')))
        ->toThrow(DomainException::class);
});

test('evolution can be added after the session', function (): void {
    $patient = registerTestPatient();
    $record = $patient->addClinicalRecord('Dolor', 'Masaje', new DateTimeImmutable());

    $record->recordEvolution('Mejora notable de la movilidad');

    expect($record->evolution())->toBe('Mejora notable de la movilidad');
});