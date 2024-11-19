import { Component } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, ValidationErrors, Validators } from '@angular/forms';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { passwordMatchValidator } from "../validators/password-match.validator";
import { Router } from '@angular/router';  // Import Router for navigation

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  registerForm: FormGroup;
  successMessage: string | null = null;
  errorMessage: string | null = null;

  constructor(private fb: FormBuilder, private http: HttpClient, private router: Router) {
    this.registerForm = this.fb.group(
      {
        first_name: ['', [Validators.required, Validators.minLength(2)]],
        last_name: ['', [Validators.required, Validators.minLength(2)]],
        username: ['', [Validators.required, Validators.minLength(4)]],
        email: ['', [Validators.required, Validators.email]],
        password: ['', [Validators.required, Validators.minLength(6)]],
        repeatPassword: ['', [Validators.required, Validators.minLength(6)]],
        role: ['', [Validators.required]],
      },
      { validators: passwordMatchValidator }
    );
  }

  onSubmit() {
    if (this.registerForm.valid) {
      const formData = this.registerForm.value;
      console.log('Form Data:', formData);

      const headers = new HttpHeaders({
        'Content-Type': 'application/json',
      });

      this.http
        .post('http://localhost:8003/api/users/register', formData, { headers })
        .subscribe(
          (response) => {
            this.successMessage = 'Registration successful! Verification email sent.';
            this.errorMessage = null;
            this.registerForm.reset();
            console.log('Registration successful', response);

            this.router.navigate(['/verify', formData.username]);
          },
          (error) => {
            this.errorMessage = 'Registration failed. Please try again.';
            this.successMessage = null;
            console.error('Registration failed', error);
          }
        );
    } else {
      this.errorMessage = 'Please fill out the form correctly.';
      this.successMessage = null;
      console.error('Form is invalid');
    }
  }
}
