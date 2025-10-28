import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CreateTaskRequest, tasksApi, projectsApi, Project } from "@/api/projects";
import { config } from "@/config";
import { useToast } from "@/hooks/use-toast";

interface CreateTaskModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTaskCreated: () => void;
  initialData?: Partial<CreateTaskRequest>;
  taskId?: string;
  defaultProjectId?: string;
}

export function CreateTaskModal({
  open,
  onOpenChange,
  onTaskCreated,
  initialData,
  taskId,
  defaultProjectId,
}: CreateTaskModalProps) {
  const [formData, setFormData] = useState<CreateTaskRequest>({
    name: initialData?.name || "",
    description: initialData?.description || "",
    priority: initialData?.priority || "P3",
    status: initialData?.status || "Backlog",
    project_id: initialData?.project_id || defaultProjectId || "none",
    assignee_id: initialData?.assignee_id || "",
  });
  const [projects, setProjects] = useState<Project[]>([]);
  const [members, setMembers] = useState<Array<{ id: string; email: string; user_profile?: { first_name?: string; last_name?: string; name?: string } }>>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [loadingMembers, setLoadingMembers] = useState(false);
  const { toast } = useToast();

  const isEditing = !!taskId;

  // Load projects when modal opens
  useEffect(() => {
    if (open) {
      loadProjects();
      loadMembers();
    }
  }, [open]);

  const loadProjects = async () => {
    setLoadingProjects(true);
    try {
      const response = await projectsApi.getProjects(1, 100);
      setProjects(response.projects);
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load projects",
        variant: "destructive",
      });
    } finally {
      setLoadingProjects(false);
    }
  };

  const loadMembers = async () => {
    setLoadingMembers(true);
    try {
      const response = await fetch(`${config.backendUrl}/api/workspace/members`, {
        method: "GET",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
      });
      if (response.ok) {
        const data = await response.json();
        setMembers(data.users || []);
      } else {
        toast({
          title: "Error",
          description: "Failed to load members",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to load members",
        variant: "destructive",
      });
    } finally {
      setLoadingMembers(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const submitData = {
        ...formData,
        project_id: formData.project_id && formData.project_id !== "none" ? formData.project_id : undefined,
        assignee_id: formData.assignee_id || undefined,
      };

      if (isEditing) {
        await tasksApi.updateTask(taskId, submitData);
        toast({
          title: "Success",
          description: "Task updated successfully",
        });
      } else {
        await tasksApi.createTask(submitData);
        toast({
          title: "Success",
          description: "Task created successfully",
        });
      }
      onTaskCreated();
      onOpenChange(false);
      setFormData({
        name: "",
        description: "",
        priority: "P3",
        status: "Backlog",
        project_id: "none",
        assignee_id: "",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "An error occurred",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>
            {isEditing ? "Edit Task" : "Create New Task"}
          </DialogTitle>
          <DialogDescription>
            {isEditing
              ? "Update the task details below."
              : "Fill in the details to create a new task."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Task Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="Enter task name"
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description || ""}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                placeholder="Enter task description"
                rows={3}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="project_id">Project</Label>
              <Select
                value={formData.project_id || ""}
                onValueChange={(value) =>
                  setFormData({ ...formData, project_id: value || undefined })
                }
                disabled={loadingProjects}
              >
                <SelectTrigger>
                  <SelectValue placeholder={loadingProjects ? "Loading projects..." : "Select project (optional)"} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No Project</SelectItem>
                  {projects.map((project) => (
                    <SelectItem key={project.id} value={project.id}>
                      {project.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="priority">Priority *</Label>
                <Select
                  value={formData.priority}
                  onValueChange={(value: "P1" | "P2" | "P3") =>
                    setFormData({ ...formData, priority: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select priority" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="P1">P1 - High</SelectItem>
                    <SelectItem value="P2">P2 - Medium</SelectItem>
                    <SelectItem value="P3">P3 - Low</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="status">Status *</Label>
                <Select
                  value={formData.status}
                  onValueChange={(value: "Backlog" | "Todo" | "InProgress" | "Done") =>
                    setFormData({ ...formData, status: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="Backlog">Backlog</SelectItem>
                    <SelectItem value="Todo">To Do</SelectItem>
                    <SelectItem value="InProgress">In Progress</SelectItem>
                    <SelectItem value="Done">Done</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="assignee_id">Assignee</Label>
              <Select
                value={formData.assignee_id || "unassigned"}
                onValueChange={(value) =>
                  setFormData({ ...formData, assignee_id: value === "unassigned" ? "" : value })
                }
                disabled={loadingMembers}
              >
                <SelectTrigger>
                  <SelectValue placeholder={loadingMembers ? "Loading members..." : "Select assignee (optional)"} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unassigned">Unassigned</SelectItem>
                  {members.map((m) => {
                    const displayName = m.user_profile?.name || [m.user_profile?.first_name, m.user_profile?.last_name].filter(Boolean).join(" ") || m.email || m.id;
                    return (
                      <SelectItem key={m.id} value={m.email}>
                        {displayName}
                      </SelectItem>
                    );
                  })}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading
                ? isEditing
                  ? "Updating..."
                  : "Creating..."
                : isEditing
                ? "Update Task"
                : "Create Task"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
