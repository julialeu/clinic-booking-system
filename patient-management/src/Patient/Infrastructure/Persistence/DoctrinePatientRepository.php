<?php

declare(strict_types=1);

namespace Clinic\Patient\Infrastructure\Persistence;

use Clinic\Patient\Domain\Model\Patient;
use Clinic\Patient\Domain\Model\PatientRepository;
use Clinic\Patient\Domain\ValueObject\PatientId;
use Clinic\Patient\Domain\ValueObject\PhoneNumber;
use Doctrine\ORM\EntityManagerInterface;

final readonly class DoctrinePatientRepository implements PatientRepository
{
    public function __construct(private EntityManagerInterface $entityManager)
    {
    }

    /**
     * No se hace flush aquí: la unidad de trabajo la controla el
     * caso de uso, que sabe cuándo la operación está completa.
     */
    public function save(Patient $patient): void
    {
        $this->entityManager->persist($patient);
    }

    public function findById(PatientId $id): ?Patient
    {
        return $this->entityManager->find(Patient::class, $id->value());
    }

    public function findByPhone(PhoneNumber $phone): ?Patient
    {
        return $this->entityManager
            ->createQueryBuilder()
            ->select('p')
            ->from(Patient::class, 'p')
            ->where('p.phone = :phone')
            ->setParameter('phone', $phone->value())
            ->getQuery()
            ->getOneOrNullResult();
    }

    public function searchByName(string $term, int $limit = 20): array
    {
        $pattern = '%' . mb_strtolower(trim($term)) . '%';

        return $this->entityManager
            ->createQueryBuilder()
            ->select('p')
            ->from(Patient::class, 'p')
            ->where('LOWER(p.name.firstName) LIKE :term')
            ->orWhere('LOWER(p.name.firstSurname) LIKE :term')
            ->orWhere('LOWER(p.name.secondSurname) LIKE :term')
            ->setParameter('term', $pattern)
            ->setMaxResults($limit)
            ->getQuery()
            ->getResult();
    }
}