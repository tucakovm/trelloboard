import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Injectable } from "@angular/core";
import {map, Observable} from "rxjs";
import { Project } from "../model/project";
import { UserFP } from "../model/userForProject";
import { catchError } from 'rxjs/operators';
import { of } from 'rxjs';
import {AuthService} from "./auth.service";


@Injectable({
    providedIn: 'root'
  })
export class ProjectService{
    private apiUrl = "http://localhost:8000/api"
    constructor(private http:HttpClient,private authService : AuthService) { }

    createProject(project: Project): Observable<Project> {
        return this.http.post<Project>(this.apiUrl+"/project", project, {
          headers: new HttpHeaders({ 'Content-Type': 'application/json' , 'Authorization': 'Bearer ' + this.authService.getToken() })
        });
      }

      // Convert protobuf Timestamp to JavaScript Date
      timestampToDate(timestamp: any): Date {
        if (timestamp && timestamp.seconds) {
          const date = new Date(0);
          date.setUTCSeconds(timestamp.seconds);
          return date;
        }
        return new Date(); // Default to current date if invalid timestamp
      }




  getAllProjects(username: string): Observable<Project[]> {
    return this.http.get<any>(`${this.apiUrl}/projects/${username}`).pipe(
      map((response: any) => {
        console.log('API Response:', response); // Proveri strukturu odgovora
        return response.projects.map((item: ProjectItem) => new Project(
          item.id,
          item.name,
          this.timestampToDate(item.completionDate),
          item.minMembers,
          item.maxMembers,
          item.manager = {
            id: item.manager.id,
            username: item.manager.username,
            role: item.manager.role
          },
          (item.members && Array.isArray(item.members)) ?
            item.members.map((member: any) => ({
              id: member.id,
              username: member.username,
              role: member.role
            })) : [] // Ako je null ili nije niz, vraća prazan niz
        ));
      }),
      catchError((error) => {
        console.error('Error fetching projects:', error);
        return of([]); // Fallback na prazan niz
      })
    );
  }


    deleteProjectById(id:string): Observable<void>{
      return this.http.delete<void>(`${this.apiUrl}/project/${id}`)
    }

    getById(id: string): Observable<Project | null> {
      return this.http.get<any>(`${this.apiUrl}/project/${id}`).pipe(
        map((response: any) => {
          console.log('API Response:', response);
          const item = response.project;
          return new Project(
            item.id,
            item.name,
            this.timestampToDate(item.completionDate),
            item.minMembers,
            item.maxMembers,
            {
              id: item.manager.id,
              username: item.manager.username,
              role: item.manager.role
            },
            (item.members && Array.isArray(item.members)) ?
              item.members.map((member: any) => ({
                id: member.id,
                username: member.username,
                role: member.role
              })) : [] // Ako je null ili nije niz, vraća prazan niz
          );
        }),
        catchError((error) => {
          console.error('Error fetching project:', error);
          return of(null); // Fallback na null ako dođe do greške
        })
      );
    }

    createMember(id:string, member:UserFP){
      console.log("Pozvan createmember servis na frontu");
      console.log("id:" + id + " member: "+UserFP)
      return this.http.put<any>(`${this.apiUrl}/projects/${id}/members`,member)
    }

    removeMember(id:string, member:UserFP){
      return this.http.delete<any>(`${this.apiUrl}/projects/${id}/members/${member.id}`)
    }
  }

interface User {
  id: string;
  username: string;
  role: string;
}

interface ProjectItem {
  id: string;
  name: string;
  completionDate: string; // Ako je to string, koristi string ili Date, zavisno od tipa u tvojoj aplikaciji
  minMembers: number;
  maxMembers: number;
  manager: User;
  members :User[];
}
