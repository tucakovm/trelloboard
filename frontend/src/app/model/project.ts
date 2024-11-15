import { UserFP } from "./userForProject";

export class Project {
    id : number;
    name: string;
    completionDate: Date;  
    minMembers: number;
    maxMembers: number;
    manager:UserFP;
    members:UserFP[];

    constructor(
        id : number,
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