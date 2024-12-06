
import { UserFP } from "./userForProject";

export class Project {
    id : string;
    name: string;
    completionDate: Date;
    minMembers: number;
    maxMembers: number;
    manager:UserFP;
    members:UserFP[];
    tasks?: Task[];

    constructor(
        id : string,
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
        manager : UserFP,
        members:UserFP[]
    ) {
        this.id = id;
        this.name = name;
        this.completionDate = completionDate;
        this.minMembers = minMembers;
        this.maxMembers = maxMembers;
        this.manager = manager
        this.members = members;
    }
}
export interface Task {
  id: string;
  name: string;
  description: string;
  dependencies: string[];
  blocked: boolean;
}

