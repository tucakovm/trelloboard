import { Component, OnInit } from '@angular/core';
import { ProjectService } from "../../services/project.service";
import { ActivatedRoute, Router } from "@angular/router";
import { TaskService } from "../../services/task.service";
import { Task } from "../../model/task";
import { Location } from '@angular/common';
import { AuthService} from "../../services/auth.service";

@Component({
  selector: 'app-tasks-details',
  templateUrl: './tasks-details.component.html',
  styleUrls: ['./tasks-details.component.css']
})
export class TasksDetailsComponent implements OnInit {
  id: string | null = null;
  task: Task = {
    id: '',
    name: '',
    status: '',
    description: '',
    project_id: '',
    members: []
  };

  constructor(
    private projectService: ProjectService,
    private taskService: TaskService,
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private location: Location,
    private authService: AuthService,
  ) {}

  ngOnInit(): void {
    this.getTask();
  }

  getTask() {
    this.id = this.activatedRoute.snapshot.paramMap.get('id');
    if (this.id) {
      this.taskService.getById(this.id).subscribe(
        (task: Task | null) => {
          if (task) {
            this.task = task;
          } else {
            console.error('Task not found or an error occurred.');
          }
        },
        (error) => {
          console.error('Error fetching task:', error);
        }
      );
    }
  }

  deleteMember(memberId: string) {
    if (!this.id) return;

    if (this.task.status === "Done") {
      alert('Cannot remove member. The task is already marked as "Done".');
      return;
    }

    this.taskService.removeMemberFromTask(this.id, memberId).subscribe(
      () => {
        console.log('Member removed successfully.');
        this.task.members = this.task.members.filter(member => member.id !== memberId);
        this.getTask();
      },
      (error) => {
        console.error('Error removing member:', error);
      }
    );
  }


  addMember() {
    if (this.id) {
      this.router.navigate(['/task-add-member', this.id ]);
      this.getTask();
    }
  }

  goBackToTasks() {
    this.location.back();
  }

  updateTask(): void {
    if (!this.task.id) return;

    this.taskService.updateTask(this.task.id, this.task).subscribe(
      (updatedTask: Task) => {
        console.log('Task updated successfully:', updatedTask);
        alert('Task updated successfully!');
      },
      (error) => {
        console.error('Error updating task:', error);
        alert('An error occurred while updating the task.');
      }
    );
  }
  isManager(): boolean{
    return this.authService.isManager()
  }
}
