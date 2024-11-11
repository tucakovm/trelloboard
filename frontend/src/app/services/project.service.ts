import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Injectable } from "@angular/core";
import {map, Observable} from "rxjs";
import { Project } from "../model/project";

@Injectable({
    providedIn: 'root'
  })
export class ProjectService{
    private apiUrl = "http://localhost:8001/api"
    constructor(private http:HttpClient){}

    createProject(project: Project): Observable<Project> {
        return this.http.post<Project>(this.apiUrl+"/projects", project, {
          headers: new HttpHeaders({ 'Content-Type': 'application/json' })
        });
      }

  getAllProjects(): Observable<Project[]> {
    return this.http.get<any[]>(`${this.apiUrl}/projects`).pipe(
      map((data: any[]) => data.map(item => new Project(
        item.id,
        item.name,
        new Date(item.completionDate),
        item.minMembers,
        item.maxMembers,
        item.manager
      )))
    );
  }

    deleteProjectById(id:number): Observable<void>{
      return this.http.delete<void>(`${this.apiUrl}/projects/${id}`)
    }
}
