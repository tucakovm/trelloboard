import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { UserService } from '../../services/user.service';
import { TaskService } from '../../services/task.service';
import { AuthService } from '../../services/auth.service';
import { UserFP } from '../../model/userForProject';
import { Task } from '../../model/task';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-add-member-task',
  templateUrl: './add-member-task.component.html',
  styleUrls: ['./add-member-task.component.css'],
})
export class AddMemberTaskComponent implements OnInit {
  addMemberForm: FormGroup;
  id: string | null = null;
  task: Task | null = null;
  userAlreadyExists: boolean = false;
  errorNotOnProject: boolean = false; // TODO !!!!!!!!!!!!!!!!!!!!!!!!!!! project-task
  private taskSubscription: Subscription | null = null;

  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private userService: UserService,
    private taskService: TaskService,
    private authService: AuthService
  ) {
    this.addMemberForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3)]],
    });
  }

  ngOnInit(): void {
    this.id = this.route.snapshot.paramMap.get('taskId');
    if (this.id) {
      this.getTask(); // Pozivanje metode za preuzimanje zadatka na početku
    }
  }

  ngOnDestroy(): void {
    if (this.taskSubscription) {
      this.taskSubscription.unsubscribe(); // Otkazivanje pretplate kada komponenta bude uništena
    }
  }

  getTask() {
    // Pretplata na observable task-a
    if (this.id) {
      this.taskSubscription = this.taskService.task$.subscribe((task) => {
        if (task && task.id === this.id) {
          this.task = task; // Ažuriranje stanja zadatka
        } else {
          // Ako zadatak nije pronađen, ponovo ga dohvatiti sa servera
          if(this.id)
          this.taskService.getById(this.id).subscribe();
        }
      });
    }
  }

  onSubmit() {
    if (this.addMemberForm.valid && this.id) {
      this.userService
        .getUserByUsername(this.addMemberForm.value.username)
        .subscribe((resp) => {
          let user: UserFP = new UserFP(resp.user.id, resp.user.username, resp.user.role);
          console.log('User:', resp.user);
          console.log('Task ID:', this.id);

          if (this.id && user && user.role !== 'Manager') {
            this.taskService.getById(this.id).subscribe((taskResp) => {
              this.task = taskResp;
              this.userAlreadyExists = false; // Resetovanje greške

              if (this.task) {
                const exists = this.task.members.some(
                  (member) => member.id === user.id || member.username === user.username
                );
                if (exists) {
                  this.userAlreadyExists = true; // Ako korisnik već postoji u zadatku
                  return;
                } else {
                  // Dodavanje člana u zadatak
                  if(this.id)
                  this.taskService.AddMemberToTask(this.id, user).subscribe(() => {
                    console.log('Member successfully added');
                    this.router.navigate([`/task-details/${this.id}`]);
                  });
                }
              }
            });
          }
        });
    }
  }
}
