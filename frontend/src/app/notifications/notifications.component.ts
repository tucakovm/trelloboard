import {Component, OnInit} from '@angular/core';
import {AuthService} from "../services/auth.service";
import {NotService} from "../services/notification.service";
import {Notification} from "../model/notification";

@Component({
  selector: 'app-notifications',
  templateUrl: './notifications.component.html',
  styleUrl: './notifications.component.css'
})
export class NotificationsComponent implements OnInit {
  notifications?:Notification[];

  constructor(private authService:AuthService, private notsService:NotService) {
  }

  ngOnInit(): void {
    this.getAllNots()
  }

  getAllNots(): void {
    let id = this.authService.getUserId();
    this.notsService.getAllNots(id).subscribe({
      next: (data) => {
        this.notifications = data;
        console.log('Notifications:', this.notifications); // Log notifications here
      },
      error: (error) => {
        this.notifications = [];
        console.error("Error loading notifications, notifications are null!");
      }
    });
  }


}


