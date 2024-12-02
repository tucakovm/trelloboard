import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { AuthService } from "./auth.service";
import { Observable, of } from "rxjs";
import { catchError, map } from "rxjs/operators";
import { Notification } from "../model/notification";

@Injectable({
  providedIn: 'root'
})
export class NotService {
  private apiUrl = "https://localhost:8000/api";

  constructor(private http: HttpClient, private authService: AuthService) {}

  getAllNots(username: string): Observable<Notification[]> {
    return this.http.get<any>(`${this.apiUrl}/notifications/${username}`).pipe(
      map((response: any) => {
        console.log('API Response:', response); // Verify response structure
        // Adjust the mapping to access `nots` array in the response
        if (response.nots && Array.isArray(response.nots)) {
          return response.nots.map((item: any) => {
            return new Notification(
              item.notId, // Map to notId
              this.timestampToDate(item.createdAt), // Convert timestamp to Date
              item.userId, // Map to userId
              item.message, // Map to message
              item.status !== 'unread' // Map status to boolean
            );
          });
        } else {
          return []; // Return empty array if no notifications found
        }
      }),
      catchError((error) => {
        console.error('Error fetching notifications:', error);
        return of([]); // Fallback to an empty array on error
      })
    );
  }

  timestampToDate(timestamp: string): Date {
    // Convert ISO string timestamp to Date object
    return new Date(timestamp);
  }
}
