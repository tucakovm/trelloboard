import { Injectable } from '@angular/core';
import {HttpClient, HttpHeaders} from '@angular/common/http';
import { Observable } from 'rxjs';
import {catchError, tap} from "rxjs/operators";
import {AuthService} from "./auth.service";

@Injectable({
  providedIn: 'root'
})
export class WorkflowService {
  private baseUrl = 'https://localhost:8000/api/workflows';

  constructor(private http: HttpClient, private authService: AuthService) {}

  getWorkflowByProjectId(projectId: string): Observable<any> {
    console.log('Fetching workflow for project ID:', projectId); // Log pred slanje zahteva

    return this.http.get<any>(`${this.baseUrl}/${projectId}`).pipe(
      tap(response => console.log('Received workflow response:', response)), // Log nakon odgovora
      catchError(error => {
        console.error('Error fetching workflow:', error); // Log ako dođe do greške
        throw error;
      })
    );
  }
  createWorkflow(workflow: { project_id: string; project_name: string }): Observable<any> {
    console.log('Creating workflow with data:', workflow); // Log pred slanje zahteva

    return this.http.post<any>(`${this.baseUrl}/create`, workflow).pipe(
      tap(response => console.log('Received create workflow response:', response)), // Log nakon odgovora
      catchError(error => {
        console.error('Error creating workflow:', error); // Log ako dođe do greške
        throw error;
      })
    );
  }

  addTaskToWorkflow(projectId: string, task: any): Observable<any> {
    return this.http.post<any>(`${this.baseUrl}/addtask`, task,  {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + this.authService.getToken(),
      }),
    });
  }

  addTask(projectId: string, task: TaskFW): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/addtask`, {
      project_id: projectId,
      task: task,
    },  {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + this.authService.getToken(),
      }),
    });
  }



}


interface TaskFW {
  id: string;
  name: string;
  description: string;
  dependencies: string[];
  blocked: boolean;
}
