import { Injectable } from '@angular/core';
import { HttpHeaders } from '@angular/common/http';
import { JwtHelperService } from '@auth0/angular-jwt';

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  constructor(private jwtHelper: JwtHelperService) { }

  tokenIsPresent() {
    if (typeof window !== 'undefined' && window.localStorage) {
      let token = localStorage.getItem('jwt');
      return token != null && token != undefined;
    }
    return false;
  }
  
  

  getToken() {
    const token = localStorage.getItem('jwt');
    return token;
  }
  getDecodedToken() {
    if (typeof window !== 'undefined' && window.localStorage) {
      const token = localStorage.getItem('jwt');
      if (token) {
        return this.jwtHelper.decodeToken(token);
      }
    }
    return null;
  }

  getUserRoles() {
    const decodedToken = this.getDecodedToken();
    if (decodedToken && decodedToken.user_role) {
      return decodedToken.user_role;
    }
    return null;
  }

  getUserName() {
    const decodedToken = this.getDecodedToken();
    if (decodedToken && decodedToken.username) {
      return decodedToken.username;
    }
    return null;
  }

  isLoggedIn(): boolean {
    let result = this.tokenIsPresent();
    return result;
  }

  isManager(): boolean{
    if (this.isLoggedIn()) {
      let roles = this.getUserRoles();
      return roles && roles.includes('Manager');
    }
    return false;
  }

  isUser(): boolean{
    if (this.isLoggedIn()) {
      let roles = this.getUserRoles();
      return roles && roles.includes('User');
    }
    return false;
  }

  logout() {
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('jwt');
    }
  }

  get headers():HttpHeaders{
    const token = localStorage.getItem('jwt');
    const loginHeaders = new HttpHeaders({
      'Accept': 'application/json',
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
    return loginHeaders;
  }

}
