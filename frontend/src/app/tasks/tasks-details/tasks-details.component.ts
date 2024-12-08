import { Component, OnInit , OnChanges, SimpleChanges} from '@angular/core';
import { ProjectService } from "../../services/project.service";
import { ActivatedRoute, Router , NavigationEnd } from "@angular/router";
import { TaskService } from "../../services/task.service";
import { Task } from "../../model/task";
import { Location } from '@angular/common';
import { AuthService} from "../../services/auth.service";
import { Subscription } from 'rxjs';

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
  private taskSubscription: Subscription | null = null;

  constructor(
    private projectService: ProjectService,
    private taskService: TaskService,
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private location: Location,
    private authService: AuthService,
    
  ) {}

  ngOnInit(): void {
    this.id = this.activatedRoute.snapshot.paramMap.get('id');
    if (this.id) {
      this.getTask();
    }
  }

  ngOnDestroy(): void {
    // Prestanak pretplate pri uništavanju komponente
    if (this.taskSubscription) {
      this.taskSubscription.unsubscribe();
    }
  }


  getTask() {
    this.taskSubscription = this.taskService.task$.subscribe(
      (task) => {
        if (task && task.id === this.id) {
          this.task = task;
        } else {
          // Ako zadatak nije pronađen, ponovo dohvatite zadatak sa API-ja
          this.taskService.getById(this.id!).subscribe();
        }
      }
    );
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
    }
  }

  goBackToTasks() {
    this.location.back();
  }

  updateTask(): void {
    if (!this.task.id) return;

    this.taskService.updateTask(this.task.id, this.task).subscribe(
      (updatedTask: Task) => {
        this.getTask();
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
