import { Component , OnInit} from '@angular/core';
import { FormGroup , FormBuilder, Validators} from '@angular/forms';
import { UserService } from '../../services/user.service';
import { ProjectService } from '../../services/project.service';
import { UserFP } from '../../model/userForProject';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../model/project';

@Component({
  selector: 'app-remove-member',
  templateUrl: './remove-member.component.html',
  styleUrl: './remove-member.component.css'
})
export class RemoveMemberComponent {
  removeMemberForm: FormGroup;
  id: string | null = null;
  project: Project | null = null;
  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private projectService: ProjectService,
    private route: ActivatedRoute,
    private router:Router,
  ) {
    this.removeMemberForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3)]]
    });
  }
  onSubmit(){
    this.id = this.route.snapshot.paramMap.get('projectId');
    if (this.removeMemberForm.valid) {
      this.userService.getUserByUsername(this.removeMemberForm.value.username).subscribe((resp) => {
        let user: UserFP = new UserFP(resp.user.id, resp.user.username, resp.user.role);
        console.log(user);
        if(user && this.id){
          this.projectService.removeMember(this.id,user).subscribe({
            next:(resp) => {
              console.log("Successfully removed member!")
            },
            error:(error) =>{
              console.error("Error removing member:",error);
            }
          })
        }
      });
    }
  }
}
