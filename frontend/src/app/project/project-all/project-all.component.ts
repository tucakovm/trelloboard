import { Component, OnInit } from '@angular/core';
import { ProjectService } from '../../services/project.service';
import { Project } from '../../model/project';
import { Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-project-all',
  templateUrl: './project-all.component.html',
  styleUrl: './project-all.component.css'
})
export class ProjectAllComponent implements OnInit{
  projects?:Project[];
  constructor(private projectService:ProjectService,  private router: Router, private tasksService:TaskService , private authService:AuthService){}

  ngOnInit(): void {
    this.getAllProjects();
  }

  getAllProjects(): void {
    let username = this.authService.getUserName()
    let userId = this.authService.getUserId()
    this.projectService.getAllProjects(username, userId).subscribe( {
      next:(data) =>{
        this.projects = data;
      },
      error:(error)=>{
        this.projects = [];
        console.error("Error loading projects, projects are null!")
      }
    })
  }
}
