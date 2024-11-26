import { Status } from "./status";
import {UserFP} from "./userForProject";

export class Task{
    id : string | null;
    name: string;
    description: string;
    status: string;
    project_id?: string;
  members:UserFP[];

  constructor(
    id : string | null,
    name: string,
    description: string,
    status: string,
    projectId: string,
    members:UserFP[],
  ) {
        this.id = id;
        this.name = name;
        this.description = description;
        this.status = status;
        this.project_id = projectId;
        this.members = members;
    }

  //   get statusText(): string {
  //     return Status[this.status];
  // }

}
