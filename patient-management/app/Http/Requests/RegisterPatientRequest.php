<?php

declare(strict_types=1);

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

final class RegisterPatientRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    /** @return array<string, mixed> */
    public function rules(): array
    {
        return [
            'firstName' => ['required', 'string', 'max:80'],
            'firstSurname' => ['required', 'string', 'max:80'],
            'secondSurname' => ['nullable', 'string', 'max:80'],
            'phone' => ['required', 'string', 'max:20'],
            'email' => ['nullable', 'email', 'max:180'],
        ];
    }
}