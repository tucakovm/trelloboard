export interface TaskFW {
  id: string; // Unique identifier for the task
  name: string; // Name of the task
  description: string; // Description of the task
  dependencies: string[]; // List of task IDs this task depends on
  blocked: boolean; // Whether the task is blocked
}
