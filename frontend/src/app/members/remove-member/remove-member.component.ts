import { Component , OnInit} from '@angular/core';
import { FormGroup , FormBuilder, Validators} from '@angular/forms';
import { UserService } from '../../services/user.service';
import { ProjectService } from '../../services/project.service';
import { UserFP } from '../../model/userForProject';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../model/project';
import { TaskService } from '../../services/task.service';
import { Task } from '../../model/task';

@Component({
  selector: 'app-remove-member',
  templateUrl: './remove-member.component.html',
  styleUrl: './remove-member.component.css'
})
export class RemoveMemberComponent {
  removeMemberForm: FormGroup;
  id: string | null = null;
  project: Project | null = null;
  tasks : Task[] | null = null; 
  cannotRemoveMemberFromProject :boolean = false;
  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private projectService: ProjectService,
    private route: ActivatedRoute,
    private router:Router,
    private taskService:TaskService
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
        if(this.id){
          this.taskService.getAllTasksByProjectId(this.id).subscribe((resp)=>{
            this.tasks = Array.isArray(resp.tasks) ? resp.tasks : Array.from(resp.tasks);
            this.cannotRemoveMemberFromProject = this.getTasksForUser(user.id)
            if(user && this.id && !this.cannotRemoveMemberFromProject){
              this.projectService.removeMember(this.id,user).subscribe({
                next:(resp) => {
                  console.log("Successfully removed member!")
                  this.router.navigate(['/all-projects', this.id]);
                },
                error:(error) =>{
                  console.error("Error removing member:",error);
                }
              })
            }
          });
        }
      });
    }
  }
  getTasksForUser(userId: string): boolean {
    if (this.tasks && userId) {
      const filteredTasks = this.tasks.filter(
        (task) =>
          task.status === 'Working' &&
          task.members.some((member) => member.id === userId)
      );
      return filteredTasks.length > 0;
    }
    return false;
  }

}
