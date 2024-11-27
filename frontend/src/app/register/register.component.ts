import { Component } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  FormGroup,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { passwordMatchValidator } from '../validators/password-match.validator';
import { Router } from '@angular/router';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  registerForm: FormGroup;
  successMessage: string | null = null;
  errorMessage: string | null = null;
  captchaResolved: boolean = false;
  captchaToken: string = '';
  captchaResponse: string = '';

  constructor(
    private fb: FormBuilder,
    private http: HttpClient,
    private router: Router
  ) {
    this.registerForm = this.fb.group(
      {
        firstname: ['', [Validators.required, Validators.minLength(2)]],
        lastname: ['', [Validators.required, Validators.minLength(2)]],
        username: ['', [Validators.required, Validators.minLength(4)]],
        email: ['', [Validators.required, Validators.email]],
        password: [
          '',
          [
            Validators.required,
            Validators.minLength(8),
            Validators.pattern(
              /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/
            ),
          ],
        ],
        repeatPassword: ['', [Validators.required]],
        role: ['', [Validators.required]],
      },
      { validators: passwordMatchValidator }
    );
  }

  get password() {
    return this.registerForm.get('password');
  }

  get repeatPassword() {
    return this.registerForm.get('repeatPassword');
  }

  onCaptchaResolved(captchaResponse: string | null) {
    this.captchaToken = captchaResponse || '';
    this.captchaResponse = captchaResponse || '';
    this.captchaResolved = !!captchaResponse;
  }

  onSubmit() {
    if (this.registerForm.valid) {
      const formData = this.registerForm.value;
      console.log('Form Data:', formData);

      const requestBody = {
        ...formData,
        captchaResponse: this.captchaToken,
      };

      const headers = new HttpHeaders({
        'Content-Type': 'application/json',
      });

      this.http
        .post('https://localhost:8000/api/users/register', requestBody, {
          headers,
        })
        .subscribe(
          (response) => {
            this.successMessage =
              'Registration successful! Verification email sent.';
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
