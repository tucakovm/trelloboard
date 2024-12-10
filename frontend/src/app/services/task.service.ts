import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpResponse } from '@angular/common/http';
import { map, Observable, of, BehaviorSubject } from 'rxjs';
import { Task } from '../model/task';
import { Project } from '../model/project';
import { catchError, tap } from 'rxjs/operators';
import { UserFP } from '../model/userForProject';
@Injectable({
  providedIn: 'root',
})
export class TaskService {
  private apiUrl = 'https://localhost:8000/api';
  constructor(private http: HttpClient) {}

  private taskSubject = new BehaviorSubject<Task | null>(null); // BehaviorSubject za emitovanje podataka
  task$: Observable<Task | null> = this.taskSubject.asObservable(); // Observable za pretplatu

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
        const task = new Task(
          response.task.id,
          response.task.name,
          response.task.description,
          response.task.status,
          response.task.project_id,
          response.task.members || []
        );
        this.taskSubject.next(task); // Emitovanje nove vrednosti zadatka
        return task;
      }),
      catchError((error) => {
        console.error('Error fetching task:', error);
        return of(null);
      })
    );
  }

  // Metoda za dodavanje člana u zadatak
  AddMemberToTask(id: string, user: UserFP, timeout: number): Observable<any> {
    const headers = new HttpHeaders({
      Timeout: timeout.toString(),
    });
    return this.http
      .put<any>(`${this.apiUrl}/task/${id}/members`, user, { headers })
      .pipe(
        map((response) => {
          console.log('Member added response:', response);
          // Osvežavanje zadatka nakon dodavanja člana
          this.getById(id).subscribe();
        }),
        catchError((error) => {
          console.error('Error adding member:', error);
          return of(null);
        })
      );
  }

  removeMemberFromTask(taskId: string, userId: string): Observable<any> {
    return this.http.delete<any>(
      `${this.apiUrl}/task/${taskId}/members/${userId}`
    );
  }
  updateTask(id: string, task: Task): Observable<Task> {
    return this.http.put<Task>(`${this.apiUrl}/tasks/${id}`, task);
  }

  //files stuff
  uploadFile(taskId: string, fileData: FormData): Observable<any> {
    return this.http.post(`${this.apiUrl}/tasks/${taskId}/files`, fileData);
  }

  getFiles(taskId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tasks/${taskId}/files`);
  }

  downloadFile(fileId: string): Observable<HttpResponse<Blob>> {
    return this.http.get(`${this.apiUrl}/tasks/files/download/${fileId}`, {
      responseType: 'blob',
      observe: 'response', // Observe the full response, including headers
    });
  }
}
