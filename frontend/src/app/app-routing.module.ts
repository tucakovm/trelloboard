import { Component, NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { RegisterComponent } from './register/register.component';
import { VerifyComponent } from './verify/verify.component';
import { ProjectCreateComponent } from './project/project-create/project-create.component';
import { TasksComponent } from './tasks/tasks.component';
import { ProjectAllComponent } from './project/project-all/project-all.component';
import { LoginComponent } from './login/login.component';
import { ProfileComponent } from './profile/profile.component';

const routes: Routes = [
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  { path: 'verify', component: VerifyComponent },
  { path: 'add-project', component: ProjectCreateComponent },
  { path: 'add-task', component: TasksComponent },
  { path: 'register', redirectTo: '/register', pathMatch: 'full' },
  { path: 'all-projects', component: ProjectAllComponent },
  { path: 'tasks/:projectId', component: TasksComponent },
  { path: 'profile', component: ProfileComponent },
];
@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
