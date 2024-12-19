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
  files: string[] = [];

  constructor(
    private activatedRoute: ActivatedRoute,
    private taskService: TaskService,
    private router: Router
  ) {
  }

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
        // Base64 encode the file content
        const fileContentBase64 = btoa(String.fromCharCode(...new Uint8Array(reader.result as ArrayBuffer)));

        // Create the request payload


        const payload = {
          taskId: this.taskId,
          fileContent: fileContentBase64,
          fileName: this.selectedFile?.name,
        };

        // Send the payload to the backend
        this.taskService.uploadFile(payload).subscribe(
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

      reader.readAsArrayBuffer(this.selectedFile);
    }
  }

  getFiles(): void {
    if (this.taskId) {
      this.taskService.getFiles(this.taskId).subscribe(
        (response: any) => {
          console.log('API Response:', response);

          if (response?.fileNames && Array.isArray(response.fileNames)) {
            this.files = response.fileNames;
            console.log(this.files)
          } else {
            console.warn('Unexpected API response format:', response);
            this.files = [];
          }
        },
        (error) => {
          console.error('Error fetching files:', error);
          alert('Failed to fetch files.');
        }
      );
    }
  }

  downloadFile(fileName: string): void {
    if (this.taskId) {
      this.taskService.downloadFile(this.taskId, fileName).subscribe(
        (response: any) => {
          const fileContent = response.fileId;
          if (fileContent) {
            // Create a Blob with the file content
            const blob = new Blob([fileContent], { type: 'text/plain' });

            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = "(" + this.taskId + ")" +  fileName  || 'download.txt';
            link.click();
          } else {
            console.error('Invalid response structure or missing fileId');
            alert('Failed to download file.');
          }
        },
        (error) => {
          console.error('Error downloading file:', error);
          alert('Failed to download file.');
        }
      );
    } else {
      console.error('Task ID is not available.');
    }
  }

  deleteFile(fileName: String): void{
    console.log("delete bttn")
    if (this.taskId){
      console.log("what")
      this.taskService.deleteFile(this.taskId, fileName).subscribe(
        (response: any)=>{
          console.log(response)
          this.getFiles()
        }
      )
    }
  }


}
