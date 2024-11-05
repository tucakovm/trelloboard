import { Component } from '@angular/core';
import { FormGroup, Validators, FormBuilder } from '@angular/forms';
import { Task } from '../model/task';
import { TaskService } from '../services/task.service';
import { Status } from '../model/status';

@Component({
  selector: 'app-tasks',
  templateUrl: './tasks.component.html',
  styleUrl: './tasks.component.css'
})
export class TasksComponent {

  taskForm : FormGroup;

  constructor(private fb: FormBuilder, private taskService:TaskService) {
    this.taskForm = this.fb.group(
      {
        Name: ['', [Validators.required, Validators.minLength(3)]],
        Description:['',[Validators.required, Validators.minLength(10)]]
      },
      
    );
  }

  
  onSubmit(): void {
    if (this.taskForm.valid) {
      const taskData: Task = this.taskForm.value;
      console.log(this.taskForm.value)
      let submittedTask: Task = new Task(taskData.Name,taskData.Description,Status.Pending);
      console.log('Submitted Task Data:', submittedTask);
      this.taskService.createTask(submittedTask).subscribe({
        next: (response) => {
            console.log('Task created successfully:', response);
        },
        error: (error) => {
            console.error('Error creating project:', error);
        },
        complete: () => {
            
        }
    });
    }
  }

  



}
