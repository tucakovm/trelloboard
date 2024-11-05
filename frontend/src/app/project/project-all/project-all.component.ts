import { Component, OnInit } from '@angular/core';
import { ProjectService } from '../../services/project.service';
import { Project } from '../../model/project';

@Component({
  selector: 'app-project-all',
  templateUrl: './project-all.component.html',
  styleUrl: './project-all.component.css'
})
export class ProjectAllComponent implements OnInit{
  projects?:Project[];
  constructor(private projectService:ProjectService){}

  ngOnInit(): void {
    this.getAllProjects();
  }

  getAllProjects(): void {
    this.projectService.getAllProjects().subscribe((data: Project[]) => {
      this.projects = data;
    });
  }

  deleteProject(id: number): void {
    this.projectService.deleteProjectById(id).subscribe({
      next:(response) => {
        console.log('Project deleted successfully:', response);
      },
      error: (error) => {
        console.error('Error deleting project:', error);
      },
    })
  }
}
