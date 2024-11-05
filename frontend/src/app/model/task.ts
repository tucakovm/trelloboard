import { Status } from "./status";

export class Task{
    Name: string;
    Description: string;
    Status: Status;

    constructor(
        name: string,
        description: string,
        status: Status
        
    ) {
        this.Name = name;
        this.Description = description;
        this.Status = status;
    }

}