import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { LoginService } from '../services/login.service';
import { LoggedUser } from '../model/LoggedUser';
import { Observer } from 'rxjs';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {
  taskForm: FormGroup;
  errorOccurred: boolean = false;
  errorMessage: string = '';

  constructor(
    private loginService: LoginService,
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService
  ) {
    this.taskForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(4)]],
      password: ['', [Validators.required, Validators.minLength(4)]],
    });
  }

  onSubmit() {
    if (this.taskForm.valid) {
      let loggedUser: LoggedUser = new LoggedUser(
        this.taskForm.value.username,
        this.taskForm.value.password
      );

      const observer: Observer<any> = {
        next: (res: any) => {
          console.log('Login success, response:', res);

          localStorage.setItem('jwt', res.token);

          const token = this.authService.getDecodedToken();
          const userRole = token?.user_role;

          if (userRole === 'Manager') {
            this.router.navigate(['/manager-dashboard']);
          } else if (userRole === 'User') {
            this.router.navigate(['/user-dashboard']);
          } else {
            console.warn('Unknown role:', userRole);
            this.errorOccurred = true;
            this.errorMessage = 'Unknown role. Please contact support.';
          }
        },
        error: (error: any) => {
          this.errorOccurred = true;
          this.errorMessage =
            'Error during login. Please check your credentials.';
          console.error('Login error:', error);
        },
        complete: () => {},
      };

      this.loginService.login(loggedUser).subscribe(observer);
    }
  }
}
