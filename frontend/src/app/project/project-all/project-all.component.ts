import { Component, OnInit } from '@angular/core';
import { ProjectService } from '../../services/project.service';
import { Project } from '../../model/project';
import { Router } from '@angular/router';

@Component({
  selector: 'app-project-all',
  templateUrl: './project-all.component.html',
  styleUrl: './project-all.component.css'
})
export class ProjectAllComponent implements OnInit{
  projects?:Project[];
  constructor(private projectService:ProjectService,  private router: Router){}

  ngOnInit(): void {
    this.getAllProjects();
  }

  getAllProjects(): void {
    this.projectService.getAllProjects().subscribe( {
      next:(data) =>{
        this.projects = data;
      },
      error:(error)=>{
        console.error("Error loading projects, projects are null!")
      }
    })
  }

  deleteProject(id: number|null): void {
    if (id != null){
      this.projectService.deleteProjectById(id).subscribe({
        next:(response) => {
          console.log('Project deleted successfully:', response);
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
