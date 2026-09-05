<?php

declare(strict_types=1);

namespace Clinic\Patient\Application\Query;

use Clinic\Patient\Domain\Model\Patient;
use Clinic\Patient\Domain\Model\PatientRepository;

final readonly class SearchPatientsHandler
{
    public function __construct(private PatientRepository $patients)
    {
    }

    /** @return list<array<string, mixed>> */
    public function __invoke(string $term, int $limit = 20): array
    {
        return array_map(
            static fn (Patient $patient): array => [
                'id' => $patient->id()->value(),
                'fullName' => $patient->name()->full(),
                'phone' => $patient->phone()->value(),
                'active' => $patient->isActive(),
            ],
            $this->patients->searchByName($term, $limit),
        );
    }
}