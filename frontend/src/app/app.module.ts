import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ProjectCreateComponent } from './project/project-create/project-create.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { RegisterComponent } from './register/register.component';
import { VerifyComponent } from './verify/verify.component';
import { TasksComponent } from './tasks/tasks-create/tasks.component';
import { ProjectAllComponent } from './project/project-all/project-all.component';
import { LoginComponent } from './login/login.component';
import { JwtHelperService, JWT_OPTIONS } from '@auth0/angular-jwt';
import { ProfileComponent } from './profile/profile.component';
import { ProjectDetailsComponent } from './project/project-details/project-details.component';
import { TasksAllComponent } from './tasks/tasks-all/tasks-all.component';
import { AddMemberComponent } from './add-member/add-member.component';


@NgModule({
  declarations: [
    AppComponent,
    ProjectCreateComponent,
    RegisterComponent,
    VerifyComponent,
    TasksComponent,
    ProjectAllComponent,
    LoginComponent,
    ProfileComponent,
    ProjectDetailsComponent,
    TasksAllComponent,
    AddMemberComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ReactiveFormsModule,
    HttpClientModule,
    FormsModule,
  ],
  providers: [
    { provide: JWT_OPTIONS, useValue: JWT_OPTIONS }, // This is necessary for JwtHelperService to work
    JwtHelperService, // Add JwtHelperService here
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
