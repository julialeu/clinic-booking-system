<?php

declare(strict_types=1);

namespace Clinic\Patient\Application\Command;

use Clinic\Patient\Domain\Model\Patient;
use Clinic\Patient\Domain\Model\PatientRepository;
use Clinic\Patient\Domain\ValueObject\FullName;
use Clinic\Patient\Domain\ValueObject\PatientId;
use Clinic\Patient\Domain\ValueObject\PhoneNumber;
use Clinic\Shared\Domain\Clock;
use Doctrine\ORM\EntityManagerInterface;
use DomainException;

final readonly class RegisterPatientHandler
{
    public function __construct(
        private PatientRepository $patients,
        private EntityManagerInterface $entityManager,
        private Clock $clock,
    ) {
    }

    public function __invoke(RegisterPatient $command): PatientId
    {
        $phone = PhoneNumber::fromString($command->phone);

        if ($this->patients->findByPhone($phone) !== null) {
            throw new DomainException('A patient with this phone number already exists');
        }

        $patient = Patient::register(
            FullName::create(
                $command->firstName,
                $command->firstSurname,
                $command->secondSurname,
            ),
            $phone,
            $this->clock->now(),
            $command->email,
        );

        $this->patients->save($patient);
        $this->entityManager->flush();

        return $patient->id();
    }
}