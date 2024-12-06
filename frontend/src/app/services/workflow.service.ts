import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {Workflow} from "../model/workflow";
import {TaskFW} from "../model/TaskFW";


@Injectable({
  providedIn: 'root'
})
export class WorkflowService {
  private apiUrl = 'https://localhost:8000/api/workflows';

  constructor(private http: HttpClient) {}

  getWorkflow(projectID: string): Observable<Workflow> {
    return this.http.get<Workflow>(`${this.apiUrl}/${projectID}`);
  }

  createWorkflow(request: { project_id: string | null; project_name: string}): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/create`, request);
  }

  addTask(request: { task: TaskFW; project_id: string | null }): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/addtask`, request);
  }
}
