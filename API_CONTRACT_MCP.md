# Project Management API Contract for MCP Server Integration

## Overview

This document provides the complete API specification for integrating with the Project Management application's backend services. The API supports full CRUD operations for Projects and Tasks with organization-scoped access control.

**Data Storage**: The application maintains local copies of User and Organization data from Scalekit for better performance. User and Organization tables contain minimal data (UUID primary key, external_id for Scalekit reference, and email/name respectively) with Scalekit remaining the source of truth for detailed user and organization information.

## Base Information

- **Base URL**: `https://your-app.run.app/api` (production)
- **Base URL**: `http://localhost:8080/api` (local development)
- **Authentication**: Cookie-based session (Scalekit)
- **Content-Type**: `application/json` (for POST/PUT requests)
- **Organization Scoping**: All data is automatically filtered by the authenticated user's organization

## Authentication

All API endpoints require authentication via Scalekit session cookies. The session cookie is automatically included in requests and contains the user's organization ID for data scoping.

**Required Cookie Header:**
```
Cookie: session_token=<scalekit_session_token>
```

## Data Models

### User Object
```json
{
  "id": "uuid",
  "external_id": "string",
  "email": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Organization Object
```json
{
  "id": "uuid",
  "external_id": "string",
  "name": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Project Object
```json
{
  "id": "uuid",
  "organization_id": "string",
  "name": "string",
  "description": "string|null",
  "priority": "P1|P2|P3",
  "status": "Backlog|Todo|InProgress|Done",
  "owner_id": "string|null",
  "tasks": [Task] (only in GET /projects/:id),
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Task Object
```json
{
  "id": "uuid",
  "organization_id": "string",
  "project_id": "uuid|null",
  "name": "string",
  "description": "string|null",
  "priority": "P1|P2|P3",
  "status": "Backlog|Todo|InProgress|Done",
  "assignee_id": "string|null",
  "project": Project (only in GET /tasks/:id),
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Enums

**Priority Values:**
- `P1` - High priority
- `P2` - Medium priority  
- `P3` - Low priority

**Status Values:**
- `Backlog` - Not started
- `Todo` - Ready to start
- `InProgress` - Currently being worked on
- `Done` - Completed

## API Endpoints

### Projects

#### 1. Get All Projects
**GET** `/projects`

**Description:** Retrieve all projects for the authenticated user's organization.

**Query Parameters:**
- `page` (integer, optional): Page number for pagination (default: 1)
- `limit` (integer, optional): Number of items per page (default: 20, max: 100)

**Response:**
```json
{
  "projects": [Project],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

**Example Request:**
```bash
GET /api/projects?page=1&limit=10
```

#### 2. Create Project
**POST** `/projects`

**Description:** Create a new project in the authenticated user's organization.

**Request Body:**
```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "priority": "P1|P2|P3 (required)",
  "status": "Backlog|Todo|InProgress|Done (required)",
  "owner_id": "string (optional)"
}
```

**Response:**
```json
{
  "message": "Project created successfully",
  "project": Project
}
```

**Example Request:**
```bash
POST /api/projects
Content-Type: application/json

{
  "name": "Website Redesign",
  "description": "Complete redesign of company website",
  "priority": "P1",
  "status": "Backlog",
  "owner_id": "user_123"
}
```

#### 3. Get Project by ID
**GET** `/projects/{id}`

**Description:** Retrieve a specific project with its associated tasks.

**Path Parameters:**
- `id` (uuid, required): Project ID

**Response:**
```json
{
  "project": Project
}
```

**Example Request:**
```bash
GET /api/projects/123e4567-e89b-12d3-a456-426614174000
```

#### 4. Update Project
**PUT** `/projects/{id}`

**Description:** Update an existing project.

**Path Parameters:**
- `id` (uuid, required): Project ID

**Request Body:**
```json
{
  "name": "string (optional)",
  "description": "string (optional)",
  "priority": "P1|P2|P3 (optional)",
  "status": "Backlog|Todo|InProgress|Done (optional)",
  "owner_id": "string (optional)"
}
```

**Response:**
```json
{
  "message": "Project updated successfully",
  "project": Project
}
```

**Example Request:**
```bash
PUT /api/projects/123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json

{
  "status": "InProgress",
  "priority": "P2"
}
```

#### 5. Delete Project
**DELETE** `/projects/{id}`

**Description:** Delete a project (soft delete).

**Path Parameters:**
- `id` (uuid, required): Project ID

**Response:**
```json
{
  "message": "Project deleted successfully"
}
```

**Example Request:**
```bash
DELETE /api/projects/123e4567-e89b-12d3-a456-426614174000
```

### Tasks

#### 1. Get All Tasks
**GET** `/tasks`

**Description:** Retrieve all tasks for the authenticated user's organization with optional filtering.

**Query Parameters:**
- `page` (integer, optional): Page number for pagination (default: 1)
- `limit` (integer, optional): Number of items per page (default: 20, max: 100)
- `project_id` (uuid, optional): Filter tasks by project ID
- `status` (string, optional): Filter tasks by status
- `assignee_id` (string, optional): Filter tasks by assignee ID

**Response:**
```json
{
  "tasks": [Task],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

**Example Request:**
```bash
GET /api/tasks?project_id=123e4567-e89b-12d3-a456-426614174000&status=InProgress&page=1&limit=10
```

#### 2. Create Task
**POST** `/tasks`

**Description:** Create a new task in the authenticated user's organization.

**Request Body:**
```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "priority": "P1|P2|P3 (required)",
  "status": "Backlog|Todo|InProgress|Done (required)",
  "project_id": "uuid (optional)",
  "assignee_id": "string (optional)"
}
```

**Response:**
```json
{
  "message": "Task created successfully",
  "task": Task
}
```

**Example Request:**
```bash
POST /api/tasks
Content-Type: application/json

{
  "name": "Design homepage layout",
  "description": "Create wireframes and mockups for new homepage",
  "priority": "P1",
  "status": "Todo",
  "project_id": "123e4567-e89b-12d3-a456-426614174000",
  "assignee_id": "user_456"
}
```

#### 3. Get Task by ID
**GET** `/tasks/{id}`

**Description:** Retrieve a specific task with its associated project information.

**Path Parameters:**
- `id` (uuid, required): Task ID

**Response:**
```json
{
  "task": Task
}
```

**Example Request:**
```bash
GET /api/tasks/456e7890-e89b-12d3-a456-426614174000
```

#### 4. Update Task
**PUT** `/tasks/{id}`

**Description:** Update an existing task.

**Path Parameters:**
- `id` (uuid, required): Task ID

**Request Body:**
```json
{
  "name": "string (optional)",
  "description": "string (optional)",
  "priority": "P1|P2|P3 (optional)",
  "status": "Backlog|Todo|InProgress|Done (optional)",
  "project_id": "uuid (optional)",
  "assignee_id": "string (optional)"
}
```

**Response:**
```json
{
  "message": "Task updated successfully",
  "task": Task
}
```

**Example Request:**
```bash
PUT /api/tasks/456e7890-e89b-12d3-a456-426614174000
Content-Type: application/json

{
  "status": "Done",
  "assignee_id": "user_789"
}
```

#### 5. Delete Task
**DELETE** `/tasks/{id}`

**Description:** Delete a task (soft delete).

**Path Parameters:**
- `id` (uuid, required): Task ID

**Response:**
```json
{
  "message": "Task deleted successfully"
}
```

**Example Request:**
```bash
DELETE /api/tasks/456e7890-e89b-12d3-a456-426614174000
```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request body"
}
```
*Occurs when request validation fails*

### 401 Unauthorized
```json
{
  "error": "Unauthorized"
}
```
*Occurs when session is invalid or expired*

### 404 Not Found
```json
{
  "error": "Project not found"
}
```
*Occurs when requested resource doesn't exist*

### 500 Internal Server Error
```json
{
  "error": "Failed to fetch projects"
}
```
*Occurs when server encounters an unexpected error*

## HTTP Status Codes

- `200` - OK (successful GET, PUT, DELETE)
- `201` - Created (successful POST)
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (authentication required)
- `404` - Not Found (resource not found)
- `500` - Internal Server Error (server error)

## Rate Limiting

Currently no rate limiting is implemented. Consider implementing rate limiting for production use.

## Testing

### Health Check
**GET** `/ping`

**Response:**
```json
{
  "message": "pong"
}
```

## MCP Server Integration Notes

1. **Organization Scoping**: All data is automatically filtered by the authenticated user's organization. No need to pass organization ID in requests.

2. **Authentication**: Use the Scalekit session cookie for all requests. The session contains the user's organization ID.

3. **Data Relationships**: 
   - Projects can have multiple tasks
   - Tasks can optionally belong to a project
   - Both projects and tasks reference Scalekit user IDs for owners/assignees

4. **Pagination**: All list endpoints support pagination with `page` and `limit` parameters.

5. **Filtering**: The tasks endpoint supports filtering by `project_id`, `status`, and `assignee_id`.

6. **Soft Deletes**: Delete operations perform soft deletes (records are marked as deleted but not physically removed).

## Example MCP Tool Mappings

| MCP Tool | API Endpoint | Method |
|----------|--------------|---------|
| `get_project` | `/projects/{id}` | GET |
| `get_all_projects` | `/projects` | GET |
| `get_all_tasks` | `/tasks` | GET |
| `get_project_tasks` | `/tasks?project_id={id}` | GET |
| `get_task` | `/tasks/{id}` | GET |
| `create_project` | `/projects` | POST |
| `create_task` | `/tasks` | POST |
| `delete_project` | `/projects/{id}` | DELETE |
| `delete_task` | `/tasks/{id}` | DELETE |

## Support

For questions or issues with the API, please contact the development team or refer to the application's internal documentation.
