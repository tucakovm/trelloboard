import {Component, OnInit} from '@angular/core';
import { FormGroup, Validators, FormBuilder } from '@angular/forms';
import { Task } from '../../model/task';
import { TaskService } from '../../services/task.service';
import { Status } from '../../model/status';
import {ActivatedRoute, Router} from "@angular/router";

@Component({
  selector: 'app-tasks',
  templateUrl: './tasks.component.html',
  styleUrls: ['./tasks.component.css']
})
export class TasksComponent implements OnInit {
  taskForm: FormGroup;
  projectId!: string;

  constructor(
    private fb: FormBuilder,
    private taskService: TaskService,
    private route: ActivatedRoute,
    private router:Router,
  ) {
    this.taskForm = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      description: ['', [Validators.required, Validators.minLength(10)]]
    });
  }
  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      this.projectId = params.get('projectId')!;
    });
  }

  onSubmit(): void {
    if (this.taskForm.valid) {
      const taskData: Task = this.taskForm.value;
      const submittedTask: Task = new Task(taskData.name, taskData.description, Status.Pending, this.projectId);

      console.log('Submitted Task Data:', submittedTask);
      this.taskService.createTask(submittedTask).subscribe({
        next: (response) => {
          this.router.navigate(['tasks',this.projectId])
          console.log('Task created successfully:', response);
        },
        error: (error) => {
          console.error('Error creating task:', error);
        }
      });
    }
  }
}
