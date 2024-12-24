import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import {WorkflowService} from "../../services/workflow.service";
import * as d3 from 'd3';


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

  mainTask: any = {
    id: '',
    name: '',
    description: '',
    dependencies: [],
    blocked: false,
  };
  dependentTask: any = {
    id: '',
    name: '',
    description: '',
    dependencies: [],
    blocked: false,
  };

  tasks: any[] = [];
  edges: any[] = [];

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
      this.projectService.getWorkflowByProjectId(this.id).subscribe({
        next: (workflow) => {
          this.workflow = workflow;
          console.log('Workflow:', this.workflow);


          // Generisanje taskova sa pozicijama
          this.tasks = this.workflow.tasks.map((task: any, index: number) => ({
            ...task,
            top: 100 + index * 100, // Y pozicija
            left: 200 + (index % 4) * 150, // X pozicija
          }));

          // Generisanje linija zavisnosti
          this.edges = [];
          this.workflow.tasks.forEach((task: any) => {
            task.dependencies.forEach((dep: string) => {
              const fromTask = this.tasks.find((t) => t.id === task.id);
              const toTask = this.tasks.find((t) => t.id === dep);
              if (fromTask && toTask) {
                this.edges.push({
                  x1: fromTask.left + 50,
                  y1: fromTask.top + 25,
                  x2: toTask.left + 50,
                  y2: toTask.top + 25,
                });
              }
            });
          });

          // Renderovanje grafikona
          this.renderGraph();

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
      error: (error) => {
        if (error.status === 412) {  // Ako je status greške 412, znači da je task deo workflow-a
          console.error("Task cannot be deleted because it is part of a workflow.");
          alert("Task cannot be deleted because it is part of a workflow.");
        } else {
          console.error("Error deleting task:", error);
          alert("An error occurred while deleting the task.");
        }
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
            this.getWorkflow();
          },
          error: (error) => {
            console.error('Error creating workflow:', error);
          }
        });
    }
  }

  addTasks() {

    console.log(this.mainTask);
    if(this.project.id) {
      // Prvo dodajemo glavnu task
      this.workflowService.addTask(this.project.id, this.mainTask).subscribe({
        next: () => {
          console.log('Main task successfully added');
          // Nakon uspešnog dodavanja glavnog taska, dodajemo zavisni task
          this.dependentTask.dependencies = [this.mainTask.id];
          console.log('Payload za dependent task:', {
            project_id: this.project.id,
            task: this.dependentTask
          });
// Povezujemo zavisni task s glavnim
          this.workflowService.addTask(this.project.id, this.dependentTask).subscribe({
            next: () => console.log('Dependent task successfully added'),
            error: (err) => console.error('Failed to add dependent task', err),
          });
        },
        error: (err) => console.error('Failed to add main task', err),
      });
    }
  }




  renderGraph(): void {
    const svg = d3.select('#workflow-graph');
    svg.selectAll('*').remove(); // Brisanje prethodnih elemenata

    // Crtanje linija (edges)
    svg.selectAll('line')
      .data(this.edges)
      .enter()
      .append('line')
      .attr('x1', (d: any) => d.x1)
      .attr('y1', (d: any) => d.y1)
      .attr('x2', (d: any) => d.x2)
      .attr('y2', (d: any) => d.y2)
      .attr('stroke', 'black')
      .attr('stroke-width', 2);

    // Crtanje čvorova (taskova)
    const node = svg.selectAll('g.node')
      .data(this.tasks)
      .enter()
      .append('g')
      .attr('class', 'node')
      .attr('transform', (d: any) => `translate(${d.left},${d.top})`);

    node.append('rect')
      .attr('width', 100)
      .attr('height', 50)
      .attr('fill', '#0074D9')
      .attr('rx', 5)
      .attr('ry', 5);

    node.append('text')
      .attr('x', 50)
      .attr('y', 25)
      .attr('dy', '.35em')
      .attr('text-anchor', 'middle')
      .attr('fill', 'white')
      .text((d: any) => d.name);
  }


  showTaskModal = false;


//  dependentTasks: { taskName: string }[] = [];

// Funkcija za otvaranje modala
  openTaskModal() {
    this.showTaskModal = true;
   // this.mainTask = { taskName: '' };
  //  this.dependentTasks = [];
  }
/*
// Dodavanje novog zavisnog taska
  addDependentTask() {
    this.dependentTasks.push({ taskName: '' });
  }

// Brisanje zavisnog taska
  removeDependentTask(index: number) {
    this.dependentTasks.splice(index, 1);
  }

// Zatvaranje modala
  closeTaskModal() {
    this.showTaskModal = false;
    this.mainTask = { taskName: '' };
    this.dependentTasks = [];
  }

// Čuvanje taskova
  saveTasks() {
    const workflowId = this.workflow?.project_id; // ID trenutnog workflow-a
    if (workflowId && this.mainTask.taskName) {
      // Kreiraj glavni task
      const mainTaskPayload = {
        task_id: `main-task-${workflowId}`,
        task_name: this.mainTask.taskName,
        dependencies: [],
      };

      // Dodaj glavni task
      this.workflowService.addTaskToWorkflow(workflowId, mainTaskPayload).subscribe({
        next: (response) => {
          console.log('Main task added:', response);

          // Dodaj zavisne taskove
          this.dependentTasks.forEach((task, index) => {
            const dependentTaskPayload = {
              task_id: `dependent-task-${workflowId}-${index}`,
              task_name: task.taskName,
              dependencies: [mainTaskPayload.task_id],
            };

            this.workflowService.addTaskToWorkflow(workflowId, dependentTaskPayload).subscribe({
              next: (res) => console.log(`Dependent task ${index} added:`, res),
              error: (err) => console.error(`Error adding dependent task ${index}:`, err),
            });
          });

          this.closeTaskModal();
          this.getWorkflow(); // Osveži prikaz workflow-a
        },
        error: (err) => {
          console.error('Error adding main task:', err);
        },
      });
    }
  }
*/

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
