import { Component} from '@angular/core';
import { FormBuilder, FormGroup, Validators , AbstractControl } from '@angular/forms';
import { User } from '../../model/user';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';

@Component({
  selector: 'app-project-create',
  templateUrl: './project-create.component.html',
  styleUrl: './project-create.component.css'
})
export class ProjectCreateComponent{
  projectForm: FormGroup;

  constructor(private fb: FormBuilder, private projectService:ProjectService) {
    this.projectForm = this.fb.group(
      {
        Name: ['', [Validators.required, Validators.minLength(3)]],
        CompletionDate: [null, Validators.required],
        MinMembers: [0, [Validators.required, Validators.min(1)]],
        MaxMembers: [0, [Validators.required, Validators.min(1)]]
      },
      { validators: this.maxGreaterThanMinValidator } // Dodajemo validator na formu
    );
  }

  // Prilagođena validacija
  maxGreaterThanMinValidator(control: AbstractControl): { [key: string]: boolean } | null {
    const minMembers = control.get('MinMembers')?.value;
    const maxMembers = control.get('MaxMembers')?.value;

    if (minMembers !== null && maxMembers !== null && maxMembers <= minMembers) {
      return { maxLessThanOrEqualMin: true }; // Greška ako maxMembers nije veći od minMembers
    }
    return null; // Bez greške
  }

  onSubmit(): void {
    if (this.projectForm.valid) {
      const projectData: Project = this.projectForm.value;
      
      let completionDate = new Date(projectData.CompletionDate);
      completionDate.setHours(0, 0, 0);

      let submittedProject: Project = new Project(
        projectData.Name,
        completionDate,
        projectData.MinMembers,
        projectData.MaxMembers
      );
  
      console.log('Submitted Project Data:', submittedProject);
  
      this.projectService.createProject(submittedProject).subscribe({
        next: (response) => {
          console.log('Project created successfully:', response);
        },
        error: (error) => {
          console.error('Error creating project:', error);
        },
        complete: () => {
          // console.log('Project creation process completed.');
        }
      });
    }
  }
  
}

