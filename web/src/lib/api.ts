// API types and client for mini-kanban

export interface Project {
  id: number;
  name: string;
}

export interface Task {
  id: number;
  title: string;
  body: string;
  status: 'todo' | 'doing' | 'done';
  dueAt?: number;
  tags: string[];
  createdAt: number;
  updatedAt: number;
  isOverdue: boolean;
}

export interface KanbanData {
  projects: Project[];
  currentProject: string;
  todoTasks: Task[];
  doingTasks: Task[];
  doneTasks: Task[];
  tags: string[];
}

const API_BASE = '/api';

export async function fetchKanban(project?: string): Promise<KanbanData> {
  const params = project ? `?project=${encodeURIComponent(project)}` : '';
  const res = await fetch(`${API_BASE}/kanban${params}`);
  if (!res.ok) throw new Error('Failed to fetch kanban data');
  return res.json();
}

export async function fetchProjects(): Promise<Project[]> {
  const res = await fetch(`${API_BASE}/projects`);
  if (!res.ok) throw new Error('Failed to fetch projects');
  return res.json();
}

export interface CreateTaskParams {
  project: string;
  title: string;
  tags?: string[];
  dueAt?: string;
}

export async function createTask(params: CreateTaskParams): Promise<Task> {
  const res = await fetch(`${API_BASE}/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) throw new Error('Failed to create task');
  return res.json();
}

export interface UpdateTaskParams {
  project: string;
  title?: string;
  tags?: string[];
  dueAt?: string;
}

export async function updateTask(id: number, params: UpdateTaskParams): Promise<void> {
  const res = await fetch(`${API_BASE}/tasks/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (!res.ok) throw new Error('Failed to update task');
}

export async function deleteTask(id: number, project: string): Promise<void> {
  const res = await fetch(`${API_BASE}/tasks/${id}?project=${encodeURIComponent(project)}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete task');
}

export async function updateTaskStatus(id: number, project: string, status: string): Promise<void> {
  const res = await fetch(`${API_BASE}/tasks/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ project, status }),
  });
  if (!res.ok) throw new Error('Failed to update task status');
}
