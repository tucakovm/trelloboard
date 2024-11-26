import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Observable } from "rxjs";
import { Task } from "../model/task";
@Injectable({
  providedIn: 'root'
})
export class TaskService {

  private apiUrl = "https://localhost:8000/api"
  constructor(private http:HttpClient){

    }


    createTask(task: Task): Observable<Task> {
      return this.http.post<Task>(this.apiUrl+"/task", task, {
        headers: new HttpHeaders({ 'Content-Type': 'application/json' })
      });
    }

    deleteTasksByProjectId(id:string): Observable<any>{
    return this.http.delete<any>(`${this.apiUrl}/task/${id}`)
    }

    getAllTasksByProjectId(id:string):Observable<any>{
      console.log("pozvan task service")
      return this.http.get<any>(`${this.apiUrl}/tasks/${id}`)
    }


}
