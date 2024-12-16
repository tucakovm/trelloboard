import { Component} from '@angular/core';
import { FormBuilder, FormGroup, Validators , AbstractControl , ValidationErrors } from '@angular/forms';
import { User } from '../../model/user';
import { Project } from '../../model/project';
import { ProjectService } from '../../services/project.service';
import { AuthService } from '../../services/auth.service';
import { UserFP } from '../../model/userForProject';
import { Router } from '@angular/router';

@Component({
  selector: 'app-project-create',
  templateUrl: './project-create.component.html',
  styleUrl: './project-create.component.css'
})
export class ProjectCreateComponent{
  projectForm: FormGroup;
 //Dodati validaciju da nema maxmember bude veci od minmembera
  constructor(private fb: FormBuilder, private projectService:ProjectService , private authService:AuthService, private router:Router) {
    this.projectForm = this.fb.group(
      {
        name: ['', [Validators.required, Validators.minLength(3)]],
        completionDate: [null, [Validators.required, this.futureDateValidator]],
        minMembers: [0, [Validators.required, Validators.min(1)]],
        maxMembers: [0, [Validators.required, Validators.min(1)]],
      },
      { validators: this.maxGreaterThanMinValidator }
    );

  }

  // Prilagođena validacija
  maxGreaterThanMinValidator(control: AbstractControl): { [key: string]: boolean } | null {
    const minMembers = control.get('minMembers')?.value;
    const maxMembers = control.get('maxMembers')?.value;
    console.log('Validator values:', { minMembers, maxMembers });

    if (!isNaN(minMembers) && !isNaN(maxMembers) && maxMembers <= minMembers) {
      return { maxLessThanOrEqualMin: true }; // Error if maxMembers <= minMembers
    }
    return null;
  }

    // Validator to check if CompletionDate is in the future
    futureDateValidator(control: AbstractControl): ValidationErrors | null {
      const selectedDate = new Date(control.value);
      const currentDate = new Date();

      // Remove time part for accurate comparison
      currentDate.setHours(0, 0, 0, 0);

      if (selectedDate <= currentDate) {
        return { notFutureDate: true };
      }
      return null;
    }



  onSubmit(): void {

    let tokenRole = this.authService.getUserRoles();
    if (this.projectForm.valid && tokenRole == "Manager") {
      const projectData: Project = this.projectForm.value;

      let completionDate = new Date(projectData.completionDate);
      completionDate.setHours(0, 0, 0);

      let isoDate = completionDate.toISOString();

      let tokenUsername = this.authService.getUserName();
      let tokenId = this.authService.getUserId();
      projectData.manager = new UserFP(tokenId,tokenUsername,tokenRole)


      let submittedProject: Project = new Project(
        "",
        projectData.name,
        completionDate,
        projectData.minMembers,
        projectData.maxMembers,
        projectData.manager,
        projectData.members
      );

      console.log('Submitted Project Data:', submittedProject);

      this.projectService.createProject(submittedProject).subscribe({
        next: (response) => {
          this.router.navigate(['/all-projects'])
          console.log('Project created successfully:', response);

          // Kreiraj workflow nakon uspešnog kreiranja projekta
          this.projectService.createWorkflow("nekiId", "nekoIme").subscribe({
            next: (workflowResponse) => {
              console.log('Workflow created:', workflowResponse);
              this.router.navigate(['/all-projects']);
            },
            error: (error) => {
              console.error('Error creating workflow:', error);
            }
          });
        },
        error: (error) => {
          console.error('Error creating project:', error);
        },
        complete: () => {
          // console.log('Project creation process completed.');
        }
      });
    }
    else{
      console.error("Error submiting form")
    }
  }

}

