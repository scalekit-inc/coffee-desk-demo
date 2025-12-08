import { BrowserRouter, Routes, Route } from "react-router-dom";
import Index from "./pages/Index";
import Dashboard from "./pages/Dashboard";
import Profile from "./pages/Profile";
import Workspace from "./pages/Workspace";
import Onboarding from "./pages/Onboarding";
import Projects from "./pages/Projects";
import ProjectDetail from "./pages/ProjectDetail";
import Tasks from "./pages/Tasks";
import TaskDetail from "./pages/TaskDetail";
import NotFound from "./pages/NotFound";
import { AuthWrapper } from "./components/AuthWrapper";

const App = () => (
  <BrowserRouter>
    <Routes>
      <Route path="/" element={<Index />} />
      <Route 
        path="/onboarding" 
        element={
          <AuthWrapper requireAuth={false}>
            <Onboarding />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard" 
        element={
          <AuthWrapper requireAuth={true}>
            <Dashboard />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/profile" 
        element={
          <AuthWrapper requireAuth={true}>
            <Profile />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/workspace" 
        element={
          <AuthWrapper requireAuth={true}>
            <Workspace />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/workspace/members" 
        element={
          <AuthWrapper requireAuth={true}>
            <Workspace />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/workspace/security" 
        element={
          <AuthWrapper requireAuth={true}>
            <Workspace />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/workspace/settings" 
        element={
          <AuthWrapper requireAuth={true}>
            <Workspace />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/projects" 
        element={
          <AuthWrapper requireAuth={true}>
            <Projects />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/projects/:id" 
        element={
          <AuthWrapper requireAuth={true}>
            <ProjectDetail />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/tasks" 
        element={
          <AuthWrapper requireAuth={true}>
            <Tasks />
          </AuthWrapper>
        } 
      />
      <Route 
        path="/dashboard/tasks/:id" 
        element={
          <AuthWrapper requireAuth={true}>
            <TaskDetail />
          </AuthWrapper>
        } 
      />
      <Route path="*" element={<NotFound />} />
    </Routes>
  </BrowserRouter>
);

export default App;
