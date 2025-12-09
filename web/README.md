# Frontend Development Guide

This directory contains the React frontend for Coffee Desk Demo. For complete setup instructions, see the [main README](../README.md).

The frontend is built with React, TypeScript, Vite, and Tailwind CSS. It communicates with the Go backend through RESTful API endpoints and uses Scalekit for authentication.

## Architecture

The frontend uses a modern React architecture with TypeScript for type safety.

### Technology stack

- **React 18**: Component-based UI library
- **TypeScript**: Type-safe JavaScript
- **Vite**: Fast build tool and development server
- **Tailwind CSS**: Utility-first CSS framework
- **shadcn/ui**: Pre-built accessible UI components
- **React Router**: Client-side routing
- **Axios**: HTTP client for API requests

### Project structure

The frontend code is organized in the `src/` directory:

```
web/src/
├── api/              # API client functions
│   └── projects.ts   # Example API client
├── components/       # Reusable UI components
│   ├── ui/           # shadcn/ui components
│   ├── AppSidebar.tsx
│   ├── AuthWrapper.tsx
│   └── ...
├── pages/            # Page components (routes)
│   ├── Dashboard.tsx
│   ├── Projects.tsx
│   ├── Tasks.tsx
│   └── ...
├── hooks/            # Custom React hooks
│   ├── useAuth.tsx   # Authentication hook
│   └── use-toast.ts  # Toast notifications
├── lib/              # Utility functions
│   └── utils.ts      # Helper functions
├── config.ts         # Frontend configuration
├── App.tsx           # Main app component with routing
└── main.tsx          # Application entry point
```

# Development

Develop the frontend using the Vite development server for hot module replacement.

### Prerequisites

Before developing the frontend, ensure you have:

- Node.js 18+ installed
- npm or yarn package manager
- Backend server running (see main README for setup)

### Running the development server

Start the frontend development server:

```bash
cd web
npm install
npm run dev
```

The development server runs on `http://localhost:5173` by default (Vite's default port). The server supports hot module replacement, so changes to your code are reflected immediately in the browser.

### Running with backend

For full-stack development, run both frontend and backend:

```bash
# Terminal 1: Frontend dev server
cd web && npm run dev

# Terminal 2: Backend server
make dev-backend
```

The frontend development server proxies API requests to the backend running on port 8080. Configure the proxy in `vite.config.ts` if you need to change ports.

### Building for production

Build the frontend for production:

```bash
cd web
npm run build
```

This creates an optimized production build in the `dist/` directory. The build is then embedded into the Go binary using the build script in `scripts/build-frontend.js`.

The main README explains how to build the complete application with embedded frontend assets.

## Component patterns

Follow these patterns when adding new components.

### Page components

Page components represent full pages and are defined in `src/pages/`. Each page:

- Uses React Router for navigation
- Handles its own data fetching
- Manages local state with React hooks
- Uses the `AuthWrapper` component for authentication

Example page structure:

```typescript
import { useAuth } from '../hooks/useAuth';
import { useEffect, useState } from 'react';

export default function MyPage() {
  const { user } = useAuth();
  const [data, setData] = useState([]);

  useEffect(() => {
    // Fetch data
  }, []);

  return <div>{/* Page content */}</div>;
}
```

### Reusable components

Reusable components live in `src/components/` and can be used across multiple pages. Components should:

- Accept props with TypeScript interfaces
- Be self-contained and reusable
- Use shadcn/ui components for consistent styling
- Follow Tailwind CSS patterns

### API client functions

API client functions in `src/api/` handle communication with the backend:

- Use Axios for HTTP requests
- Include authentication tokens automatically
- Handle errors consistently
- Return typed responses

Example API function:

```typescript
import axios from 'axios';
import { config } from '../config';

export async function getProjects() {
  const response = await axios.get(`${config.apiUrl}/api/projects`);
  return response.data;
}
```

# Styling

The frontend uses Tailwind CSS for styling with shadcn/ui components.

### Tailwind CSS

Tailwind CSS provides utility classes for styling. Use Tailwind classes directly in your JSX:

```tsx
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
  <h1 className="text-2xl font-bold">Title</h1>
</div>
```

Customize Tailwind in `tailwind.config.ts`. The configuration extends the default theme with project-specific colors and spacing.

### shadcn/ui components

Use shadcn/ui components from `src/components/ui/` for consistent, accessible UI elements:

- Buttons, inputs, and forms
- Dialogs and modals
- Tables and cards
- Navigation components

Import components directly:

```tsx
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
```

## Integration with backend

The frontend integrates with the Go backend through RESTful API endpoints.

### API configuration

API configuration is set in `src/config.ts`. The config determines:

- API base URL (development vs production)
- Authentication token handling
- Request/response interceptors

### Authentication

Authentication is handled through Scalekit. The `useAuth` hook provides:

- Current user information
- Authentication status
- Login/logout functions
- Session management

All API requests automatically include authentication tokens through Axios interceptors.

### Data flow

Data flows from backend to frontend:

1. User interacts with UI component
2. Component calls API function from `src/api/`
3. API function makes HTTP request to backend
4. Backend processes request and returns JSON
5. Component updates state with response data
6. UI re-renders with new data

## Building and deployment

The frontend is built and embedded into the Go binary for deployment.

### Development build

For development, use the Vite dev server which provides:

- Fast hot module replacement
- Source maps for debugging
- TypeScript type checking
- ESLint integration

### Production build

For production, build the frontend:

```bash
cd web
npm run build
```

The build process:

1. Compiles TypeScript to JavaScript
2. Bundles and minifies code
3. Optimizes assets
4. Generates production-ready files in `dist/`

The `scripts/build-frontend.js` script then embeds these files into the Go binary.

### Standalone deployment

While the frontend is typically embedded in the Go binary, you can deploy it separately:

- **Vercel**: Connect the `web/` directory and set build command to `npm run build`
- **Netlify**: Set build command to `npm run build` and publish directory to `dist`
- **GitHub Pages**: Build and push the `dist/` directory to the `gh-pages` branch

For standalone deployment, update `src/config.ts` to point to your backend API URL.

## Troubleshooting

Common frontend development issues and solutions.

<details>
<summary><strong>TypeScript errors</strong></summary>

**Problem**: TypeScript compilation errors.

**Solutions**:

- Check `tsconfig.json` for correct configuration
- Ensure all imports have correct types
- Run `npm run build` to see all TypeScript errors
- Verify type definitions are installed: `npm install --save-dev @types/react`

</details>

<details>
<summary><strong>Styling not applied</strong></summary>

**Problem**: Tailwind CSS classes not working.

**Solutions**:

- Verify Tailwind is configured in `tailwind.config.ts`
- Check that `index.css` imports Tailwind directives
- Restart the dev server after changing Tailwind config
- Clear browser cache

</details>

<details>
<summary><strong>API requests failing</strong></summary>

**Problem**: Frontend cannot reach backend API.

**Solutions**:

- Verify backend server is running on port 8080
- Check `src/config.ts` has correct API URL
- Ensure CORS is configured in backend
- Check browser console for network errors
- Verify authentication tokens are being sent

</details>

<details>
<summary><strong>Hot reload not working</strong></summary>

**Problem**: Changes not reflected in browser.

**Solutions**:

- Restart the dev server
- Clear browser cache
- Check for syntax errors in console
- Verify file is being watched (check Vite output)

</details>

For more help, see the troubleshooting section in the [main README](../README.md) or [join the Scalekit community on Slack](https://join.slack.com/t/scalekit-community/shared_invite/zt-3gsxwr4hc-0tvhwT2b_qgVSIZQBQCWRw) to ask questions and get support.
