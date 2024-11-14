import { AbstractControl, ValidationErrors } from '@angular/forms';

export function passwordMatchValidator(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password')?.value;
  const repeatPassword = control.get('repeatPassword')?.value;

  // Return error if passwords do not match
  return password && repeatPassword && password !== repeatPassword
    ? { mismatch: true }
    : null; // No error if passwords match
}
