import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { map, Observable, of } from 'rxjs';
import { Task } from '../model/task';
import { Project } from '../model/project';
import { catchError } from 'rxjs/operators';
import { UserFP } from '../model/userForProject';
@Injectable({
  providedIn: 'root',
})
export class TaskService {
  private apiUrl = 'https://localhost:8000/api';
  constructor(private http: HttpClient) {}

  createTask(task: Task): Observable<Task> {
    return this.http.post<Task>(this.apiUrl + '/task', task, {
      headers: new HttpHeaders({ 'Content-Type': 'application/json' }),
    });
  }

  deleteTasksByProjectId(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/task/${id}`);
  }

  getAllTasksByProjectId(id: string): Observable<any> {
    console.log('pozvan task service');
    return this.http.get<any>(`${this.apiUrl}/tasks/${id}`);
  }

  getById(id: string): Observable<Task | null> {
    return this.http.get<any>(`${this.apiUrl}/task/${id}`).pipe(
      map((response: any) => {
        console.log('API Response:', response);
        const item = response.task;
        console.log('Mapped task members:', item.users); // Proveri podatke

        return new Task(
          item.id,
          item.name,
          item.description,
          item.status,
          item.project_id,
          item.members && Array.isArray(item.members)
            ? item.members.map((user: any) => ({
                id: user.id,
                username: user.username,
                role: user.role,
              }))
            : [] // Ako nema korisnika, vraća prazan niz
        );
      }),
      catchError((error) => {
        console.error('Error fetching task:', error);
        return of(null);
      })
    );
  }

  AddMemberToTask(id: string, member: UserFP, timeout: number) {
    const headers = new HttpHeaders({
      'Timeout': timeout.toString() // Dodavanje timeout-a u header
    });

    return this.http.put<any>(`${this.apiUrl}/task/${id}/members`, member, { headers });
  }


  removeMemberFromTask(taskId: string, userId: string): Observable<any> {
    return this.http.delete<any>(
      `${this.apiUrl}/task/${taskId}/members/${userId}`
    );
  }
  updateTask(id: string, task: Task): Observable<Task> {
    return this.http.put<Task>(`${this.apiUrl}/tasks/${id}`, task);
  }
}
