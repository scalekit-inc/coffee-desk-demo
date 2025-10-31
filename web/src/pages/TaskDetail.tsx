import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { CreateTaskModal } from "@/components/CreateTaskModal";
import { StatusBadge } from "@/components/StatusBadge";
import { PriorityBadge } from "@/components/PriorityBadge";
import { AppSidebar } from "@/components/AppSidebar";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { tasksApi, Task } from "@/api/projects";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/hooks/useAuth";
import { hasPermission } from "@/components/RBACGuard";
import { ArrowLeft, MoreHorizontal, Edit, Trash2 } from "lucide-react";

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const { toast } = useToast();
  const { user } = useAuth();

  // RBAC: Check if user has permission to edit/delete tasks
  const canManageTasks = hasPermission(user?.permissions, "workspace:admin");

  const loadTask = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const response = await tasksApi.getTask(id);
      setTask(response.task);
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load task",
        variant: "destructive",
      });
      navigate("/dashboard/tasks");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTask();
  }, [id]);

  const handleDelete = async () => {
    if (!task || !confirm("Are you sure you want to delete this task?")) return;

    try {
      await tasksApi.deleteTask(task.id);
      toast({
        title: "Success",
        description: "Task deleted successfully",
      });
      navigate("/dashboard/tasks");
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete task",
        variant: "destructive",
      });
    }
  };

  const handleTaskUpdated = () => {
    loadTask();
    setIsEditModalOpen(false);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-muted-foreground">Loading task...</div>
      </div>
    );
  }

  if (!task) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-muted-foreground">Task not found</div>
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
            src="/uploads/coffee-desk-name-icon.png" 
            alt="Coffeedesk Logo" 
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
            {/* Back button and task header */}
            <div className="flex items-center gap-4 mb-6">
              <Button
                variant="ghost"
                onClick={() => navigate("/dashboard/tasks")}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Tasks
              </Button>
            </div>

            {/* Task details */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2">
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-2xl">{task.name}</CardTitle>
                        <CardDescription>
                          {task.description || "No description provided"}
                        </CardDescription>
                      </div>
                      {canManageTasks && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => setIsEditModalOpen(true)}>
                              <Edit className="h-4 w-4 mr-2" />
                              Edit Task
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={handleDelete}
                              className="text-destructive"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              Delete Task
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Priority</label>
                        <div className="mt-1">
                          <PriorityBadge priority={task.priority} />
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Status</label>
                        <div className="mt-1">
                          <StatusBadge status={task.status} />
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Project</label>
                        <div className="mt-1">
                          {task.project ? (
                            <Badge variant="outline">{task.project.name}</Badge>
                          ) : (
                            <span className="text-muted-foreground">No Project</span>
                          )}
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Assignee</label>
                        <div className="mt-1">
                          {task.assignee ? (
                            <Badge variant="outline">{task.assignee.email}</Badge>
                          ) : (
                            <span className="text-muted-foreground">Unassigned</span>
                          )}
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Created</label>
                        <div className="mt-1 text-sm">
                          {new Date(task.created_at).toLocaleDateString()}
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium text-muted-foreground">Last Updated</label>
                        <div className="mt-1 text-sm">
                          {new Date(task.updated_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              <div>
                <Card>
                  <CardHeader>
                    <CardTitle>Task Information</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div>
                        <div className="text-sm font-medium text-muted-foreground">Task ID</div>
                        <div className="text-sm font-mono">{task.id}</div>
                      </div>
                      <div>
                        <div className="text-sm font-medium text-muted-foreground">Organization ID</div>
                        <div className="text-sm font-mono">{task.organization_id}</div>
                      </div>
                      {task.project && (
                        <div>
                          <div className="text-sm font-medium text-muted-foreground">Project ID</div>
                          <div className="text-sm font-mono">{task.project_id}</div>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Edit Task Modal */}
      <CreateTaskModal
        open={isEditModalOpen}
        onOpenChange={setIsEditModalOpen}
        onTaskCreated={handleTaskUpdated}
        initialData={task}
        taskId={task.id}
      />
    </div>
  );
}
