import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TasksAllComponent } from './tasks-all.component';

describe('TasksAllComponent', () => {
  let component: TasksAllComponent;
  let fixture: ComponentFixture<TasksAllComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [TasksAllComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(TasksAllComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
