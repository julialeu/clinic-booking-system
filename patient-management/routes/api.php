<?php

declare(strict_types=1);

use App\Http\Controllers\PatientController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1')->group(function (): void {
    Route::get('patients', [PatientController::class, 'index']);
    Route::post('patients', [PatientController::class, 'store']);
    Route::get('patients/{id}', [PatientController::class, 'show']);
});