export class User {
  id: number;
  username: string;
  password: string;
  firstname: string;
  lastname: string;
  email: string;
  role: string;

  constructor(
    id: number,
    username: string,
    password: string,
    firstname: string,
    lastname: string,
    email: string,
    role: string
  ) {
    this.id = id;
    this.username = username;
    this.password = password;
    this.firstname = firstname;
    this.lastname = lastname;
    this.email = email;
    this.role = role;
  }
}
