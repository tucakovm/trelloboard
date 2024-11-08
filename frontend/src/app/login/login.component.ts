import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { LoginService } from '../services/login.service';
import { LoggedUser } from '../model/LoggedUser';
import { Observer } from 'rxjs';
import { Router } from '@angular/router';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrl: './login.component.css'
})
export class LoginComponent {

  taskForm:FormGroup;
  errorOccurred: boolean = false;
  errorMessage: string = ''; 
  
  constructor(private loginService:LoginService,private fb:FormBuilder,private router:Router){
    this.taskForm = this.fb.group({
      username : ['',[Validators.required, Validators.minLength(4)]],
      password : ['',[Validators.required,Validators.minLength(4)]]
    });
  }
  
  onSubmit(){
    if(this.taskForm.valid){
      let loggedUser : LoggedUser = new LoggedUser(this.taskForm.value.username,this.taskForm.value.password);
      const observer: Observer<any> = {
        next: (res:any) => {
          console.log('Login success, resp:', res);
          // localStorage.setItem("jwt", res.accessToken);
          // this.router.navigate(['']);
        },
        error: (error:any) => {
          this.errorOccurred = true;
          this.errorMessage = 'Greška prilikom prijave. Proverite vaše podatke.';
          console.error('Greška prilikom prijave:', error);
        },
        complete:()=>{}
      };
      this.loginService.login(loggedUser).subscribe(observer);
  }
  }

}
