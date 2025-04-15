import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import {WorkflowService} from "../../services/workflow.service";
import { Task } from "../../model/task";
import * as d3 from 'd3';
import {main} from "@popperjs/core";
import { ToastrService } from 'ngx-toastr';



@Component({
  selector: 'app-project-details',
  templateUrl: './project-details.component.html',
  styleUrl: './project-details.component.css'
})
export class ProjectDetailsComponent implements OnInit {
  id: string | null = null;
  userId: string | null = null;
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
    members: [
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

  comboBoxTasks: Task[] = [];
  tasks: any[] = [];
  edges: any[] = [];

  workflow: any | null = null;
  maxLengthAchieved: boolean = false;

  constructor(private projectService: ProjectService, private workflowService: WorkflowService, private route: ActivatedRoute, private tasksService: TaskService, private router: Router, private authService: AuthService, private toastr: ToastrService) {
  }

  ngOnInit(): void {
    this.getProject();
    this.getWorkflow();
    console.log(this.project)
    console.log(this.workflow)
  }

  getProject() {
    console.log("test1");
    this.id = this.route.snapshot.paramMap.get('id');
    this.userId = this.authService.getUserId();

    if (this.id && this.userId) {
      this.projectService.getById(this.id, this.userId).subscribe(
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
    if (!this.id) {
      console.error('Project ID is missing.');
      return;
    }

    this.projectService.getWorkflowByProjectId(this.id).subscribe({
      next: (workflow) => {
        console.log('Workflow from API:', workflow);

        this.workflow = workflow;
        this.comboBoxTasks = this.workflow.aco.tasks;

        // Mapa za čuvanje pozicija zadataka
        const taskPositions = new Map<string, { left: number; top: number }>();
        let yLevel = 100; // Početna Y koordinata
        const xSpacing = 150; // Horizontalni razmak između čvorova
        const ySpacing = 100; // Vertikalni razmak između čvorova
        let direction = 1; // 1 za desno, -1 za levo
        let levelCounter = 0; // Brojač za praćenje nivoa
        const maxOffset = 200; // Maksimalni pomak u levo ili desno

        // Funkcija za rekurzivno dodeljivanje pozicija
        const setTaskPosition = (task: any, level: number) => {
          if (!taskPositions.has(task.id)) {
            // Ako je brojač dostigao 3, promeni smer i resetuj brojač
            if (levelCounter >= 3) {
              direction *= -1; // Promeni smer
              levelCounter = 0; // Resetuj brojač
            }

            // Ograniči X koordinatu kako ne bi otišla previše u levo ili desno
            const xOffset = direction === 1 ? Math.min(level * xSpacing, maxOffset) : Math.max(-level * xSpacing, -maxOffset);

            taskPositions.set(task.id, { left: 400 + xOffset, top: yLevel });

            // Povećaj Y koordinatu za sledeći čvor
            yLevel += ySpacing;
            levelCounter++; // Povećaj brojač
          }
        };

        // Postavljanje pozicija zadataka prema zavisnostima
        this.workflow.aco.workflow.tasks.forEach((task: any) => {
          if (!task.dependencies.length) {
            // Ako nema zavisnosti, postavi na prvi nivo
            setTaskPosition(task, 0);
          } else {
            // Ako ima zavisnosti, postavi na sledeći nivo
            const maxLevel = Math.max(...task.dependencies.map((dep: string) => {
              const depTask = this.workflow.aco.workflow.tasks.find((t: any) => t.id === dep);
              return depTask ? (taskPositions.get(depTask.id)?.left ?? 0) / xSpacing : 0;
            }));
            setTaskPosition(task, maxLevel + 1);
          }
        });

        // Generisanje lista zadataka sa ispravnim pozicijama
        this.tasks = this.workflow.aco.workflow.tasks.map((task: any) => ({
          ...task,
          left: taskPositions.get(task.id)!.left,
          top: taskPositions.get(task.id)!.top,
        }));

        // Generisanje ivica
        this.edges = [];
        this.workflow.aco.workflow.tasks.forEach((task: any) => {
          task.dependencies.forEach((dep: string) => {
            const fromTask = this.tasks.find((t) => t.id === dep);
            const toTask = this.tasks.find((t) => t.id === task.id);

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

        // Renderovanje grafa
        this.renderGraph();
      },
      error: (error) => {
        console.error('Error fetching workflow:', error);
        this.workflow = null;
      },
    });
  }

  deleteAllTasksByProjectId(id: string) {
    this.tasksService.deleteTasksByProjectId(id).subscribe({
      next: (response) => {
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
    if (this.id != null) {
      this.projectService.deleteProjectById(this.id).subscribe({
        next: (response) => {
          console.log('Project deleted successfully:', response);
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
      this.projectService.createWorkflow(this.project.id, this.project.name).subscribe(
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
    if (this.project.id) {
      // Prvo dodajemo glavni task
      this.workflowService.addTask(this.project.id, this.mainTask).subscribe({
        next: () => {
          console.log('Main task successfully added');
          this.toastr.success('Main task successfully added'); // Prikaz uspeha

          // Nakon uspešnog dodavanja glavnog taska, dodajemo zavisni task
          this.dependentTask.dependencies = [this.mainTask.id];
          console.log('Payload za dependent task:', {
            project_id: this.project.id,
            task: this.dependentTask,
          });

          // Povezujemo zavisni task s glavnim
          this.workflowService.addTask(this.project.id, this.dependentTask).subscribe({
            next: () => {
              console.log('Dependent task successfully added');
              this.toastr.success('Dependent task successfully added'); // Prikaz uspeha
            },
            error: (err) => {
              console.error('Failed to add dependent task', err);
              this.toastr.error(`Failed to add dependent task: ${err.message}`); // Prikaz greške
            },
          });
        },
        error: (err) => {
          console.error('Failed to add main task', err);
          this.toastr.error(`Failed to add main task: ${err.message}`); // Prikaz greške
        },
      });
    }
  }

  renderGraph(): void {
    const svg = d3.select('#workflow-graph');
    svg.selectAll('*').remove(); // Brisanje prethodnih elemenata

    // Crtanje linija (edges)
    const edges = svg.selectAll('line')
      .data(this.edges)
      .enter()
      .append('line')
      .attr('x1', (d: any) => d.x1)
      .attr('y1', (d: any) => d.y1)
      .attr('x2', (d: any) => d.x2)
      .attr('y2', (d: any) => d.y2)
      .attr('stroke', 'black')
      .attr('stroke-width', 2);

    // Dodavanje strelica na linije
    svg.append('defs').selectAll('marker')
      .data(this.edges)
      .enter()
      .append('marker')
      .attr('id', (d: any, i: number) => `arrowhead-${i}`)
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 10) // Podesite refX za pozicioniranje strelice
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', 'black');

    edges.attr('marker-end', (d: any, i: number) => `url(#arrowhead-${i})`);

    // Dodavanje teksta "depends on" na linije
    svg.selectAll('text.edge-label')
      .data(this.edges)
      .enter()
      .append('text')
      .attr('class', 'edge-label')
      .attr('x', (d: any) => (d.x1 + d.x2) / 2)
      .attr('y', (d: any) => (d.y1 + d.y2) / 2)
      .attr('dy', -5)
      .attr('text-anchor', 'middle')
      .attr('fill', 'black')
      .text('depends on');

    // Crtanje čvorova (taskova)
    const node = svg.selectAll('g.node')
      .data(this.tasks)
      .enter()
      .append('g')
      .attr('class', 'node')
      .attr('transform', (d: any) => `translate(${d.left},${d.top})`);

    // Pravougaonik za task
    node.append('rect')
      .attr('width', 100)
      .attr('height', 50)
      .attr('fill', '#0074D9')
      .attr('rx', 5)
      .attr('ry', 5);

    // Ime taska
    node.append('text')
      .attr('x', 50)
      .attr('y', 25)
      .attr('dy', '.35em')
      .attr('text-anchor', 'middle')
      .attr('fill', 'white')
      .text((d: any) => d.name);

    // Dodavanje teksta "Blocked" ispod imena taska
    node.append('text')
      .attr('x', 50)
      .attr('y', 45) // Pomeramo tekst malo niže
      .attr('dy', '.35em')
      .attr('text-anchor', 'middle')
      .attr('fill', 'red') // Crvena boja za "Blocked"
      .text((d: any) => d.blocked);

    // Dodavanje strelica ka taskovima od kojih zavisi
    this.tasks.forEach((task: any) => {
      task.dependencies.forEach((dep: string) => {
        const fromTask = this.tasks.find((t: any) => t.id === dep);
        const toTask = task;

        if (fromTask && toTask) {
          svg.append('line')
            .attr('x1', fromTask.left + 50)
            .attr('y1', fromTask.top + 25)
            .attr('x2', toTask.left + 50)
            .attr('y2', toTask.top + 25)
            .attr('stroke', 'black')
            .attr('stroke-width', 2)
            .attr('marker-end', 'url(#arrowhead)'); // Dodajemo strelicu
        }
      });
    });
  }

  showTaskModal = false


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

  protected readonly main = main;
}
