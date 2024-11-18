export class UserFP{
    id:string
    username:string;
    role:string;


    constructor(
        id: string,
        username: string,
        role:string

    ) {
        this.id = id;
        this.username = username;
        this.role = role;
    }
}