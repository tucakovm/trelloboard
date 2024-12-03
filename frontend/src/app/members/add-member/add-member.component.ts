import { Component , OnInit} from '@angular/core';
import { FormGroup , FormBuilder, Validators} from '@angular/forms';
import { UserService } from '../../services/user.service';
import { ProjectService } from '../../services/project.service';
import { UserFP } from '../../model/userForProject';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../model/project';
import {AuthService} from "../../services/auth.service";
import { TaskService } from '../../services/task.service';
import { Task } from '../../model/task';

@Component({
  selector: 'app-add-member',
  templateUrl: './add-member.component.html',
  styleUrl: './add-member.component.css'
})
export class AddMemberComponent implements OnInit {
  addMemberForm: FormGroup;
  id: string | null = null;
  project: Project | null = null;
  maximumNumber: boolean = false;
  userAlreadyExists: boolean = false; // Dodato
  tasks : Task[] | null = null;
  cannotAddmemberToProject :boolean = false;

  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private projectService: ProjectService,
    private route: ActivatedRoute,
    private router:Router,
    private authService : AuthService,
    private taskService:TaskService,
  ) {
    this.addMemberForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3)]]
    });
  }

  ngOnInit(): void {}

  onSubmit() {
    this.id = this.route.snapshot.paramMap.get('projectId');
    if (this.addMemberForm.valid) {
      this.userService.getUserByUsername(this.addMemberForm.value.username).subscribe((resp) => {
        let user: UserFP = new UserFP(resp.user.id, resp.user.username, resp.user.role);

        if (this.id && user) {
          this.projectService.getById(this.id).subscribe((resp) => {
            this.project = resp;

            this.userAlreadyExists = false; // Resetuje se pri svakom pokušaju dodavanja
            if(this.id){
              this.taskService.getAllTasksByProjectId(this.id).subscribe((resp)=>{
                this.tasks = Array.isArray(resp.tasks) ? resp.tasks : Array.from(resp.tasks);
                console.log(this.tasks);
                if (this.project) {
                  this.maximumNumber = this.project.members.length >= this.project.maxMembers;
                  // Provera da li korisnik već postoji u projektu
                  const exists = this.project.members.some(
                    (member) => member.id === user.id || member.username === user.username
                  );

                  if(this.tasks){
                    this.cannotAddmemberToProject = this.tasks.every(task => task.status === 'Done');
                    console.log(this.cannotAddmemberToProject)
                  }
                  if (exists) {
                    this.userAlreadyExists = true;
                  } else if (!this.maximumNumber && this.id && !this.cannotAddmemberToProject && user.role != "Manager") {
                    this.projectService
                      .createMember(this.id, user)
                      .subscribe(() => console.log('Member successfully added'));
                      this.router.navigate(['/all-projects', this.id]);
                  }
                }
              })
            }
          });
        }
      });
    }
  }
}
