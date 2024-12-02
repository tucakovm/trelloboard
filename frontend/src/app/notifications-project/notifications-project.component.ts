import {Component, OnInit} from '@angular/core';
import {NotService} from "../services/notification.service";
import {Notification} from "../model/notification";
import {ActivatedRoute} from "@angular/router";

@Component({
  selector: 'app-notifications-project',
  templateUrl: './notifications-project.component.html',
  styleUrl: './notifications-project.component.css'
})
export class NotificationsProjectComponent implements OnInit {
  notifications?:Notification[];
  id: string | null = null;
  constructor(private notsService:NotService, private route : ActivatedRoute) {
  }

  ngOnInit(): void {
    this.getAllNots()
  }

  getAllNots(): void {
    this.id = this.route.snapshot.paramMap.get('projectId');
    if (this.id != null) {
      this.notsService.getAllNots(this.id).subscribe({
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


}


