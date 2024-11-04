import { Component } from '@angular/core';
import { FormGroup, Validators, FormBuilder } from '@angular/forms';
import { Task } from '../model/task';
import { TaskService } from '../services/task.service';

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
        name: ['', [Validators.required, Validators.minLength(3)]],
        description:['',[Validators.required, Validators.minLength(10)]]
      },
      
    );
  }

  
  onSubmit(): void {
    if (this.taskForm.valid) {
      const taskData: Task = this.taskForm.value;
      let submittedTask: Task = new Task(taskData.Name,taskData.Description);
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
