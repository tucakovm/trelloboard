import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NotificationsProjectComponent } from './notifications-project.component';

describe('NotificationsProjectComponent', () => {
  let component: NotificationsProjectComponent;
  let fixture: ComponentFixture<NotificationsProjectComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [NotificationsProjectComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(NotificationsProjectComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
