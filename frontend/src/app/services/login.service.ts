import { Injectable } from '@angular/core';
import { HttpHeaders, HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class LoginService {
  private apiUrl = 'http://localhost:8000/api/users/login';

  constructor(private http: HttpClient) {}

  login(user: any): Observable<any> {
    const loginHeaders = new HttpHeaders({
      Accept: 'application/json',
      'Content-Type': 'application/json',
    });
    return this.http.post(this.apiUrl, user, { headers: loginHeaders });
  }
}
