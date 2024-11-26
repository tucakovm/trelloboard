import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AddMemberTaskComponent } from './add-member-task.component';

describe('AddMemberTaskComponent', () => {
  let component: AddMemberTaskComponent;
  let fixture: ComponentFixture<AddMemberTaskComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [AddMemberTaskComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AddMemberTaskComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
