import { Component , OnInit} from '@angular/core';
import { FormGroup , FormBuilder, Validators} from '@angular/forms';
import { UserService } from '../services/user.service';
import { ProjectService } from '../services/project.service';
import { UserFP } from '../model/userForProject';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-add-member',
  templateUrl: './add-member.component.html',
  styleUrl: './add-member.component.css'
})
export class AddMemberComponent implements OnInit {
  addMemberForm: FormGroup;
  id: string | null = null;

  constructor(private fb: FormBuilder, private userService:UserService, private projectService:ProjectService, private route:ActivatedRoute) {
    this.addMemberForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3)]]
    });
  }

  ngOnInit(): void {}

  onSubmit() {
    this.id = this.route.snapshot.paramMap.get('projectId');
    if (this.addMemberForm.valid) {
      this.userService.getUserByUsername(this.addMemberForm.value.username).subscribe((resp)=>{
        let user:UserFP = new UserFP(resp.id,resp.username,resp.role);
        console.log(this.id)
        if(this.id){
          console.log("id:" + this.id + "user:" + user +" uslo u submit 2");
          this.projectService.createMember(this.id,user).subscribe(resp=>console.log(resp));
      }
      })
    }
  }
}
