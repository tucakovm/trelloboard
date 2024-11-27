export class Notification {
  notId : string;
  createdAt: Date;
  userId: string;
  message: string;
  status: boolean;

  constructor(
    notId : string,
    createdAt: Date,
    userId: string,
    message: string,
    status: boolean,

  ) {
    this.notId = notId;
    this.createdAt = createdAt;
    this.userId = userId;
    this.message = message;
    this.status = status;

  }
}
