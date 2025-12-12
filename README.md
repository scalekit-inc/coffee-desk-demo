# Coffee Desk Demo

Full-featured demo app demonstrating Scalekit's auth stack including workspace creation, user provisioning, granular permission management, and enterprise-grade login methods (SAML/OIDC SSO, social providers, passwordless auth). Can be used as a project template for building B2B applications with multi-tenant support.

This application provides a complete project management system with a Go backend and React frontend, showcasing how to build production-ready B2B applications with Scalekit's authentication and authorization features.

## Features

Coffee Desk Demo includes the following features:

- **Multi-tenant Architecture**: Organizations and users are managed entirely by Scalekit, eliminating the need to build your own user management system. Each organization operates in complete isolation with its own data and users.

- **Project Management**: Create, manage, and track projects with customizable priorities (P1, P2, P3) and status tracking (Backlog, Todo, InProgress, Done). Projects are automatically scoped to the user's organization.

- **Task Management**: Assign tasks to projects and users with full CRUD operations. Tasks support the same priority and status system as projects, enabling consistent workflow management.

- **Real-time Dashboard**: View live metrics and recent activity across your organization. The dashboard provides an overview of project and task statuses at a glance.

- **Role-based Access Control**: Integrated with Scalekit for granular permission management. Control who can create projects, assign tasks, and manage workspaces through Scalekit's permission system.

- **Cloud-Ready Deployment**: Optimized for GCP Cloud Run deployment with Docker containerization. The application is designed to scale horizontally and handle production workloads.

## Compatibility

Coffee Desk Demo requires the following software and services:

| Requirement      | Version | Notes                                           |
| ---------------- | ------- | ----------------------------------------------- |
| Go               | 1.24+   | Backend runtime and build tool                  |
| Node.js          | 18+     | Frontend build tool and development server      |
| npm or yarn      | Latest  | Package manager for frontend dependencies       |
| PostgreSQL       | 12+     | Database (local or Cloud SQL)                   |
| Scalekit Account | -       | Required for authentication and user management |
| Docker           | Latest  | Optional, for containerized deployment          |

The application has been tested on Linux, macOS, and Windows. For production deployments, Linux containers are recommended.

## Installing

Install Coffee Desk Demo by cloning the repository and setting up the required dependencies.

### Clone the repository

Clone the repository to your local machine:

```bash
git clone <repository-url>
cd coffee-desk-demo
```

### Install dependencies

Install both backend and frontend dependencies:

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd web && npm install && cd ..
```

You can also use the Makefile to install all dependencies:

```bash
make install
```

### Set up environment variables

Create a `.env` file in the root directory with your configuration:

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

The `SCALEKIT_ENVIRONMENT_URL` is your Scalekit environment endpoint. The `SCALEKIT_CLIENT_ID` and `SCALEKIT_CLIENT_SECRET` are obtained from your Scalekit dashboard when you create an application.

### Set up PostgreSQL database

Set up a PostgreSQL database for local development or production use.

#### Local development

For local development, you can either install PostgreSQL directly or use Docker:

**Option 1: Install PostgreSQL locally**

```bash
# Install PostgreSQL (varies by OS)
# macOS: brew install postgresql
# Ubuntu: sudo apt-get install postgresql

# Create the database
createdb project_management
```

**Option 2: Use Docker**

```bash
docker run --name postgres-pm \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=project_management \
  -p 5432:5432 \
  -d postgres:15
```

The Docker approach is recommended for quick setup and easy cleanup.

#### Cloud SQL (production)

For production deployments on GCP Cloud Run:

1. Create a Cloud SQL PostgreSQL instance in the Google Cloud Console
2. Use the connection string format:
   ```
   postgresql://username:password@/dbname?host=/cloudsql/project:region:instance
   ```
3. Configure the connection string in Cloud Run environment variables

### Build and run

Build the application using the Makefile:

```bash
# Build the application (builds frontend and backend)
make build

# Run the application
make run
```

The application will be available at `http://localhost:8080`.

For development, you can run the frontend and backend separately:

```bash
# Terminal 1: Run frontend dev server
make dev-frontend

# Terminal 2: Run backend server
make dev-backend
```

This allows hot-reloading for frontend changes while the backend runs separately.

## Tutorial

This tutorial guides you through using Coffee Desk Demo to manage projects and tasks.

### First time setup

After installing and running the application, you'll need to authenticate with Scalekit.

1. Navigate to `http://localhost:8080` in your browser
2. Click the login button to authenticate with Scalekit
3. If this is your first time, you'll be prompted to create a workspace
4. Complete the onboarding flow to set up your organization

Once authenticated, you'll see the dashboard with an overview of your workspace.

### Create your first project

Projects organize related tasks and provide a way to track work across your organization.

1. Navigate to the **Projects** page from the sidebar
2. Click the **Create Project** button
3. Fill in the project details:
   - **Name**: Enter a descriptive project name
   - **Description**: Add details about the project's purpose
   - **Priority**: Select P1 (high), P2 (medium), or P3 (low)
   - **Status**: Choose Backlog, Todo, InProgress, or Done
4. Click **Create** to save the project

The project appears in your projects list and is automatically associated with your organization.

### Create and assign tasks

Tasks represent individual work items that can be assigned to projects and team members.

1. Navigate to the **Tasks** page from the sidebar
2. Click the **Create Task** button
3. Fill in the task details:
   - **Name**: Enter a clear task name
   - **Description**: Add detailed information about the task
   - **Project**: Optionally link the task to a project
   - **Priority**: Set the task priority
   - **Status**: Set the initial status
   - **Assignee**: Select a team member to assign the task
4. Click **Create** to save the task

Tasks linked to projects appear in the project detail view. You can filter tasks by project, status, priority, or assignee.

### Navigate the dashboard

The dashboard provides an overview of your organization's activity.

1. View **Recent Activity**: See the latest projects and tasks created or updated
2. Check **Metrics**: View counts of projects and tasks by status
3. Access **Quick Actions**: Create new projects or tasks directly from the dashboard

The dashboard updates in real-time as you and your team make changes.

### Manage workspaces

Workspaces represent organizations in Scalekit. Each workspace has its own projects, tasks, and users.

1. Navigate to the **Workspace** page from the sidebar
2. View workspace details including member count and settings
3. Invite users to your workspace using the **Invite User** button
4. Manage user roles and permissions through Scalekit's permission system

All data is automatically scoped to your current workspace, ensuring complete isolation between organizations.

## More features

Coffee Desk Demo includes additional features for production use.

### API endpoints

The application provides RESTful API endpoints for programmatic access. All endpoints require authentication via Scalekit and are organization-scoped.

#### Projects

| Method | Endpoint            | Description                                   |
| ------ | ------------------- | --------------------------------------------- |
| GET    | `/api/projects`     | List all projects in the current organization |
| POST   | `/api/projects`     | Create a new project                          |
| GET    | `/api/projects/:id` | Get project details by ID                     |
| PUT    | `/api/projects/:id` | Update an existing project                    |
| DELETE | `/api/projects/:id` | Delete a project                              |

#### Tasks

| Method | Endpoint         | Description                                                            |
| ------ | ---------------- | ---------------------------------------------------------------------- |
| GET    | `/api/tasks`     | List all tasks (supports filters: project, status, priority, assignee) |
| POST   | `/api/tasks`     | Create a new task                                                      |
| GET    | `/api/tasks/:id` | Get task details by ID                                                 |
| PUT    | `/api/tasks/:id` | Update an existing task                                                |
| DELETE | `/api/tasks/:id` | Delete a task                                                          |

All API requests must include a valid Scalekit authentication token in the `Authorization` header. Responses are JSON-formatted.

### Database schema

The application uses PostgreSQL with the following main tables. All tables include `created_at` and `updated_at` timestamps.

#### Projects table

| Column            | Type             | Description                         |
| ----------------- | ---------------- | ----------------------------------- |
| `id`              | UUID             | Primary key                         |
| `organization_id` | String (Indexed) | References Scalekit organization    |
| `name`            | String           | Required project name               |
| `description`     | Text             | Optional project description        |
| `priority`        | Enum             | P1, P2, or P3                       |
| `status`          | Enum             | Backlog, Todo, InProgress, or Done  |
| `owner_id`        | String           | Optional reference to Scalekit user |

#### Tasks table

| Column            | Type             | Description                         |
| ----------------- | ---------------- | ----------------------------------- |
| `id`              | UUID             | Primary key                         |
| `organization_id` | String (Indexed) | References Scalekit organization    |
| `project_id`      | UUID             | Optional foreign key to projects    |
| `name`            | String           | Required task name                  |
| `description`     | Text             | Optional task description           |
| `priority`        | Enum             | P1, P2, or P3                       |
| `status`          | Enum             | Backlog, Todo, InProgress, or Done  |
| `assignee_id`     | String           | Optional reference to Scalekit user |

The database uses GORM for object-relational mapping and automatically handles migrations on application startup.

### Project structure

The codebase is organized into clear directories for maintainability:

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

### Architecture

Coffee Desk Demo uses a modern full-stack architecture:

- **Backend**: Go with Gin framework for HTTP routing, GORM ORM for database operations, and PostgreSQL for data persistence. The backend serves both API endpoints and embedded frontend assets.

- **Frontend**: React with TypeScript for type safety, Vite for fast development and building, and Tailwind CSS for styling. The frontend is built and embedded into the Go binary for single-binary deployment.

- **Authentication**: Scalekit handles all authentication and authorization. The application integrates with Scalekit's API to verify user sessions and manage permissions.

- **Database**: PostgreSQL with auto-migrations via GORM. Database schema changes are automatically applied when models are updated.

- **Deployment**: Docker container optimized for GCP Cloud Run. The application can be deployed as a single container with all dependencies included.

- **API Routes**: All API endpoints are under `/api/*`. Static assets (CSS, JS, images) are served from `/assets/*`.

## Configuring

Configure Coffee Desk Demo using environment variables and Scalekit settings.

### Environment variables

Set the following environment variables in your `.env` file or deployment environment:

| Variable                   | Description                  | Required |
| -------------------------- | ---------------------------- | -------- |
| `SCALEKIT_ENVIRONMENT_URL` | Scalekit environment URL     | Yes      |
| `SCALEKIT_CLIENT_ID`       | Scalekit client ID           | Yes      |
| `SCALEKIT_CLIENT_SECRET`   | Scalekit client secret       | Yes      |
| `DATABASE_URL`             | PostgreSQL connection string | Yes      |
| `PORT`                     | Server port (default: 8080)  | No       |

The `PORT` variable defaults to `8080` if not specified. For production deployments, set this to match your hosting provider's requirements.

### Scalekit configuration

Configure Scalekit settings in your Scalekit dashboard:

1. Create a new application in your Scalekit environment
2. Set the redirect URI to `http://localhost:8080/callback` for local development
3. For production, set the redirect URI to your production domain: `https://yourdomain.com/callback`
4. Copy the client ID and client secret to your environment variables
5. Configure authentication methods (SAML, OIDC, social providers) as needed

The application automatically handles Scalekit authentication flows and session management.

# Development

Develop new features by following the established patterns in the codebase.

### Adding new features

Add features by extending the existing handlers and components.

#### Backend API

1. Add new handlers in the `handlers/` directory following the existing pattern
2. Update models in `database/models.go` if you need new database tables or fields
3. Register new routes in `main.go` under the appropriate API group

All handlers should:

- Verify Scalekit authentication using the session middleware
- Scope operations to the user's organization
- Return appropriate HTTP status codes and JSON responses

#### Frontend

1. Add new pages in `web/src/pages/` for major features
2. Create reusable components in `web/src/components/` for shared UI elements
3. Add API client functions in `web/src/api/` to communicate with the backend
4. Update routes in `web/src/App.tsx` to include new pages

The frontend uses React Router for navigation and Axios for API calls.

### Database migrations

The application uses GORM auto-migration. When you add new fields to models:

1. Update the model struct in `database/models.go`
2. Restart the application - migrations run automatically on startup
3. For production, consider using explicit migration files for better control

GORM automatically creates tables and adds columns, but does not remove columns or modify existing data. Plan migrations carefully for production deployments.

### Adding new dependencies

Add dependencies using the standard package managers.

#### React dependencies

```bash
cd web
npm install <package-name>
```

#### Go dependencies

```bash
go get <package-name>
go mod tidy
```

Always run `go mod tidy` after adding Go dependencies to clean up the `go.mod` file.

### Development workflow

Use the separate development servers for faster iteration:

1. Start the frontend dev server: `make dev-frontend`
2. Start the backend server: `make dev-backend`
3. Make changes to either frontend or backend code
4. Frontend changes hot-reload automatically
5. Backend changes require restarting the server

For production builds, use `make build` to create a single binary with embedded frontend assets.

## Deployment

Deploy Coffee Desk Demo to production using Docker and GCP Cloud Run.

### GCP Cloud Run

Deploy to Google Cloud Run for serverless container hosting.

#### Build and push Docker image

Build the Docker image and push it to Google Container Registry:

```bash
# Build the Docker image
docker build -t gcr.io/PROJECT_ID/project-management-app .

# Push to Google Container Registry
docker push gcr.io/PROJECT_ID/project-management-app
```

Replace `PROJECT_ID` with your Google Cloud project ID.

#### Deploy to Cloud Run

Deploy the container to Cloud Run:

```bash
gcloud run deploy project-management-app \
  --image gcr.io/PROJECT_ID/project-management-app \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars="DATABASE_URL=postgresql://username:password@/dbname?host=/cloudsql/project:region:instance,SCALEKIT_ENVIRONMENT_URL=your_url,SCALEKIT_CLIENT_ID=your_id,SCALEKIT_CLIENT_SECRET=your_secret"
```

Set all required environment variables in the deployment command or through the Cloud Run console.

#### Set up Cloud SQL connection

Connect to Cloud SQL for database access:

1. Enable the Cloud SQL Admin API in your Google Cloud project
2. Create a Cloud SQL PostgreSQL instance
3. Configure the connection string using the format:
   ```
   postgresql://username:password@/dbname?host=/cloudsql/project:region:instance
   ```
4. Set the connection string in Cloud Run environment variables
5. Ensure the Cloud Run service has the Cloud SQL Client role

The application automatically connects to Cloud SQL using the Unix socket path specified in the connection string.

### Other deployment options

Coffee Desk Demo can be deployed to any platform that supports Docker containers:

- **AWS ECS/Fargate**: Use the Docker image with ECS task definitions
- **Azure Container Instances**: Deploy the container directly
- **Kubernetes**: Use the Docker image with Kubernetes deployments
- **Self-hosted**: Run the Docker container on any server with Docker installed

Ensure your deployment platform provides:

- PostgreSQL database access
- Environment variable configuration
- HTTPS support for Scalekit callbacks
- Sufficient memory and CPU for the application

## Troubleshooting

Common issues and solutions when running Coffee Desk Demo.

<details>
<summary><strong>Database connection errors</strong></summary>

**Problem**: Application fails to connect to PostgreSQL.

**Solutions**:

- Verify the `DATABASE_URL` environment variable is set correctly
- Check that PostgreSQL is running: `pg_isready` or `docker ps` for Docker containers
- Ensure the database exists: `psql -l` to list databases
- Verify network connectivity if using a remote database
- Check PostgreSQL logs for connection attempts

</details>

<details>
<summary><strong>Scalekit authentication errors</strong></summary>

**Problem**: Users cannot log in or authentication fails.

**Solutions**:

- Verify `SCALEKIT_ENVIRONMENT_URL`, `SCALEKIT_CLIENT_ID`, and `SCALEKIT_CLIENT_SECRET` are set correctly
- Check that the redirect URI in Scalekit matches your application URL (including `/callback`)
- Ensure your Scalekit environment is active and accessible
- Check browser console for JavaScript errors during authentication flow
- Verify CORS settings in Scalekit if accessing from a different domain

</details>

<details>
<summary><strong>Frontend build errors</strong></summary>

**Problem**: `make build` fails with frontend build errors.

**Solutions**:

- Ensure Node.js 18+ is installed: `node --version`
- Clear node_modules and reinstall: `cd web && rm -rf node_modules && npm install`
- Check for TypeScript errors: `cd web && npm run build`
- Verify all dependencies are installed: `cd web && npm install`
- Check `web/package.json` for version conflicts

</details>

<details>
<summary><strong>Port already in use</strong></summary>

**Problem**: Application cannot start because port 8080 is already in use.

**Solutions**:

- Change the `PORT` environment variable to a different port
- Find and stop the process using port 8080: `lsof -i :8080` (macOS/Linux) or `netstat -ano | findstr :8080` (Windows)
- Use a different port for development: `PORT=3000 make run`

</details>

<details>
<summary><strong>Database migration errors</strong></summary>

**Problem**: Database schema changes are not applied.

**Solutions**:

- Check database connection is working
- Verify GORM models match the expected schema
- Review application logs for migration errors
- Manually inspect database schema: `psql project_management -c "\d projects"`
- For production, consider using explicit migration files instead of auto-migration

</details>

<details>
<summary><strong>CORS errors in development</strong></summary>

**Problem**: Frontend cannot make API requests due to CORS errors.

**Solutions**:

- Ensure you're using the development workflow with separate frontend and backend servers
- Check that the frontend dev server is running on a different port than the backend
- Verify CORS settings in `main.go` allow your frontend origin
- Use the embedded frontend build (`make build && make run`) instead of separate servers

</details>

For additional help, check the application logs, open an issue in the repository, or [join the Scalekit community on Slack](https://join.slack.com/t/scalekit-community/shared_invite/zt-3gsxwr4hc-0tvhwT2b_qgVSIZQBQCWRw) to ask questions and get support.
