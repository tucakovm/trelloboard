import { UserFP } from "./userForProject";

export class Project {
    id : string;
    name: string;
    completionDate: Date;  
    minMembers: number;
    maxMembers: number;
    manager:UserFP;

    constructor(
        id : string,
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
        manager : UserFP
    ) {
        this.id = id;
        this.name = name;
        this.completionDate = completionDate;
        this.minMembers = minMembers;
        this.maxMembers = maxMembers;
        this.manager = manager
    }
}
