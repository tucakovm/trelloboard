import { User } from "./user";
import { UserFP } from "./userForProject";

export class Project {
    id : number;
    Name: string;
    CompletionDate: Date;  
    MinMembers: number;
    MaxMembers: number;
    Manager:UserFP;

    constructor(
        id : number,
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
        manager : UserFP
    ) {
        this.id = id;
        this.Name = name;
        this.CompletionDate = completionDate;
        this.MinMembers = minMembers;
        this.MaxMembers = maxMembers;
        this.Manager = manager
    }
}