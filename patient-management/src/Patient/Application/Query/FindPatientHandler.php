<?php

declare(strict_types=1);

namespace Clinic\Patient\Application\Query;

use Clinic\Patient\Domain\Model\PatientRepository;
use Clinic\Patient\Domain\ValueObject\PatientId;
use RuntimeException;

final readonly class FindPatientHandler
{
    public function __construct(private PatientRepository $patients)
    {
    }

    /** @return array<string, mixed> */
    public function __invoke(string $patientId): array
    {
        $patient = $this->patients->findById(PatientId::fromString($patientId));

        if ($patient === null) {
            throw new RuntimeException('Patient not found');
        }

        return [
            'id' => $patient->id()->value(),
            'firstName' => $patient->name()->firstName(),
            'firstSurname' => $patient->name()->firstSurname(),
            'secondSurname' => $patient->name()->secondSurname(),
            'fullName' => $patient->name()->full(),
            'phone' => $patient->phone()->value(),
            'email' => $patient->email(),
            'active' => $patient->isActive(),
            'registeredAt' => $patient->registeredAt()->format(DATE_ATOM),
        ];
    }
}