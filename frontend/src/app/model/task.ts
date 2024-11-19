import { Status } from "./status";

export class Task{
    id : string;
    name: string;
    description: string;
    status: Status;
    project_id?: string

  constructor(
    id : string,
    name: string,
    description: string,
    status: Status,
    projectId: string
  ) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.status = status;
        this.project_id = projectId;
    }

    get statusText(): string {
      return Status[this.status];
  }

}
