# Project Management B2B App

A full-stack project management application built with Go backend and React frontend, designed for B2B companies with multi-tenant organization support via Scalekit.

## Features

- **Multi-tenant Architecture**: Organizations and users managed by Scalekit
- **Project Management**: Create, manage, and track projects with priorities and status
- **Task Management**: Assign tasks to projects and users with full CRUD operations
- **Real-time Dashboard**: Live metrics and recent activity
- **Role-based Access**: Integrated with Scalekit for user permissions
- **Cloud-Ready**: Optimized for GCP Cloud Run deployment

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 18+
- npm or yarn
- PostgreSQL database (local or Cloud SQL)
- Scalekit account for authentication

### Setup

1. **Clone and install dependencies:**
   ```bash
   git clone <repository-url>
   cd coffee-desk-demo
   go mod download
   cd web && npm install && cd ..
   ```

2. **Set up environment variables:**
   Create a `.env` file in the root directory:
   ```env
   # Scalekit Configuration
   SCALEKIT_ENVIRONMENT_URL=your_scalekit_environment_url
   SCALEKIT_CLIENT_ID=your_client_id
   SCALEKIT_CLIENT_SECRET=your_client_secret
   
   # Database Configuration
   DATABASE_URL=postgresql://username:password@localhost:5432/project_management?sslmode=disable
   
   # Server Configuration
   PORT=8080
   ```

3. **Set up PostgreSQL database:**
   
   **Local Development:**
   ```bash
   # Install PostgreSQL locally
   # Create database
   createdb project_management
   
   # Or use Docker
   docker run --name postgres-pm -e POSTGRES_PASSWORD=password -e POSTGRES_DB=project_management -p 5432:5432 -d postgres:15
   ```
   
   **Cloud SQL (Production):**
   - Create a Cloud SQL PostgreSQL instance
   - Use the connection string format:
     ```
     postgresql://username:password@/dbname?host=/cloudsql/project:region:instance
     ```

4. **Build and run:**
   ```bash
   # Build the application
   make build
   
   # Run the application
   make run
   ```

   The application will be available at `http://localhost:8080`

## 🗄️ Database Schema

The application uses PostgreSQL with the following main tables:

### Projects Table
- `id` (UUID, Primary Key)
- `organization_id` (String, Indexed) - References Scalekit organization
- `name` (String, Required)
- `description` (Text, Optional)
- `priority` (Enum: P1, P2, P3)
- `status` (Enum: Backlog, Todo, InProgress, Done)
- `owner_id` (String, Optional) - References Scalekit user
- `created_at`, `updated_at` (Timestamps)

### Tasks Table
- `id` (UUID, Primary Key)
- `organization_id` (String, Indexed) - References Scalekit organization
- `project_id` (UUID, Optional) - Foreign key to projects
- `name` (String, Required)
- `description` (Text, Optional)
- `priority` (Enum: P1, P2, P3)
- `status` (Enum: Backlog, Todo, InProgress, Done)
- `assignee_id` (String, Optional) - References Scalekit user
- `created_at`, `updated_at` (Timestamps)

## 🚀 Deployment

### GCP Cloud Run

1. **Build and push Docker image:**
   ```bash
   # Build the Docker image
   docker build -t gcr.io/PROJECT_ID/project-management-app .
   
   # Push to Google Container Registry
   docker push gcr.io/PROJECT_ID/project-management-app
   ```

2. **Deploy to Cloud Run:**
   ```bash
   gcloud run deploy project-management-app \
     --image gcr.io/PROJECT_ID/project-management-app \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars="DATABASE_URL=postgresql://username:password@/dbname?host=/cloudsql/project:region:instance"
   ```

3. **Set up Cloud SQL connection:**
   - Enable Cloud SQL Admin API
   - Create a Cloud SQL instance
   - Configure the connection string in Cloud Run environment variables

## 📚 API Endpoints

### Projects
- `GET /api/projects` - List all projects
- `POST /api/projects` - Create new project
- `GET /api/projects/:id` - Get project details
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project

### Tasks
- `GET /api/tasks` - List all tasks (with optional filters)
- `POST /api/tasks` - Create new task
- `GET /api/tasks/:id` - Get task details
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task

All endpoints require authentication via Scalekit and are organization-scoped.



## 📁 Project Structure

```
coffee-desk-demo/
├── database/             # Database models and connection
│   ├── database.go       # Connection management
│   └── models.go         # GORM models
├── handlers/             # API handlers
│   ├── projects.go       # Project CRUD operations
│   ├── tasks.go          # Task CRUD operations
│   ├── users.go          # User management (Scalekit)
│   ├── workspace.go      # Workspace management (Scalekit)
│   └── ...               # Other handlers
├── internal/templates/   # Embedded React templates
├── web/                  # React frontend
│   ├── src/
│   │   ├── api/          # API client functions
│   │   ├── components/   # Reusable UI components
│   │   ├── pages/        # Page components
│   │   └── hooks/        # Custom React hooks
│   └── package.json
├── config/               # Configuration management
├── scripts/              # Build scripts
├── main.go              # Application entry point
├── go.mod               # Go dependencies
├── Dockerfile           # Container configuration
└── README.md            # This file
```

## 🏗️ Architecture

- **Backend**: Go with Gin framework, GORM ORM, PostgreSQL database
- **Frontend**: React with TypeScript, Vite build system, Tailwind CSS
- **Authentication**: Scalekit for multi-tenant user management
- **Database**: PostgreSQL with auto-migrations via GORM
- **Deployment**: Docker container optimized for GCP Cloud Run
- **API Routes**: All API endpoints are under `/api/*`
- **Static Assets**: CSS, JS, and other assets are served from `/assets/*`

## 🔧 Development

### Adding New Features

1. **Backend API:**
   - Add handlers in `handlers/` directory
   - Update models in `database/models.go` if needed
   - Register routes in `main.go`

2. **Frontend:**
   - Add pages in `web/src/pages/`
   - Create components in `web/src/components/`
   - Add API functions in `web/src/api/`
   - Update routes in `web/src/App.tsx`

### Database Migrations

The application uses GORM auto-migration. When you add new fields to models:

1. Update the model struct in `database/models.go`
2. Restart the application - migrations will run automatically
3. For production, consider using explicit migration files

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `SCALEKIT_ENVIRONMENT_URL` | Scalekit environment URL | Yes |
| `SCALEKIT_CLIENT_ID` | Scalekit client ID | Yes |
| `SCALEKIT_CLIENT_SECRET` | Scalekit client secret | Yes |
| `SCALEKIT_WEBHOOK_SECRET` | Scalekit Webhook Signing Secret for SCIM Webhooks | Yes
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `PORT` | Server port (default: 8080) | No |



### Adding New Dependencies

#### React Dependencies
```bash
cd web
npm install <package-name>
```

#### Go Dependencies
```bash
go get <package-name>
```


