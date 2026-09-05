<?php

declare(strict_types=1);

namespace App\Http\Controllers;

use App\Http\Requests\RegisterPatientRequest;
use Clinic\Patient\Application\Command\RegisterPatient;
use Clinic\Patient\Application\Command\RegisterPatientHandler;
use Clinic\Patient\Application\Query\FindPatientHandler;
use Clinic\Patient\Application\Query\SearchPatientsHandler;
use DomainException;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use InvalidArgumentException;
use RuntimeException;

final class PatientController extends Controller
{
    public function store(
        RegisterPatientRequest $request,
        RegisterPatientHandler $handler,
    ): JsonResponse {
        try {
            $id = $handler(new RegisterPatient(
                firstName: $request->string('firstName')->toString(),
                firstSurname: $request->string('firstSurname')->toString(),
                secondSurname: $request->input('secondSurname'),
                phone: $request->string('phone')->toString(),
                email: $request->input('email'),
            ));

            return response()->json(['id' => $id->value()], 201);
        } catch (DomainException $exception) {
            return response()->json(['error' => $exception->getMessage()], 409);
        } catch (InvalidArgumentException $exception) {
            return response()->json(['error' => $exception->getMessage()], 422);
        }
    }

    public function show(string $id, FindPatientHandler $handler): JsonResponse
    {
        try {
            return response()->json($handler($id));
        } catch (RuntimeException) {
            return response()->json(['error' => 'Patient not found'], 404);
        } catch (InvalidArgumentException) {
            return response()->json(['error' => 'Invalid patient id'], 422);
        }
    }

    public function index(Request $request, SearchPatientsHandler $handler): JsonResponse
    {
        $term = $request->string('q')->toString();

        if (trim($term) === '') {
            return response()->json(['error' => 'Query parameter q is required'], 422);
        }

        return response()->json(['data' => $handler($term)]);
    }
}