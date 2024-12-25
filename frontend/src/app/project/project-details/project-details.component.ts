import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-project-details',
  templateUrl: './project-details.component.html',
  styleUrl: './project-details.component.css'
})
export class ProjectDetailsComponent implements OnInit{
  id: string | null = null;
  project:Project = {
    id: '',
    name: '',
    completionDate: new Date(),
    minMembers: 0,
    maxMembers: 0,
    manager: {
        id: '',
        username: '',
        role: ''
    },
    members:[
      {
        id: '',
        username: '',
        role: ''
      },
    ]
  }
  maxLengthAchieved:boolean = false;
  constructor(private projectService:ProjectService,private route: ActivatedRoute,private tasksService:TaskService, private router:Router, private authService:AuthService){}

  ngOnInit(): void {
    this.getProject();
    console.log(this.project)
    this.getAnalytics();
  }

  getProject() {
    console.log("test1");
    this.id = this.route.snapshot.paramMap.get('id');

    if (this.id) {
      this.projectService.getById(this.id).subscribe(
        (project: Project | null) => {
          if (project) {
            this.project = project;
            this.maxLengthAchieved = this.project.members.length >= this.project.maxMembers;
          } else {
            console.error('Project not found or an error occurred.');
          }
        },
        (error) => {
          console.error('Error fetching project:', error);
        }
      );
    }
  }


  deleteAllTasksByProjectId(id:string){
    this.tasksService.deleteTasksByProjectId(id).subscribe({
      next:(response)=>{
        console.log("Tasks deleted sucessfuly")
      },
      error:(error)=>{
        console.error("Error deleting tasks:"+ error)
      }
    })
  }

  deleteProject(): void {
    if (this.id != null){
      this.projectService.deleteProjectById(this.id).subscribe({
        next:(response) => {
          console.log('Project deleted successfully:', response);
          if(this.id){
            this.deleteAllTasksByProjectId(this.id);
          }
          this.router.navigate(['/all-projects'])
        },
        error: (error) => {
          console.error('Error deleting project:', error);
        },
      })
    }
  }
  addTask(): void {
    if (this.id) {
      this.router.navigate(['/tasks/create', this.id]);
    }
  }

  allTasks():void{
    if (this.id) {
      this.router.navigate(['/tasks', this.id]);
    }
  }

  viewNotifications(): void {
    if (this.id) {
      this.router.navigate(['/app-notifications-project', this.id]);
    }
  }

  addMember(){
    if (this.id) {
      this.router.navigate(['/all-projects', this.id,"add-member" ]);
    }
  }
  removeMember(){
    if (this.id) {
      this.router.navigate(['/all-projects', this.id,"remove-member" ]);
    }
  }

  isManager(){
    return this.authService.isManager();
  }

  getAnalytics() {
    if (this.id) {
      this.projectService.getAnalyticsByProjectId(this.id).subscribe(
        (analyticsData) => {
          console.log('Analytics Data:', analyticsData);
        },
        (error) => {
          console.error('Error fetching analytics:', error);
        }
      );
    }
  }
}

