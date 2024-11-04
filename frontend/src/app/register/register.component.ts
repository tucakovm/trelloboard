import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { HttpHeaders } from '@angular/common/http';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  registerForm: FormGroup;

  constructor(private fb: FormBuilder, private http: HttpClient) {
    this.registerForm = this.fb.group({
      first_name: ['', Validators.required],
      last_name: ['', Validators.required],
      username: ['', Validators.required],
      email: ['', [Validators.required, Validators.email]],
    });
  }

  onSubmit() {
    if (this.registerForm.valid) {
      const formData = this.registerForm.value;
      console.log('Form Data:', formData);

      const headers = new HttpHeaders({
        'Content-Type': 'application/json',
      });

      this.http
        .post('http://localhost:8080/register', formData, { headers })
        .subscribe(
          (response) => {
            console.log('Registration successful', response);
          },
          (error) => {
            console.error('Registration failed', error);
          }
        );
    } else {
      console.error('Form is invalid');
    }
  }
}
