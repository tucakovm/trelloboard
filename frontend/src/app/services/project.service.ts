import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { Project } from "../model/project";

@Injectable({
    providedIn: 'root'
  })
export class ProjectService{
    private apiUrl = "http://localhost:8000/api" 
    constructor(private http:HttpClient){

    }

    createProject(project: Project): Observable<Project> {
        return this.http.post<Project>(this.apiUrl+"/projects", project, {
          headers: new HttpHeaders({ 'Content-Type': 'application/json' })
        });
      }
}