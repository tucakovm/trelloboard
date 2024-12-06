import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ProjectCreateComponent } from './project/project-create/project-create.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { RegisterComponent } from './register/register.component';
import { VerifyComponent } from './verify/verify.component';
import { TasksComponent } from './tasks/tasks-create/tasks.component';
import { ProjectAllComponent } from './project/project-all/project-all.component';
import { LoginComponent } from './login/login.component';
import { JwtHelperService, JWT_OPTIONS } from '@auth0/angular-jwt';
import { ProfileComponent } from './profile/profile.component';
import { ProjectDetailsComponent } from './project/project-details/project-details.component';
import { TasksAllComponent } from './tasks/tasks-all/tasks-all.component';
import { AddMemberComponent } from './members/add-member/add-member.component';
import { RecaptchaModule } from 'ng-recaptcha';
import { AuthInterceptor } from './auth.interceptor';
import { RemoveMemberComponent } from './members/remove-member/remove-member.component';
import { MagicLinkComponent } from './magic-link/magic-link.component';
import { PassRecoveryComponent } from './pass-recovery/pass-recovery.component';
import { TasksDetailsComponent } from './tasks/tasks-details/tasks-details.component';
import { AddMemberTaskComponent } from './tasks/add-member-task/add-member-task.component';
import { UnauthorizedComponent } from './unauthorized/unauthorized.component';
import { NotificationsComponent } from './notifications/notifications.component';
import { NotificationsProjectComponent } from './notifications-project/notifications-project.component';
import { NgxGraphModule } from '@swimlane/ngx-graph';
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
    RemoveMemberComponent,
    MagicLinkComponent,
    PassRecoveryComponent,
    TasksDetailsComponent,
    AddMemberTaskComponent,
    UnauthorizedComponent,
    NotificationsComponent,
    NotificationsProjectComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ReactiveFormsModule,
    HttpClientModule,
    FormsModule,
    RecaptchaModule,
    NgxGraphModule,

  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true, // Ovim omogućavamo više interceptora
    },
    { provide: JWT_OPTIONS, useValue: JWT_OPTIONS }, // This is necessary for JwtHelperService to work
    JwtHelperService, // Add JwtHelperService here
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
