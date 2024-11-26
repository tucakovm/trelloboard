import { Component, OnInit } from '@angular/core';
import { User } from '../model/user';
import { AuthService } from '../services/auth.service';
import { UserService } from '../services/user.service';
import { Modal } from 'bootstrap';
import { Router } from '@angular/router';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css'],
})
export class ProfileComponent implements OnInit {
  user: User = {
    id: 0,
    username: '',
    password: '',
    firstname: '',
    lastname: '',
    email: '',
    role: '',
  };
  errorMessage: string = '';
  showChangePassword = false;
  passwordForm = {
    currentPassword: '',
    newPassword: '',
    repeatNewPassword: '',
  };

  constructor(
    private authService: AuthService,
    private userService: UserService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.getUser();
  }

  getUser() {
    const username = this.authService.getUserName();
    this.userService.getUserByUsername(username).subscribe(
      (response) => {
        this.user = response.user;
        console.log(this.user);
      },
      (error) => {
        console.error('Error fetching user by username', error);
      }
    );
  }

  toggleChangePassword() {
    this.showChangePassword = !this.showChangePassword;
  }

  validatePassword(password: string): boolean {
    const passwordRegex =
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/;
    return passwordRegex.test(password);
  }

  changePassword() {
    const { currentPassword, newPassword, repeatNewPassword } =
      this.passwordForm;

    if (!this.validatePassword(newPassword)) {
      alert(
        'Password must include at least one uppercase letter, one lowercase letter, one number, one special character, and must be at least 8 characters long.'
      );
      return;
    }

    if (newPassword !== repeatNewPassword) {
      alert('New passwords do not match!');
      return;
    }

    this.userService
      .changePassword(this.user.username, currentPassword, newPassword)
      .subscribe(
        (response) => {
          console.log('Password changed successfully', response);
          alert('Password changed successfully!');
          this.showChangePassword = false;
          this.passwordForm = {
            currentPassword: '',
            newPassword: '',
            repeatNewPassword: '',
          };
        },
        (error) => {
          console.error('Error changing password', error);
          alert('Failed to change password. Please try again.');
        }
      );
  }

  deleteProfile() {
    if (confirm('Are you sure you want to delete your profile?')) {
      const username = this.authService.getUserName();
      this.userService.deleteUserByUsername(username).subscribe(
        (response) => {
          console.log('Profile deleted successfully', response);
          this.authService.logout();
          this.router.navigate(['/login']);
        },
        (error) => {
          console.error('Error deleting profile', error);
          if (this.authService.getUserRoles() == 'Manager') {
            this.errorMessage =
              "You can't delete your account, you have projects that you are in charge of.";
          } else {
            this.errorMessage =
              "You can't delete your account, you have unfinished work on some project.";
          }
          const modalElement = document.getElementById('errorModal');
          if (modalElement) {
            const modal = new Modal(modalElement);
            modal.show();
          }
        }
      );
    } else {
      console.log('Profile deletion cancelled');
    }
  }
}
