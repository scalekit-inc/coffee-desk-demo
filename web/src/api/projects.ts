import { config } from '@/config';

export interface User {
  id: string;
  external_id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  priority: 'P1' | 'P2' | 'P3';
  status: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  owner_id?: string;
  owner?: User;
  created_at: string;
  updated_at: string;
  tasks?: Task[];
}

export interface Task {
  id: string;
  organization_id: string;
  project_id?: string;
  name: string;
  description?: string;
  priority: 'P1' | 'P2' | 'P3';
  status: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  assignee_id?: string;
  assignee?: User;
  created_at: string;
  updated_at: string;
  project?: Project;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  priority: 'P1' | 'P2' | 'P3';
  status: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  owner_id?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  priority?: 'P1' | 'P2' | 'P3';
  status?: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  owner_id?: string;
}

export interface CreateTaskRequest {
  name: string;
  description?: string;
  priority: 'P1' | 'P2' | 'P3';
  status: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  project_id?: string;
  assignee_id?: string;
}

export interface UpdateTaskRequest {
  name?: string;
  description?: string;
  priority?: 'P1' | 'P2' | 'P3';
  status?: 'Backlog' | 'Todo' | 'InProgress' | 'Done';
  project_id?: string;
  assignee_id?: string;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ProjectsResponse {
  projects: Project[];
  pagination: PaginationInfo;
}

export interface TasksResponse {
  tasks: Task[];
  pagination: PaginationInfo;
}

// Helper function to get auth headers
const getAuthHeaders = () => {
  return {
    'Content-Type': 'application/json',
    // Cookies are automatically sent with requests
  };
};

// Projects API
export const projectsApi = {
  // Get all projects
  async getProjects(page = 1, limit = 20): Promise<ProjectsResponse> {
    const response = await fetch(`${config.backendUrl}/api/projects?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch projects: ${response.statusText}`);
    }

    return response.json();
  },

  // Get single project
  async getProject(id: string): Promise<{ project: Project }> {
    const response = await fetch(`${config.backendUrl}/api/projects/${id}`, {
      method: 'GET',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch project: ${response.statusText}`);
    }

    return response.json();
  },

  // Create project
  async createProject(data: CreateProjectRequest): Promise<{ message: string; project: Project }> {
    const response = await fetch(`${config.backendUrl}/api/projects`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Failed to create project: ${response.statusText}`);
    }

    return response.json();
  },

  // Update project
  async updateProject(id: string, data: UpdateProjectRequest): Promise<{ message: string; project: Project }> {
    const response = await fetch(`${config.backendUrl}/api/projects/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Failed to update project: ${response.statusText}`);
    }

    return response.json();
  },

  // Delete project
  async deleteProject(id: string): Promise<{ message: string }> {
    const response = await fetch(`${config.backendUrl}/api/projects/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to delete project: ${response.statusText}`);
    }

    return response.json();
  },
};

// Tasks API
export const tasksApi = {
  // Get all tasks
  async getTasks(page = 1, limit = 20, filters?: {
    project_id?: string;
    status?: string;
    assignee_id?: string;
  }): Promise<TasksResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });

    if (filters?.project_id) params.append('project_id', filters.project_id);
    if (filters?.status) params.append('status', filters.status);
    if (filters?.assignee_id) params.append('assignee_id', filters.assignee_id);

    const response = await fetch(`${config.backendUrl}/api/tasks?${params}`, {
      method: 'GET',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch tasks: ${response.statusText}`);
    }

    return response.json();
  },

  // Get single task
  async getTask(id: string): Promise<{ task: Task }> {
    const response = await fetch(`${config.backendUrl}/api/tasks/${id}`, {
      method: 'GET',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch task: ${response.statusText}`);
    }

    return response.json();
  },

  // Create task
  async createTask(data: CreateTaskRequest): Promise<{ message: string; task: Task }> {
    const response = await fetch(`${config.backendUrl}/api/tasks`, {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Failed to create task: ${response.statusText}`);
    }

    return response.json();
  },

  // Update task
  async updateTask(id: string, data: UpdateTaskRequest): Promise<{ message: string; task: Task }> {
    const response = await fetch(`${config.backendUrl}/api/tasks/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Failed to update task: ${response.statusText}`);
    }

    return response.json();
  },

  // Delete task
  async deleteTask(id: string): Promise<{ message: string }> {
    const response = await fetch(`${config.backendUrl}/api/tasks/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to delete task: ${response.statusText}`);
    }

    return response.json();
  },
};
