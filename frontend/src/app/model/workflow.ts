
export interface TaskNode {
  taskID: string;
  taskName: string;
  taskDescription: string;
  dependencies: string[];  // Array of dependent task IDs
  blocked: boolean;
}

export interface Workflow {
  projectID: string;
  projectName: string;
  tasks: TaskNode[];
}
