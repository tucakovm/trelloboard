import { Component } from '@angular/core';
import { AuthService } from './services/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent {
  constructor(private authService:AuthService, private router:Router){}
  title = 'frontend';

  isManager(){
    return this.authService.isManager();
  }
  isUser(){
    return this.authService.isUser();
  }
  isLoggedIn(){
    return this.authService.isLoggedIn();
  }
  logOut(){
    this.router.navigate(['/login'])
    return this.authService.logout();
  }

}
