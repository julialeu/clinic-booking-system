<?php

declare(strict_types=1);

namespace Clinic\Patient\Domain\Model;

use DateTimeImmutable;
use Doctrine\ORM\Mapping as ORM;
use InvalidArgumentException;

#[ORM\Entity]
#[ORM\Table(name: 'clinical_records')]
class ClinicalRecord
{
    #[ORM\Id]
    #[ORM\Column(type: 'string', length: 36)]
    private string $id;

    #[ORM\ManyToOne(targetEntity: Patient::class, inversedBy: 'clinicalRecords')]
    #[ORM\JoinColumn(name: 'patient_id', referencedColumnName: 'id', nullable: false)]
    private Patient $patient;

    #[ORM\Column(name: 'appointment_id', type: 'string', length: 36, nullable: true)]
    private ?string $appointmentId;

    #[ORM\Column(name: 'consultation_reason', type: 'text')]
    private string $consultationReason;

    #[ORM\Column(name: 'treatment', type: 'text')]
    private string $treatment;

    #[ORM\Column(name: 'evolution', type: 'text', nullable: true)]
    private ?string $evolution;

    #[ORM\Column(name: 'recorded_at', type: 'datetimetz_immutable')]
    private DateTimeImmutable $recordedAt;

    public function __construct(
        string $id,
        Patient $patient,
        string $consultationReason,
        string $treatment,
        DateTimeImmutable $recordedAt,
        ?string $appointmentId = null,
        ?string $evolution = null,
    ) {
        if (trim($consultationReason) === '') {
            throw new InvalidArgumentException('Consultation reason cannot be empty');
        }

        if (trim($treatment) === '') {
            throw new InvalidArgumentException('Treatment cannot be empty');
        }

        $this->id = $id;
        $this->patient = $patient;
        $this->consultationReason = trim($consultationReason);
        $this->treatment = trim($treatment);
        $this->recordedAt = $recordedAt;
        $this->appointmentId = $appointmentId;
        $this->evolution = $evolution;
    }

    public function id(): string
    {
        return $this->id;
    }

    public function consultationReason(): string
    {
        return $this->consultationReason;
    }

    public function treatment(): string
    {
        return $this->treatment;
    }

    public function evolution(): ?string
    {
        return $this->evolution;
    }

    public function recordedAt(): DateTimeImmutable
    {
        return $this->recordedAt;
    }

    public function appointmentId(): ?string
    {
        return $this->appointmentId;
    }

    /**
     * La evolución se añade después de la sesión, cuando el
     * fisioterapeuta valora la respuesta al tratamiento.
     */
    public function recordEvolution(string $evolution): void
    {
        if (trim($evolution) === '') {
            throw new InvalidArgumentException('Evolution cannot be empty');
        }

        $this->evolution = trim($evolution);
    }
}