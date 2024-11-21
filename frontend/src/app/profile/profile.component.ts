import { Component, OnInit } from '@angular/core';
import { User } from '../model/user';
import { AuthService } from '../services/auth.service';
import { UserService } from '../services/user.service';

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

  showChangePassword = false;
  passwordForm = {
    currentPassword: '',
    newPassword: '',
    repeatNewPassword: ''
  };

  constructor(
    private authService: AuthService,
    private userService: UserService
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

  changePassword() {
    if (this.passwordForm.newPassword !== this.passwordForm.repeatNewPassword) {
      alert("New passwords do not match!");
      return;
    }

    this.userService.changePassword(this.user.username, this.passwordForm.currentPassword, this.passwordForm.newPassword).subscribe(
      (response) => {
        console.log('Password changed successfully', response);
        alert("Password changed successfully!");
        this.showChangePassword = false;
        this.passwordForm = { currentPassword: '', newPassword: '', repeatNewPassword: '' };
      },
      (error) => {
        console.error('Error changing password', error);
        alert("Failed to change password. Please try again.");
      }
    );
  }

  deleteProfile() {
    if (confirm('Are you sure you want to delete your profile?')) {
      const username = this.authService.getUserName();
      this.userService.deleteUserByUsername(username).subscribe(
        (response) => {
          console.log('Profile deleted successfully', response);
        },
        (error) => {
          console.error('Error deleting profile', error);
        }
      );
    } else {
      console.log('Profile deletion cancelled');
    }
  }
}
