import { Component, OnInit } from '@angular/core';
import { User } from '../model/user';
import { AuthService } from '../services/auth.service';
import { UserService } from '../services/user.service';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css'
})


export class ProfileComponent implements OnInit{
  constructor(private authService:AuthService,private userService:UserService){}
  user: User = {
    id: 0,
    username: '',
    password: '',
    firstName: '',
    lastName: '',
    email: ''
  };

  ngOnInit(): void {
    this.getUser()
}

  getUser() {
    const username = this.authService.getUserName();
    this.userService.getUserByUsername(username).subscribe(
      (response) => {
        this.user = response;
      },
      (error) => {
        console.error('Error fetching user by username', error);
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