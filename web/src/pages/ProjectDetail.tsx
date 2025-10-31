import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { CreateProjectModal } from "@/components/CreateProjectModal";
import { CreateTaskModal } from "@/components/CreateTaskModal";
import { StatusBadge } from "@/components/StatusBadge";
import { PriorityBadge } from "@/components/PriorityBadge";
import { AppSidebar } from "@/components/AppSidebar";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { projectsApi, tasksApi, Project, Task } from "@/api/projects";
import { useToast } from "@/hooks/use-toast";
import { ArrowLeft, Plus, MoreHorizontal, Edit, Trash2, Eye } from "lucide-react";

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [project, setProject] = useState<Project | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isCreateTaskModalOpen, setIsCreateTaskModalOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const { toast } = useToast();

  const loadProject = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const response = await projectsApi.getProject(id);
      setProject(response.project);
      setTasks(response.project.tasks || []);
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load project",
        variant: "destructive",
      });
      navigate("/dashboard/projects");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProject();
  }, [id]);

  const handleDeleteProject = async () => {
    if (!project || !confirm("Are you sure you want to delete this project?")) return;

    try {
      await projectsApi.deleteProject(project.id);
      toast({
        title: "Success",
        description: "Project deleted successfully",
      });
      navigate("/dashboard/projects");
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete project",
        variant: "destructive",
      });
    }
  };

  const handleDeleteTask = async (taskId: string) => {
    if (!confirm("Are you sure you want to delete this task?")) return;

    try {
      await tasksApi.deleteTask(taskId);
      toast({
        title: "Success",
        description: "Task deleted successfully",
      });
      loadProject(); // Reload to get updated task list
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete task",
        variant: "destructive",
      });
    }
  };

  const handleEditTask = (task: Task) => {
    setEditingTask(task);
    setIsCreateTaskModalOpen(true);
  };

  const handleProjectUpdated = () => {
    loadProject();
    setIsEditModalOpen(false);
  };

  const handleTaskCreated = () => {
    loadProject();
    setIsCreateTaskModalOpen(false);
    setEditingTask(null);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-muted-foreground">Loading project...</div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-muted-foreground">Project not found</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen w-full">
      {/* Top section with logo and organization dropdown */}
      <div className="flex items-center justify-between p-4 border-b bg-background z-10">
        <div className="flex items-center gap-4">
          {/* App logo in top-left */}
          <img 
            src="/uploads/fe8916e1-c333-4b24-9051-655a97f99240.png" 
            alt="DevRamp Logo" 
            className="h-8 w-auto"
          />
          
          {/* Organization switcher immediately next to logo */}
          <WorkspaceDropdown />
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
          <div className="p-6">
            {/* Back button and project header */}
            <div className="flex items-center gap-4 mb-6">
              <Button
                variant="ghost"
                onClick={() => navigate("/dashboard/projects")}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Projects
              </Button>
            </div>

            {/* Project details */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
              <div className="lg:col-span-2">
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-2xl">{project.name}</CardTitle>
                        <CardDescription>
                          {project.description || "No description provided"}
                        </CardDescription>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => setIsEditModalOpen(true)}>
                            <Edit className="h-4 w-4 mr-2" />
                            Edit Project
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={handleDeleteProject}
                            className="text-destructive"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete Project
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Priority</label>
                        <div className="mt-1">
                          <PriorityBadge priority={project.priority} />
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Status</label>
                        <div className="mt-1">
                          <StatusBadge status={project.status} />
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Owner</label>
                        <div className="mt-1">
                          {project.owner_id ? (
                            <Badge variant="outline">{project.owner_id}</Badge>
                          ) : (
                            <span className="text-muted-foreground">Unassigned</span>
                          )}
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Created</label>
                        <div className="mt-1 text-sm">
                          {new Date(project.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              <div>
                <Card>
                  <CardHeader>
                    <CardTitle>Quick Stats</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div>
                        <div className="text-2xl font-bold">{tasks.length}</div>
                        <div className="text-sm text-muted-foreground">Total Tasks</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold">
                          {tasks.filter(task => task.status === 'Done').length}
                        </div>
                        <div className="text-sm text-muted-foreground">Completed</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold">
                          {tasks.filter(task => task.status === 'InProgress').length}
                        </div>
                        <div className="text-sm text-muted-foreground">In Progress</div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>

            {/* Tasks section */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Tasks</CardTitle>
                    <CardDescription>
                      {tasks.length} tasks in this project
                    </CardDescription>
                  </div>
                  <Button onClick={() => setIsCreateTaskModalOpen(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    Add Task
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {tasks.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-8">
                    <div className="text-muted-foreground mb-4">No tasks yet</div>
                    <Button onClick={() => setIsCreateTaskModalOpen(true)}>
                      <Plus className="h-4 w-4 mr-2" />
                      Create your first task
                    </Button>
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Priority</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Assignee</TableHead>
                        <TableHead>Created</TableHead>
                        <TableHead className="w-[50px]"></TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {tasks.map((task) => (
                        <TableRow key={task.id}>
                          <TableCell>
                            <div>
                              <div className="font-medium">{task.name}</div>
                              {task.description && (
                                <div className="text-sm text-muted-foreground truncate max-w-xs">
                                  {task.description}
                                </div>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            <PriorityBadge priority={task.priority} />
                          </TableCell>
                          <TableCell>
                            <StatusBadge status={task.status} />
                          </TableCell>
                          <TableCell>
                            {task.assignee ? (
                              <Badge variant="outline">{task.assignee.email}</Badge>
                            ) : (
                              <span className="text-muted-foreground">Unassigned</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {new Date(task.created_at).toLocaleDateString()}
                          </TableCell>
                          <TableCell>
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => handleEditTask(task)}>
                                  <Edit className="h-4 w-4 mr-2" />
                                  Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                  onClick={() => handleDeleteTask(task.id)}
                                  className="text-destructive"
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      {/* Edit Project Modal */}
      <CreateProjectModal
        open={isEditModalOpen}
        onOpenChange={setIsEditModalOpen}
        onProjectCreated={handleProjectUpdated}
        initialData={project}
        projectId={project.id}
      />

      {/* Create/Edit Task Modal */}
      <CreateTaskModal
        open={isCreateTaskModalOpen}
        onOpenChange={setIsCreateTaskModalOpen}
        onTaskCreated={handleTaskCreated}
        initialData={editingTask}
        taskId={editingTask?.id}
        defaultProjectId={project.id}
      />
    </div>
  );
}
