import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/useAuth";
import { config } from "@/config";

interface EditableUserProfileProps {
  user: {
    id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    name?: string;
  };
  onUserUpdate: () => void;
}

export function EditableUserProfile({ user, onUserUpdate }: EditableUserProfileProps) {
  const [firstName, setFirstName] = useState(user.first_name || "");
  const [lastName, setLastName] = useState(user.last_name || "");
  const [isSaving, setIsSaving] = useState(false);
  const { refetchSession } = useAuth();

  // Update local state when user prop changes
  useEffect(() => {
    setFirstName(user.first_name || "");
    setLastName(user.last_name || "");
  }, [user]);

  // Check if any field has been modified
  const hasChanges = 
    firstName !== (user.first_name || "") ||
    lastName !== (user.last_name || "");

  const handleSave = async () => {
    if (!hasChanges) return;

    try {
      setIsSaving(true);
      
      const response = await fetch(`${config.backendUrl}/api/user/profile`, {
        method: "PUT",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          firstName: firstName.trim(),
          lastName: lastName.trim(),
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to update profile: ${response.status}`);
      }

      onUserUpdate(); 
    } catch (error) {
      console.error("Error updating user profile:", error);
      // TODO: Add toast notification
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    // Reset to original values
    setFirstName(user.first_name || "");
    setLastName(user.last_name || "");
  };

  return (
    <Card className="mb-8">
      <CardHeader>
        <CardTitle>User Profile</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="space-y-2">
            <label className="text-sm font-medium">First Name</label>
            <Input
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              placeholder="Enter first name"
              disabled={isSaving}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Last Name</label>
            <Input
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              placeholder="Enter last name"
              disabled={isSaving}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Email</label>
            <input 
              type="text" 
              value={user.email}
              className="w-full px-3 py-2 border rounded-md bg-muted"
              readOnly
            />
          </div>
        </div>

        <div className="flex gap-2">
          <Button 
            onClick={handleSave}
            disabled={isSaving || !hasChanges}
          >
            {isSaving ? 'Saving...' : 'Save'}
          </Button>
          <Button 
            variant="outline" 
            onClick={handleCancel}
            disabled={isSaving || !hasChanges}
          >
            Cancel
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
