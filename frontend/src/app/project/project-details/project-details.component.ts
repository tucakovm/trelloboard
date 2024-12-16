import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import {WorkflowService} from "../../services/workflow.service";

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
  workflow: any | null = null;
  maxLengthAchieved:boolean = false;
  constructor(private projectService:ProjectService, private workflowService: WorkflowService,private route: ActivatedRoute,private tasksService:TaskService, private router:Router, private authService:AuthService){}

  ngOnInit(): void {
    this.getProject();
    this.getWorkflow();
    console.log(this.project)
    console.log(this.workflow)
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


  getWorkflow(): void {
    if (this.id) {
      this.projectService.getWorkflowByProjectId(this.project.id).subscribe({
        next: (workflow) => {
          this.workflow = workflow;
          console.log('Workflow:', this.workflow);
        },
        error: (error) => {
          console.error('Error fetching workflow:', error);
          this.workflow = null;
        }
      });
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
//ovde sam dodao project.id umesto id
  createWorkflow() {
    if (this.project.id && this.project.name) {
      //const newWorkflow = { project_id: this.project.id, project_name: this.project.name };
      this.projectService.createWorkflow(this.project.id,this.project.name).subscribe(
        {
          next: (workflowResponse) => {
            console.log('Workflow created:', workflowResponse);
          },
          error: (error) => {
            console.error('Error creating workflow:', error);
          }
        });
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

}
