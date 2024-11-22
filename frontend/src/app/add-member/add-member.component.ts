import { Component , OnInit} from '@angular/core';
import { FormGroup , FormBuilder, Validators} from '@angular/forms';
import { UserService } from '../services/user.service';
import { ProjectService } from '../services/project.service';
import { UserFP } from '../model/userForProject';
import { ActivatedRoute } from '@angular/router';
import { Project } from '../model/project';

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

  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private projectService: ProjectService,
    private route: ActivatedRoute
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
        console.log('user resp: ' + JSON.stringify(resp), resp);
        let user: UserFP = new UserFP(resp.user.id, resp.user.username, resp.user.role);
        console.log(user);

        if (this.id && user) {
          this.projectService.getById(this.id).subscribe((resp) => {
            this.project = resp;
            this.userAlreadyExists = false; // Resetuje se pri svakom pokušaju dodavanja

            if (this.project) {
              console.log("Maxinum number bool1: ",this.maximumNumber)
              this.maximumNumber = this.project.members.length >= this.project.maxMembers;
              console.log("Maxinum number bool2: ",this.maximumNumber)
              console.log("project members lenght: ",this.project.members.length)
              console.log("project maxmembers: ",this.project.maxMembers)
              // Provera da li korisnik već postoji u projektu
              const exists = this.project.members.some(
                (member) => member.id === user.id || member.username === user.username
              );
              if (exists) {
                this.userAlreadyExists = true;
              } else if (!this.maximumNumber && this.id) {
                this.projectService
                  .createMember(this.id, user)
                  .subscribe(() => console.log('Member successfully added'));
              }
            }
          });
        }
      });
    }
  }
}