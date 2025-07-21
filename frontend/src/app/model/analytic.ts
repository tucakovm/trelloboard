export interface Analytics {
  totalTasks: number;
  statusCounts: { [key: string]: number };
  taskStatusDurations: {
    [taskId: string]: {
      taskId: string;
      statusDurations: { status: string; duration: number }[];
    };
  };
  memberTasks: {
    [memberId: string]: {
      memberId: string;
      tasks: string[];
    };
  };
  finishedEarly: boolean;
}
