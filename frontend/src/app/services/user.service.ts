import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})

export class UserService {


    private baseUrl = 'http://localhost:8003/user';

    constructor(private http: HttpClient) {}
  
    getUserByUsername(username: string): Observable<any> {
      const url = `${this.baseUrl}/${username}`;
      return this.http.get<any>(url);
    }
    deleteUserByUsername(username: string):Observable<any>{
        const url = `${this.baseUrl}/${username}`;
        return this.http.delete<any>(url)
    }
}