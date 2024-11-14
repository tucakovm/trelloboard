import { Status } from "./status";

export class Task{
    name: string;
    description: string;
    status: Status;
    project_id?: string

  constructor(
    name: string,
    description: string,
    status: Status,
    projectId: string
  ) {
        this.name = name;
        this.description = description;
        this.status = status;
        this.project_id = projectId;
    }

}
