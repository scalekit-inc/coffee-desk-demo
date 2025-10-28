import { useState } from "react";
import { useForm } from "react-hook-form";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { config } from "@/config";

interface CreateWorkspaceModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

interface CreateWorkspaceFormData {
  workspaceName: string;
}

interface CreateWorkspaceResponse {
  message: string;
  workspace: {
    id: string;
    display_name: string;
    external_id: string;
  };
  membership: {
    user_id: string;
    email: string;
  };
}

export function CreateWorkspaceModal({
  open,
  onOpenChange,
  onSuccess,
}: CreateWorkspaceModalProps) {
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<CreateWorkspaceFormData>({
    defaultValues: {
      workspaceName: "",
    },
  });

  const onSubmit = async (data: CreateWorkspaceFormData) => {
    setIsLoading(true);

    try {
      
      const response = await fetch(`${config.backendUrl}/api/workspace/create`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          workspaceName: data.workspaceName,
        }),
      });

      if (!response.ok) {
        throw new Error(`Workspace creation failed: ${response.status}`);
      }

      const result: CreateWorkspaceResponse = await response.json();
      
      // Reset form and close modal
      form.reset();
      onOpenChange(false);
      
      // Redirect to authorize with the new organization ID to switch to the new workspace
      window.location.href = `${config.backendUrl}/api/authorize?organization_id=${result.workspace.id}&prompt=select_account`;
      
    } catch (error) {
      console.error("Workspace creation error:", error);
      // You can add toast notification here for error handling
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create New Workspace</DialogTitle>
          <DialogDescription>
            Enter a name for your new workspace. You can always change this later.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="workspaceName"
              rules={{ required: "Workspace name is required" }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Workspace name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="e.g. Acme Corp"
                      {...field}
                      disabled={isLoading}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
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
                {isLoading ? "Creating..." : "Create Workspace"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
