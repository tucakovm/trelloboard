import { User } from "./user";

export class Project {
    Id:number|null
    Name: string;
    CompletionDate: Date;  
    MinMembers: number;
    MaxMembers: number;

    constructor(
        id:number|null,
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
    ) {
        this.Id = id;
        this.Name = name;
        this.CompletionDate = completionDate;
        this.MinMembers = minMembers;
        this.MaxMembers = maxMembers;
    }
}