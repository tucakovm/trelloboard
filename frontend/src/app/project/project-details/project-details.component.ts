import { Component, OnInit } from '@angular/core';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import { Edge, Node, Graph } from '@swimlane/ngx-graph';
import { WorkflowService} from "../../services/workflow.service";
import { Workflow } from '../../model/workflow';
import {TaskFW} from "../../model/TaskFW";

@Component({
  selector: 'app-project-details',
  templateUrl: './project-details.component.html',
  styleUrls: ['./project-details.component.css']
})
export class ProjectDetailsComponent implements OnInit {
  id: string | null = null;
  workflow: Workflow = {  // Define workflow property
    projectID: '',
    projectName: '',
    tasks: []
  };
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
      },
    ]

  };

  maxLengthAchieved: boolean = false;
  nodes: Node[] = [];
  edges: Edge[] = [];

  constructor(
    private projectService: ProjectService,
    private route: ActivatedRoute,
    private tasksService: TaskService,
    private router: Router,
    private authService: AuthService,
    private workflowService: WorkflowService
  ) {}

  ngOnInit(): void {
    this.getProject();
    console.log(this.project);
    this.id = this.route.snapshot.paramMap.get('projectId');
    if (this.id) {
      this.getWorkflowDetails(this.id);
    }
  }

  getProject() {
    console.log("Fetching project...");
    this.id = this.route.snapshot.paramMap.get('id');

    if (this.id) {
      this.projectService.getById(this.id).subscribe(
        (project: Project | null) => {
          if (project) {
            this.project = project;
            this.maxLengthAchieved = this.project.members.length >= this.project.maxMembers;
            this.createGraphData();
            this.getWorkflowDetails(this.project.id);
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

  getWorkflowDetails(projectId: string): void {
    this.workflowService.getWorkflow(projectId).subscribe(
      (workflow: Workflow) => {
        this.workflow = workflow;
        console.log('Workflow details:', this.workflow);
      },
      (error) => {
        console.error('Error fetching workflow:', error);
      }
    );
  }
  // Create nodes and edges for the graph
  createGraphData() {
    // Ensure tasks is defined in workflow before processing
    if (this.workflow.tasks && this.workflow.tasks.length > 0) {
      const taskMap = new Map();

      // Iterate over the tasks in workflow
      this.workflow.tasks.forEach((task) => {
        // Create a node for each task
        const taskNode: Node = {
          id: task.taskID,  // Use taskID as node ID
          label: task.taskName,  // Use taskName as node label
        };

        taskMap.set(task.taskID, taskNode);  // Add the task node to the task map
        this.nodes.push(taskNode);  // Add task node to the nodes array

        // Iterate over the task's dependencies to create edges
        task.dependencies.forEach((dependencyId) => {
          const edge: Edge = {
            source: dependencyId,  // Dependency task
            target: task.taskID,  // Current task
          };
          this.edges.push(edge);  // Add edge to the edges array
        });
      });
    } else {
      console.error('No tasks available in the workflow.');
    }
  }
  deleteAllTasksByProjectId(id: string) {
    this.tasksService.deleteTasksByProjectId(id).subscribe({
      next: (response) => {
        console.log("Tasks deleted successfully");
      },
      error: (error) => {
        console.error("Error deleting tasks: " + error);
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










  createTestWorkflow(): void {
    this.id = this.route.snapshot.paramMap.get('id');
    if (!this.project.id || !this.project.name) {
      console.log(this.project.id)
      console.log(this.project.name)
      console.error('Project ID or name is missing.');
      return;
    }

    // Step 1: Create Workflow
    this.workflowService.createWorkflow({
      project_id: this.id,
      project_name: this.project.name,

    }).subscribe(
      () => {
        console.log('Workflow created successfully.');

        // Step 2: Add Main Task
        const mainTask: TaskFW = {
          id: 'main-task',
          name: 'Main Task',
          description: 'This is the main task',
          dependencies: [],
          blocked: false
        };

        this.workflowService.addTask({project_id: this.id, task: mainTask}).subscribe(
          () => {
            console.log('Main Task added.');

            // Step 3: Add Dependent Task
            const dependentTask: TaskFW = {
              id: 'dependent-task',
              name: 'Dependent Task',
              description: 'This task depends on Main Task',
              dependencies: ['main-task'],
              blocked: true
            };

            this.workflowService.addTask({ project_id: this.id, task: dependentTask }).subscribe(
              () => {
                console.log('Dependent Task added.');
              },
              (error) => console.error('Error adding Dependent Task:', error)
            );
          },
          (error) => console.error('Error adding Main Task:', error)
        );
      },
      (error) => console.error('Error creating workflow:', error)
    );
  }
}

