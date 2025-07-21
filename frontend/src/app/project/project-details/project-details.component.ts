import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import { WorkflowService } from '../../services/workflow.service';
import { Task } from '../../model/task';
import * as d3 from 'd3';
import { ToastrService } from 'ngx-toastr';
import { Analytics } from '../../model/analytic';

@Component({
  selector: 'app-project-details',
  templateUrl: './project-details.component.html',
  styleUrls: ['./project-details.component.css']
})
export class ProjectDetailsComponent implements OnInit {

  analytics: Analytics | null = null;
  id: string | null = null;
  userId: string | null = null;

  project: Project = {
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
      }
    ]
  };

  maxLengthAchieved: boolean = false;

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

  showTaskModal = false;

  constructor(
    private projectService: ProjectService,
    private workflowService: WorkflowService,
    private route: ActivatedRoute,
    private tasksService: TaskService,
    private router: Router,
    private authService: AuthService,
    private toastr: ToastrService
  ) { }

  ngOnInit(): void {
    this.getProject();
    this.getWorkflow();
    this.getAnalytics();
  }

  getProject() {
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
        this.workflow = workflow;
        this.comboBoxTasks = this.workflow.aco.tasks;

        const taskPositions = new Map<string, { left: number; top: number }>();
        let yLevel = 100;
        const xSpacing = 150;
        const ySpacing = 100;
        let direction = 1;
        let levelCounter = 0;
        const maxOffset = 200;

        const setTaskPosition = (task: any, level: number) => {
          if (!taskPositions.has(task.id)) {
            if (levelCounter >= 3) {
              direction *= -1;
              levelCounter = 0;
            }
            const xOffset = direction === 1 ? Math.min(level * xSpacing, maxOffset) : Math.max(-level * xSpacing, -maxOffset);
            taskPositions.set(task.id, { left: 400 + xOffset, top: yLevel });
            yLevel += ySpacing;
            levelCounter++;
          }
        };

        this.workflow.aco.workflow.tasks.forEach((task: any) => {
          if (!task.dependencies.length) {
            setTaskPosition(task, 0);
          } else {
            const maxLevel = Math.max(...task.dependencies.map((dep: string) => {
              const depTask = this.workflow.aco.workflow.tasks.find((t: any) => t.id === dep);
              return depTask ? (taskPositions.get(depTask.id)?.left ?? 0) / xSpacing : 0;
            }));
            setTaskPosition(task, maxLevel + 1);
          }
        });

        this.tasks = this.workflow.aco.workflow.tasks.map((task: any) => ({
          ...task,
          left: taskPositions.get(task.id)!.left,
          top: taskPositions.get(task.id)!.top,
        }));

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
      next: () => {
        console.log("Tasks deleted successfully");
      },
      error: (error) => {
        if (error.status === 412) {
          console.error("Task cannot be deleted because it is part of a workflow.");
          alert("Task cannot be deleted because it is part of a workflow.");
        } else {
          console.error("Error deleting task:", error);
          alert("An error occurred while deleting the task.");
        }
      }
    });
  }

  deleteProject(): void {
    if (this.id != null) {
      this.projectService.deleteProjectById(this.id).subscribe({
        next: (response) => {
          console.log('Project deleted successfully:', response);
          if (this.id) {
            this.deleteAllTasksByProjectId(this.id);
          }
          this.router.navigate(['/all-projects']);
        },
        error: (error) => {
          console.error('Error deleting project:', error);
        },
      });
    }
  }

  createWorkflow() {
    if (this.project.id && this.project.name) {
      this.projectService.createWorkflow(this.project.id, this.project.name).subscribe({
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
    if (this.project.id) {
      this.workflowService.addTask(this.project.id, this.mainTask).subscribe({
        next: () => {
          this.toastr.success('Main task successfully added');
          this.dependentTask.dependencies = [this.mainTask.id];
          this.workflowService.addTask(this.project.id, this.dependentTask).subscribe({
            next: () => {
              this.toastr.success('Dependent task successfully added');
            },
            error: (err) => {
              this.toastr.error(`Failed to add dependent task: ${err.message}`);
            },
          });
        },
        error: (err) => {
          this.toastr.error(`Failed to add main task: ${err.message}`);
        },
      });
    }
  }

  renderGraph(): void {
    const svg = d3.select('#workflow-graph');
    svg.selectAll('*').remove();

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

    svg.append('defs').selectAll('marker')
      .data(this.edges)
      .enter()
      .append('marker')
      .attr('id', (d: any, i: number) => `arrowhead-${i}`)
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 10)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', 'black');

    edges.attr('marker-end', (d: any, i: number) => `url(#arrowhead-${i})`);

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

    node.append('text')
      .attr('x', 50)
      .attr('y', 45)
      .attr('dy', '.35em')
      .attr('text-anchor', 'middle')
      .attr('fill', 'red')
      .text((d: any) => d.blocked);
  }

  openTaskModal() {
    this.showTaskModal = true;
  }

  addTask(): void {
    if (this.id) {
      this.router.navigate(['/tasks/create', this.id]);
    }
  }

  allTasks(): void {
    if (this.id) {
      this.router.navigate(['/tasks', this.id]);
    }
  }

  viewNotifications(): void {
    if (this.id) {
      this.router.navigate(['/app-notifications-project', this.id]);
    }
  }

  addMember() {
    if (this.id) {
      this.router.navigate(['/all-projects', this.id, "add-member"]);
    }
  }

  removeMember() {
    if (this.id) {
      this.router.navigate(['/all-projects', this.id, "remove-member"]);
    }
  }

  isManager() {
    return this.authService.isManager();
  }

  objectKeys(obj: any): string[] {
    return obj ? Object.keys(obj) : [];
  }

  getAnalytics() {
    if (this.id) {
      this.projectService.getAnalyticsByProjectId(this.id).subscribe(
        (response: any) => {
          this.analytics = response.analytic;
        },
        (error) => {
          console.error('Error fetching analytics:', error);
        }
      );
    }
  }
}
