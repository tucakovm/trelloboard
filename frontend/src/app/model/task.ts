import { Status } from "./status";

export class Task{
    Name: string;
    Description: string;
    Status: Status;
    public project_id?: string

  constructor(
    name: string,
    description: string,
    status: Status,
    projectId: string
  ) {
        this.Name = name;
        this.Description = description;
        this.Status = status;
        this.project_id = projectId;
    }

}
