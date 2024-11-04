import { User } from "./user";

export class Project {
    Name: string;
    CompletionDate: Date;  
    MinMembers: number;
    MaxMembers: number;

    constructor(
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
    ) {
        this.Name = name;
        this.CompletionDate = completionDate;
        this.MinMembers = minMembers;
        this.MaxMembers = maxMembers;
    }
}