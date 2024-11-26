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
import { RemoveMemberComponent } from './members/remove-member/remove-member.component';
import { MagicLinkComponent } from './magic-link/magic-link.component';
import { PassRecoveryComponent } from './pass-recovery/pass-recovery.component';
import { AddMemberComponent} from "./members/add-member/add-member.component";
import {TasksDetailsComponent} from "./tasks/tasks-details/tasks-details.component";
import {AddMemberTaskComponent} from "./tasks/add-member-task/add-member-task.component";
import { AuthGuard } from './auth.guard';
import { UnauthorizedComponent } from './unauthorized/unauthorized.component';

const routes: Routes = [
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  { path: 'verify/:username', component: VerifyComponent},  
  { path: 'add-project', component: ProjectCreateComponent , canActivate: [AuthGuard], data: { roles: ['Manager'] },},
  { path: 'add-task', component: TasksComponent },
  { path: 'register', redirectTo: '/register', pathMatch: 'full' },
  { path: 'all-projects', component: ProjectAllComponent , canActivate: [AuthGuard], data: { roles: ['User', 'Manager'] },},
  { path: 'tasks/create/:projectId', component: TasksComponent , canActivate: [AuthGuard], data: { roles: ['Manager'] },},
  { path: 'profile', component: ProfileComponent , canActivate: [AuthGuard], data: { roles: ['User', 'Manager'] },},
  { path: 'all-projects/:id', component: ProjectDetailsComponent , canActivate: [AuthGuard], data: { roles: ['User', 'Manager'] },},
  { path: 'all-projects/:projectId/add-member', component: AddMemberComponent , canActivate: [AuthGuard], data: { roles: ['Manager'] },},
  {
    path: 'all-projects/:projectId/remove-member',
    component: RemoveMemberComponent, canActivate: [AuthGuard], data: { roles: ['Manager'] },
  },
  { path: 'tasks/:projectId', component: TasksAllComponent , canActivate: [AuthGuard], data: { roles: ['User', 'Manager'] },},
  { path: 'magic-login', component: MagicLinkComponent },
  { path: 'change-password', component: PassRecoveryComponent },
  { path: 'task-details/:id', component: TasksDetailsComponent , canActivate: [AuthGuard], data: { roles: ['User', 'Manager'] },},
  { path: 'task-add-member/:taskId', component: AddMemberTaskComponent , canActivate: [AuthGuard], data: { roles: ['Manager'] },},
  { path: 'unauthorized', component: UnauthorizedComponent },

];
@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
