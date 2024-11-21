import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  selector: 'app-verify',
  templateUrl: './verify.component.html',
  styleUrls: ['./verify.component.css'],
})
export class VerifyComponent implements OnInit {
  verifyForm: FormGroup;
  username: string | null = null;

  constructor(
    private fb: FormBuilder,
    private http: HttpClient,
    private route: ActivatedRoute,
    private router: Router
  ) {
    this.verifyForm = this.fb.group({
      code: [
        '',
        [Validators.required, Validators.minLength(6), Validators.maxLength(6)],
      ],
    });
  }

  ngOnInit(): void {
    this.username = this.route.snapshot.paramMap.get('username');
  }

  onSubmit() {
    if (this.verifyForm.valid) {
      const formData = this.verifyForm.value;
      const headers = new HttpHeaders({
        'Content-Type': 'application/json',
      });

      const requestData = {
        ...formData,
        username: this.username,
      };

      this.http
        .post('http://localhost:8000/api/users/verify', requestData, { headers })
        .subscribe(
          (response) => {
            console.log('Verification successful', response);
            this.verifyForm.reset();
            this.router.navigate(['/login']);
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
