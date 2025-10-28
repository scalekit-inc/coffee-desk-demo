import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LayoutDashboard, Users, TrendingUp, Activity, Search, User, LogOut, FolderOpen, CheckSquare } from "lucide-react";
import { config } from "@/config";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/AppSidebar";
import { DashboardHeader } from "@/components/DashboardHeader";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { projectsApi, tasksApi } from "@/api/projects";

const Dashboard = () => {
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [stats, setStats] = useState({
    totalProjects: 0,
    totalTasks: 0,
    completedTasks: 0,
    inProgressTasks: 0,
    recentProjects: [],
    recentTasks: []
  });
  const [loadingStats, setLoadingStats] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, loading } = useAuth();

  const handleLogout = () => {
    const redirectUri = window.location.href;
    
    // Use direct navigation instead of fetch to avoid CORS issues
    window.location.href = `${config.backendUrl}/api/logout?redirect_uri=${encodeURIComponent(redirectUri)}`;
    
    setIsProfileOpen(false);
  };

  const handleProfile = () => {
    navigate("/dashboard/profile");
    setIsProfileOpen(false);
  };

  const displayName = user?.name || user?.first_name || "User";
  const initials = user?.first_name && user?.last_name 
    ? `${user.first_name[0]}${user.last_name[0]}` 
    : user?.name 
    ? user.name.split(' ').map(n => n[0]).join('').slice(0, 2)
    : "U";

  // Load dashboard stats
  useEffect(() => {
    const loadStats = async () => {
      if (!user) return;
      
      try {
        setLoadingStats(true);
        const [projectsResponse, tasksResponse] = await Promise.all([
          projectsApi.getProjects(1, 10),
          tasksApi.getTasks(1, 10)
        ]);

        const projects = projectsResponse.projects;
        const tasks = tasksResponse.tasks;

        setStats({
          totalProjects: projectsResponse.pagination.total,
          totalTasks: tasksResponse.pagination.total,
          completedTasks: tasks.filter(task => task.status === 'Done').length,
          inProgressTasks: tasks.filter(task => task.status === 'InProgress').length,
          recentProjects: projects.slice(0, 5),
          recentTasks: tasks.slice(0, 5)
        });
      } catch (error) {
        console.error('Failed to load dashboard stats:', error);
      } finally {
        setLoadingStats(false);
      }
    };

    loadStats();
  }, [user]);

  // Get first letter of email for profile icon
  const emailInitial = user?.email ? user.email[0].toUpperCase() : "U";

  return (
    <div className="flex flex-col min-h-screen w-full">
      {/* Top section with logo and workspace dropdown */}
      <div className="flex items-center justify-between p-4 border-b bg-background z-10">
        <div className="flex items-center gap-4">
          {/* App logo in top-left */}
          <img 
            src="/uploads/fe8916e1-c333-4b24-9051-655a97f99240.png" 
            alt="DevRamp Logo" 
            className="h-8 w-auto"
          />
          
          {/* Workspace switcher immediately next to logo */}
          <WorkspaceDropdown />
        </div>
        
        {/* Right side of header with search, documentation, and profile */}
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search"
              className="pl-10 pr-4 py-2 border border-input rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
            />
          </div>

          <Button variant="ghost" size="sm">
            Documentation
          </Button>

          <Popover open={isProfileOpen} onOpenChange={setIsProfileOpen}>
            <PopoverTrigger asChild>
              <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-primary text-primary-foreground font-bold">
                    {emailInitial}
                  </AvatarFallback>
                </Avatar>
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-48 p-2 bg-background border shadow-lg" align="end">
              <div className="space-y-1">
                <Button
                  variant="ghost"
                  className="w-full justify-start gap-2 h-auto py-2"
                  onClick={handleProfile}
                >
                  <User className="h-4 w-4" />
                  My Profile
                </Button>
                <Button
                  variant="ghost"
                  className="w-full justify-start gap-2 h-auto py-2 text-destructive hover:text-destructive"
                  onClick={handleLogout}
                >
                  <LogOut className="h-4 w-4" />
                  Logout
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      {/* Main content area with sidebar and dashboard content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 flex-shrink-0 border-r bg-background">
          <AppSidebar />
        </div>

        {/* Main content area */}
        <div className="flex-1 overflow-auto">
          {/* Dashboard content */}
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Projects</CardTitle>
                  <FolderOpen className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {loadingStats ? "..." : stats.totalProjects}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Active projects
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Tasks</CardTitle>
                  <CheckSquare className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {loadingStats ? "..." : stats.totalTasks}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    All tasks
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Completed</CardTitle>
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {loadingStats ? "..." : stats.completedTasks}
                  </div>
                  <p className="text-xs text-green-600 flex items-center gap-1">
                    <TrendingUp className="h-3 w-3" />
                    {stats.totalTasks > 0 ? Math.round((stats.completedTasks / stats.totalTasks) * 100) : 0}% completion rate
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">In Progress</CardTitle>
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {loadingStats ? "..." : stats.inProgressTasks}
                  </div>
                  <p className="text-xs text-blue-600 flex items-center gap-1">
                    Currently active
                  </p>
                </CardContent>
              </Card>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>Recent Projects</CardTitle>
                  <CardDescription>Latest projects created</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {loadingStats ? (
                      <div className="text-muted-foreground">Loading...</div>
                    ) : stats.recentProjects.length === 0 ? (
                      <div className="text-muted-foreground">No projects yet</div>
                    ) : (
                      stats.recentProjects.map((project) => (
                        <div key={project.id} className="flex items-center space-x-4">
                          <div className="w-2 h-2 bg-primary rounded-full"></div>
                          <div className="flex-1">
                            <p className="text-sm font-medium">{project.name}</p>
                            <p className="text-xs text-muted-foreground">
                              Created {new Date(project.created_at).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Quick Actions</CardTitle>
                  <CardDescription>Frequently used operations</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-4">
                    <Button 
                      variant="outline" 
                      className="h-20 flex-col"
                      onClick={() => navigate("/dashboard/projects")}
                    >
                      <FolderOpen className="h-6 w-6 mb-2" />
                      Manage Projects
                    </Button>
                    <Button 
                      variant="outline" 
                      className="h-20 flex-col"
                      onClick={() => navigate("/dashboard/tasks")}
                    >
                      <CheckSquare className="h-6 w-6 mb-2" />
                      Manage Tasks
                    </Button>
                    <Button 
                      variant="outline" 
                      className="h-20 flex-col"
                      onClick={() => navigate("/dashboard/workspace")}
                    >
                      <Users className="h-6 w-6 mb-2" />
                      Workspace
                    </Button>
                    <Button 
                      variant="outline" 
                      className="h-20 flex-col"
                      onClick={() => navigate("/dashboard/profile")}
                    >
                      <User className="h-6 w-6 mb-2" />
                      My Profile
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
