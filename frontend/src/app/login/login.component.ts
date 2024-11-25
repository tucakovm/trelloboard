import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { LoginService } from '../services/login.service';
import { LoggedUser } from '../model/LoggedUser';
import { Observer } from 'rxjs';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { UserService } from '../services/user.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {
  taskForm: FormGroup;
  errorOccurred: boolean = false;
  errorMessage: string = '';
  captchaResolved: boolean = false;
  captchaToken: string = '';
  captchaResponse: string = '';
  showMagicLinkInput: boolean = false;

  constructor(
    private loginService: LoginService,
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService,
    private userService: UserService
  ) {
    this.taskForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(4)]],
      password: ['', [Validators.required, Validators.minLength(4)]],
      email: ['', [Validators.email]],
    });
  }

  toggleMagicLinkInput() {
    this.showMagicLinkInput = true;

    const emailControl = this.taskForm.get('email');
    if (emailControl) {
      emailControl.reset();
    }
  }

  onCaptchaResolved(captchaResponse: string | null) {
    this.captchaToken = captchaResponse || '';
    this.captchaResponse = captchaResponse || '';
    this.captchaResolved = !!captchaResponse;
  }

  onSubmit() {
    if (this.taskForm.valid && this.captchaResolved) {
      let loggedUser: LoggedUser = new LoggedUser(
        this.taskForm.value.username,
        this.taskForm.value.password
      );
      if (!this.captchaResponse) {
        console.error('Captcha not resolved.');
        return;
      }

      const loginRequest = {
        ...loggedUser,
        captchaToken: this.captchaToken,
      };

      const observer: Observer<any> = {
        next: (res: any) => {
          console.log('Login success, response:', res);

          localStorage.setItem('jwt', res.token);
          this.router.navigate(['/all-projects']);
        },
        error: (error: any) => {
          this.errorOccurred = true;
          this.errorMessage =
            'Error during login. Please check your credentials.';
          console.error('Login error:', error);
        },
        complete: () => {},
      };

      this.loginService
        .login({ ...loggedUser, key: this.captchaToken })
        .subscribe(observer);
    }
  }

  onMagicLinkRequest() {
    const email = this.taskForm.get('email')?.value;

    this.userService.requestMagicLink(email).subscribe({
      next: (res) => {
        console.log('Magic link sent successfully:', res);
        this.errorOccurred = false;
        this.errorMessage = '';
        alert('Magic link sent to your email. Please check your inbox!');
      },
      error: (err) => {
        console.error('Error sending magic link:', err);
        this.errorOccurred = true;
        this.errorMessage = 'Error sending magic link. Please try again.';
      },
    });
  }
}
