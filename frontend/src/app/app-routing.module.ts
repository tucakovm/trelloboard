import { Component, NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { RegisterComponent } from './register/register.component';
import { VerifyComponent } from './verify/verify.component';
import { ProjectCreateComponent } from './project/project-create/project-create.component';
import { TasksComponent } from './tasks/tasks-create/tasks.component';
import { ProjectAllComponent } from './project/project-all/project-all.component';
import { LoginComponent } from './login/login.component';
import { ProfileComponent } from './profile/profile.component';
import { ProjectDetailsComponent } from './project/project-details/project-details.component';
import { TasksAllComponent } from './tasks/tasks-all/tasks-all.component';
import { AddMemberComponent } from './members/add-member/add-member.component';
import { RemoveMemberComponent } from './members/remove-member/remove-member.component';

const routes: Routes = [
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  { path: 'verify/:username', component: VerifyComponent },
  { path: 'add-project', component: ProjectCreateComponent },
  { path: 'add-task', component: TasksComponent },
  { path: 'register', redirectTo: '/register', pathMatch: 'full' },
  { path: 'all-projects', component: ProjectAllComponent },
  { path: 'tasks/create/:projectId', component: TasksComponent },
  { path: 'profile', component: ProfileComponent },
  { path: 'all-projects/:id', component: ProjectDetailsComponent },
  { path: 'all-projects/:projectId/add-member', component: AddMemberComponent },
  { path: 'all-projects/:projectId/remove-member', component: RemoveMemberComponent },
  { path: 'tasks/:projectId', component: TasksAllComponent },
  
];
@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
