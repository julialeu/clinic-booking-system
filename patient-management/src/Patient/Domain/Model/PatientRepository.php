<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\Model;

use Clinic\Patient\Domain\ValueObject\PatientId;
use Clinic\Patient\Domain\ValueObject\PhoneNumber;

interface PatientRepository
{
    public function save(Patient $patient): void;

    public function findById(PatientId $id): ?Patient;

    public function findByPhone(PhoneNumber $phone): ?Patient;

    /** @return list<Patient> */
    public function searchByName(string $term, int $limit = 20): array;
}