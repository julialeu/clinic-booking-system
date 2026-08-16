<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\Model;

use Clinic\Patient\Domain\ValueObject\FullName;
use Clinic\Patient\Domain\ValueObject\PatientId;
use Clinic\Patient\Domain\ValueObject\PhoneNumber;
use DateTimeImmutable;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;
use DomainException;
use Ramsey\Uuid\Uuid;

#[ORM\Entity]
#[ORM\Table(name: 'patients')]
class Patient
{
    #[ORM\Id]
    #[ORM\Column(type: 'string', length: 36)]
    private string $id;

    #[ORM\Embedded(class: FullName::class, columnPrefix: false)]
    private FullName $name;

    #[ORM\Column(name: 'phone', type: 'string', length: 20)]
    private string $phone;

    #[ORM\Column(name: 'email', type: 'string', length: 180, nullable: true)]
    private ?string $email;

    #[ORM\Column(name: 'registered_at', type: 'datetimetz_immutable')]
    private DateTimeImmutable $registeredAt;

    #[ORM\Column(name: 'discharged_at', type: 'datetimetz_immutable', nullable: true)]
    private ?DateTimeImmutable $dischargedAt;

    /** @var Collection<int, ClinicalRecord> */
    #[ORM\OneToMany(
        targetEntity: ClinicalRecord::class,
        mappedBy: 'patient',
        cascade: ['persist'],
        fetch: 'EXTRA_LAZY',
    )]
    #[ORM\OrderBy(['recordedAt' => 'DESC'])]
    private Collection $clinicalRecords;

    private function __construct(
        PatientId $id,
        FullName $name,
        PhoneNumber $phone,
        ?string $email,
        DateTimeImmutable $registeredAt,
    ) {
        $this->id = $id->value();
        $this->name = $name;
        $this->phone = $phone->value();
        $this->email = $email;
        $this->registeredAt = $registeredAt;
        $this->dischargedAt = null;
        $this->clinicalRecords = new ArrayCollection();
    }

    public static function register(
        FullName $name,
        PhoneNumber $phone,
        DateTimeImmutable $now,
        ?string $email = null,
    ): self {
        return new self(PatientId::generate(), $name, $phone, $email, $now);
    }

    public function id(): PatientId
    {
        return PatientId::fromString($this->id);
    }

    public function name(): FullName
    {
        return $this->name;
    }

    public function phone(): PhoneNumber
    {
        return PhoneNumber::fromString($this->phone);
    }

    public function email(): ?string
    {
        return $this->email;
    }

    public function registeredAt(): DateTimeImmutable
    {
        return $this->registeredAt;
    }

    public function isActive(): bool
    {
        return $this->dischargedAt === null;
    }

    /** @return Collection<int, ClinicalRecord> */
    public function clinicalRecords(): Collection
    {
        return $this->clinicalRecords;
    }

    public function changePhone(PhoneNumber $phone): void
    {
        $this->phone = $phone->value();
    }

    public function changeName(FullName $name): void
    {
        $this->name = $name;
    }

    /**
     * Invariante: no se registra actividad clínica sobre un paciente
     * que ya causó baja.
     */
    public function addClinicalRecord(
        string $consultationReason,
        string $treatment,
        DateTimeImmutable $recordedAt,
        ?string $appointmentId = null,
    ): ClinicalRecord {
        if (! $this->isActive()) {
            throw new DomainException('Cannot add clinical records to a discharged patient');
        }

        $record = new ClinicalRecord(
            Uuid::uuid4()->toString(),
            $this,
            $consultationReason,
            $treatment,
            $recordedAt,
            $appointmentId,
        );

        $this->clinicalRecords->add($record);

        return $record;
    }

    public function discharge(DateTimeImmutable $now): void
    {
        if (! $this->isActive()) {
            throw new DomainException('Patient is already discharged');
        }

        $this->dischargedAt = $now;
    }
}