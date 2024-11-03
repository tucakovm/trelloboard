import { User } from "./user";

export class Project {
    id: number | null;  
    name: string;
    completionDate: Date;  
    minMembers: number;
    maxMembers: number;
    Manager: User | null;  
    Members: User[] | null;  

    constructor(
        id: number | null,
        name: string,
        completionDate: Date,
        minMembers: number,
        maxMembers: number,
        Manager: User | null,
        Members: User[] | null
    ) {
        this.id = id;
        this.name = name;
        this.completionDate = completionDate;
        this.minMembers = minMembers;
        this.maxMembers = maxMembers;
        this.Manager = Manager;
        this.Members = Members;
    }
}