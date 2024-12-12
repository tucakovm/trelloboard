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
        // The result here is an ArrayBuffer (raw bytes)
        const fileContent = new Uint8Array(reader.result as ArrayBuffer);

        // Prepare the FormData object to send the raw binary data
        const formData = new FormData();
        // @ts-ignore
        formData.append('taskId', this.taskId);
        // @ts-ignore
        formData.append('fileName', this.selectedFile.name);
        formData.append('fileContent', new Blob([fileContent])); // Append the raw byte content as Blob

        // Send the FormData to the backend
        this.taskService.uploadFile(formData).subscribe(
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

      // Read the file as an ArrayBuffer (raw binary data)
      reader.readAsArrayBuffer(this.selectedFile);
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
