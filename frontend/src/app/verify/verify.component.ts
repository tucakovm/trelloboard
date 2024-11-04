import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { HttpClient, HttpHeaders } from '@angular/common/http';

@Component({
  selector: 'app-verify',
  templateUrl: './verify.component.html',
  styleUrls: ['./verify.component.css'],
})
export class VerifyComponent {
  verifyForm: FormGroup;

  constructor(private fb: FormBuilder, private http: HttpClient) {
    this.verifyForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      code: [
        '',
        [Validators.required, Validators.minLength(6), Validators.maxLength(6)],
      ],
    });
  }

  onSubmit() {
    if (this.verifyForm.valid) {
      const formData = this.verifyForm.value;
      const headers = new HttpHeaders({
        'Content-Type': 'application/json',
      });

      this.http
        .post('http://localhost:8080/verify', formData, { headers })
        .subscribe(
          (response) => {
            console.log('Verification successful', response);
          },
          (error) => {
            console.error('Verification failed', error);
          }
        );
    } else {
      console.error('Form is invalid');
    }
  }
}
