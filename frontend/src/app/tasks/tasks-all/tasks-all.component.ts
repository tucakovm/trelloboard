import { Component, OnInit } from '@angular/core';
import { TaskService } from '../../services/task.service';
import { ActivatedRoute } from '@angular/router';
import { Task } from '../../model/task';
import { Status } from '../../model/status';

@Component({
  selector: 'app-tasks-all',
  templateUrl: './tasks-all.component.html',
  styleUrl: './tasks-all.component.css'
})
export class TasksAllComponent implements OnInit{
  id: string | null = null;
  constructor(private tasksService:TaskService,private route: ActivatedRoute){}
  tasks?:Task[];

  ngOnInit(): void {
      this.getAll();
  }

  getAll() {
    this.id = this.route.snapshot.paramMap.get('projectId');
    console.log("id:" + this.id);
    if (this.id) {
      this.tasksService.getAllTasksByProjectId(this.id).subscribe(
        (tasks: Task[]) => {
          // Proveri da li je tasks niz i nije prazan
          if (Array.isArray(tasks) && tasks.length > 0) {
            this.tasks = tasks.map(task => ({
              ...task,
              statusText: task.status
            }));
            console.log("Tasks: ", this.tasks);
          } else {
            this.tasks = []; // Ako nije niz ili je prazan, postavi praznu listu
            console.log("Empty task list.");
          }
        },
        (error) => {
          console.error('Error fetching tasks:', error);
        }
      );
    }
  }

}
