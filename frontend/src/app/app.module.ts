import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ProjectCreateComponent } from './project/project-create/project-create.component';
import { ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { RegisterComponent } from './register/register.component';
import { VerifyComponent } from './verify/verify.component';
import { TasksComponent } from './tasks/tasks.component';
import { ProjectAllComponent } from './project/project-all/project-all.component';

@NgModule({
  declarations: [
    AppComponent,
    ProjectCreateComponent,
    RegisterComponent,
    VerifyComponent,
    TasksComponent,
    ProjectAllComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ReactiveFormsModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
