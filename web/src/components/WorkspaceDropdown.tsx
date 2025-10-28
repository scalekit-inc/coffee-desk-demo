
import { useState } from "react";
import { ChevronDown, Plus } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/useAuth";
import { config } from "@/config";
import { CreateWorkspaceModal } from "./CreateWorkspaceModal";

export function WorkspaceDropdown() {
  const { workspaces, currentWorkspace, loading } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  const handleWorkspaceSwitch = (workspaceId: string) => {
    
    // Make API call to authorize with organization_id
    window.location.href = `${config.backendUrl}/api/authorize?organization_id=${workspaceId}&prompt=select_account`;
    
    setIsOpen(false);
  };

  const handleCreateWorkspace = () => {
    setIsOpen(false);
    setIsCreateModalOpen(true);
  };

  const handleWorkspaceCreated = () => {
    // The modal now handles the redirect directly, so we don't need to do anything here
  };

  if (loading) {
    return (
      <Button variant="ghost" className="flex items-center gap-2 px-3 py-2 h-auto" disabled>
        <div className="w-6 h-6 bg-muted rounded-md flex items-center justify-center">
          <span className="text-muted-foreground font-semibold text-xs">W</span>
        </div>
        <span className="font-medium text-muted-foreground">Loading...</span>
        <ChevronDown className="h-4 w-4 text-muted-foreground" />
      </Button>
    );
  }

  return (
    <>
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="flex items-center gap-2 px-3 py-2 h-auto">
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 bg-primary rounded-md flex items-center justify-center">
                <span className="text-primary-foreground font-semibold text-xs">
                  {currentWorkspace?.display_name?.[0]?.toUpperCase() || "W"}
                </span>
              </div>
              <span className="font-medium text-foreground">
                {currentWorkspace?.display_name || "Workspace"}
              </span>
            </div>
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-64 bg-background border shadow-lg" align="start">
          <DropdownMenuLabel>Switch Workspace</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {workspaces.map((workspace) => (
            <DropdownMenuItem
              key={workspace.id}
              onClick={() => handleWorkspaceSwitch(workspace.id)}
              className={workspace.is_current ? "bg-accent" : ""}
            >
              <div className="flex items-center gap-3 w-full">
                <div className="w-6 h-6 bg-primary rounded-md flex items-center justify-center">
                  <span className="text-primary-foreground font-semibold text-xs">
                    {workspace.display_name[0]?.toUpperCase()}
                  </span>
                </div>
                <div className="flex-1">
                  <div className="font-medium">{workspace.display_name}</div>
                  {workspace.is_current && (
                    <div className="text-xs text-muted-foreground">Current</div>
                  )}
                </div>
              </div>
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleCreateWorkspace}>
            <div className="flex items-center gap-3 w-full">
              <div className="w-6 h-6 border-2 border-dashed border-muted-foreground rounded-md flex items-center justify-center">
                <Plus className="h-3 w-3 text-muted-foreground" />
              </div>
              <span>Create new workspace</span>
            </div>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <CreateWorkspaceModal
        open={isCreateModalOpen}
        onOpenChange={setIsCreateModalOpen}
        onSuccess={handleWorkspaceCreated}
      />
    </>
  );
}
