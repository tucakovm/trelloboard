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
    this.projectService.getAllProjects(username).subscribe( {
      next:(data) =>{
        this.projects = data;
      },
      error:(error)=>{
        this.projects = [];
        console.error("Error loading projects, projects are null!")
      }
    })
  }

  deleteAllTasksByProjectId(id:string){
    this.tasksService.deleteTasksByProjectId(id).subscribe({
      next:(response)=>{
        console.log("Tasks deleted sucessfuly:"+response)
      },
      error:(error)=>{
        console.error("Error deleting tasks:"+ error)
      }
    })
  }

  deleteProject(id: number|null): void {
    if (id != null){
      this.projectService.deleteProjectById(id).subscribe({
        next:(response) => {
          console.log('Project deleted successfully:', response);
          this.deleteAllTasksByProjectId(id.toString());
          this.getAllProjects()
        },
        error: (error) => {
          console.error('Error deleting project:', error);
        },
      })
    }
  }
  addTask(projectId: number | null): void {
    if (projectId != null) {
      this.router.navigate(['/tasks', projectId]);
    }
  }
}
