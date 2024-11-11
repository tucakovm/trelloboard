export class User{
    id:number;
    username:string;
    password:string;
    firstname:string;
    lastname:string;
    email:string;

    constructor(
        id: number,
        username: string,
        password: string,
        firstName: string,
        lastName: string,
        email: string
    ) {
        this.id = id;
        this.username = username;
        this.password = password;
        this.firstname = firstName;
        this.lastname = lastName;
        this.email = email;
    }
}