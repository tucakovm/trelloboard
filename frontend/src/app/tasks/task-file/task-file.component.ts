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
  files: any[] = [];

  constructor(
    private activatedRoute: ActivatedRoute,
    private taskService: TaskService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.taskId = this.activatedRoute.snapshot.paramMap.get('id');
    if (this.taskId) {
      this.getFiles();
    }
  }

  onFileSelected(event: any): void {
    this.selectedFile = event.target.files[0];
  }

  uploadFile(): void {
    if (this.selectedFile && this.taskId) {
      const reader = new FileReader();

      reader.onload = () => {
        const fileContent = (reader.result as string).split(',')[1];
        // Extract base64 content

        // @ts-ignore
        this.taskService.uploadFile(this.taskId, this.selectedFile.name, fileContent).subscribe(
          () => {
            alert('File uploaded successfully!');
            this.getFiles();
          },
          (error) => {
            console.error('Error uploading file:', error);
            alert('Failed to upload file.');
          }
        );
      };

      reader.onerror = (error) => {
        console.error('Error reading file:', error);
        alert('Failed to process file.');
      };

      reader.readAsDataURL(this.selectedFile);
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
        const blob = response.body as Blob; // Access the Blob body
        const fileName = response.headers.get('filename') || 'file'; // Get the filename from headers

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
