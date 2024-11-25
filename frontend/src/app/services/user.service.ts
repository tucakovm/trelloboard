import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class UserService {
  private baseUrl = 'http://localhost:8000/api/users';

  constructor(private http: HttpClient) {}

  getUserByUsername(username: string): Observable<any> {
    const url = `${this.baseUrl}/${username}`;
    return this.http.get<any>(url);
  }
  deleteUserByUsername(username: string): Observable<any> {
    const url = `${this.baseUrl}/${username}`;
    return this.http.delete<any>(url);
  }
  changePassword(
    username: string,
    currentPassword: string,
    newPassword: string
  ): Observable<any> {
    const url = `${this.baseUrl}/change-password`;
    const body = {
      currentPassword: currentPassword,
      newPassword: newPassword,
      userName: username,
    };
    return this.http.put(url, body);
  }

  requestMagicLink(email: string) {
    const apiUrl = `${this.baseUrl}/magic-link`;
    const headers = { 'Content-Type': 'application/json' };
    return this.http.post(apiUrl, { email }, { headers });
  }
}
