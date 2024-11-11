export class User {
  id: number;
  username: string;
  password: string;
  firstName: string;
  lastName: string;
  email: string;
  role: string;

  constructor(
    id: number,
    username: string,
    password: string,
    firstName: string,
    lastName: string,
    email: string,
    role: string
  ) {
    this.id = id;
    this.username = username;
    this.password = password;
    this.firstName = firstName;
    this.lastName = lastName;
    this.email = email;
    this.role = role;
  }
}
