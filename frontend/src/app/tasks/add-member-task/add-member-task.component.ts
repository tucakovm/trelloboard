import {Component, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from "@angular/forms";
import {Task} from "../../model/task";
import {ActivatedRoute, Router} from "@angular/router";
import {UserFP} from "../../model/userForProject";
import {UserService} from "../../services/user.service";
import {TaskService} from "../../services/task.service";

@Component({
  selector: 'app-add-member-task',
  templateUrl: './add-member-task.component.html',
  styleUrl: './add-member-task.component.css'
})
export class AddMemberTaskComponent implements OnInit {
  addMemberForm: FormGroup;
  id: string | null = null;
  task: Task | null = null;
  userAlreadyExists: boolean = false;
  errorNotOnProject: boolean = false; // TODO !!!!!!!!!!!!!!!!!!!!!!!!!!! project-task

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private userService: UserService,
    private taskService: TaskService,
  ) {
    this.addMemberForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3)]]
    });
  }

  ngOnInit(): void {
  }

  onSubmit() {
    this.id = this.route.snapshot.paramMap.get('taskId');
    if (this.addMemberForm.valid) {
      this.userService.getUserByUsername(this.addMemberForm.value.username).subscribe((resp) => {
        let user: UserFP = new UserFP(resp.user.id, resp.user.username, resp.user.role);
        console.log("res1" + resp.user)
        console.log("res1id" + this.id)

        if (this.id && user) {
          this.taskService.getById(this.id).subscribe((resp) => {
            console.log("res2" + resp)
            this.task = resp;
            this.userAlreadyExists = false; // Resetuje se pri svakom pokušaju dodavanja

            if (this.task) {
              const exists = this.task.members.some(
                (member) => member.id === user.id || member.username === user.username);
              if (exists) {
                this.userAlreadyExists = true;
                return;
              } else if (this.id) {
                this.taskService
                  .AddMemberToTask(this.id, user)
                  .subscribe(() => console.log('Member successfully added'));
                this.router.navigate([`/task-details/${this.id}`]);
              }
            }
          });
        }
      });
    }
  }
}


