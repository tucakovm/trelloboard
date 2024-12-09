import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TaskService } from '../../services/task.service';

@Component({
  selector: 'app-task-file',
  templateUrl: './task-file.component.html',
  styleUrls: ['./task-file.component.css'],
})
export class TaskFileComponent implements OnInit {
  taskId: string | null = null;
  selectedFile: File | null = null;
  files: any[] = []; // Array to hold the files associated with the task

  constructor(
    private activatedRoute: ActivatedRoute,
    private taskService: TaskService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.taskId = this.activatedRoute.snapshot.paramMap.get('id');
    if (this.taskId) {
      this.getFiles(); // Fetch the files associated with the task when the component initializes
    }
  }

  onFileSelected(event: any): void {
    this.selectedFile = event.target.files[0];
  }

  uploadFile(): void {
    if (this.selectedFile && this.taskId) {
      const formData = new FormData();
      formData.append('file', this.selectedFile, this.selectedFile.name);

      this.taskService.uploadFile(this.taskId, formData).subscribe(
        () => {
          alert('File uploaded successfully!');
          this.getFiles(); // Refresh the list of files after upload
        },
        (error) => {
          console.error('Error uploading file:', error);
          alert('Failed to upload file.');
        }
      );
    } else {
      alert('Please select a file first.');
    }
  }

  getFiles(): void {
    if (this.taskId) {
      this.taskService.getFiles(this.taskId).subscribe(
        (files) => {
          this.files = files;
        },
        (error) => {
          console.error('Error fetching files:', error);
          alert('Failed to fetch files.');
        }
      );
    }
  }

  downloadFile(fileId: string): void {
    this.taskService.downloadFile(fileId).subscribe(
      (response) => {
        const blob = response.body as Blob;  // Access the Blob body
        const fileName = response.headers.get('filename') || 'file';  // Get the filename from headers

        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = fileName;
        link.click();
      },
      (error) => {
        console.error('Error downloading file:', error);
        alert('Failed to download file.');
      }
    );
  }

}
