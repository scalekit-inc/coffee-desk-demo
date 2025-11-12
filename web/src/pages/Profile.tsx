import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/AppSidebar";
import { DashboardHeader } from "@/components/DashboardHeader";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { MoreHorizontal, Search, User, LogOut } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useNavigate, useLocation } from "react-router-dom";
import { config } from "@/config";
import { EditableUserProfile } from "@/components/EditableUserProfile";

const Profile = () => {
  const { user, workspaces } = useAuth();
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

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

  const handleUserUpdate = () => {
    // The useAuth hook will automatically update the user data through refetchSession
  };

  const handleAddPasskey = () => {
    // Open backend endpoint in new tab, which will redirect to Scalekit passkeys page
    window.open(`${config.backendUrl}/api/scalekit/passkeys`, '_blank');
  };

  const displayName = user?.name || user?.first_name || "User";
  const initials = user?.first_name && user?.last_name 
    ? `${user.first_name[0]}${user.last_name[0]}` 
    : user?.name 
    ? user.name.split(' ').map(n => n[0]).join('').slice(0, 2)
    : "U";
  // Get first letter of email for profile icon
  const emailInitial = user?.email ? user.email[0].toUpperCase() : "U";

  if (!user) {
    return (
      <div className="flex flex-col min-h-screen w-full">
        <div className="flex items-center justify-between p-4 border-b">
          <h1 className="text-xl font-semibold">My Profile</h1>
        </div>
        <div className="flex-1 flex items-center justify-center">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen w-full">
      {/* Header */}
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

      {/* Main content area with sidebar and profile content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 flex-shrink-0 border-r bg-background">
          <AppSidebar />
        </div>

        {/* Main content area */}
        <div className="flex-1 overflow-auto">
          <div className="p-6">
            <div className="flex items-center gap-3 mb-6">
              <h1 className="text-2xl font-semibold">My Profile</h1>
              <span className="inline-flex items-center text-sm font-semibold text-primary bg-primary/10 border border-primary/20 px-3 py-1 rounded-full shadow-sm">
                Powered by Scalekit
              </span>
            </div>
            
            {/* User Profile Section */}
            <EditableUserProfile 
              user={user} 
              onUserUpdate={handleUserUpdate}
            />

            {/* Organizations Section */}
            <Card>
              <CardHeader>
                <CardTitle>Organizations</CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Organization</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="w-[50px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {workspaces.map((workspace) => (
                      <TableRow key={workspace.id}>
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 bg-primary rounded-md flex items-center justify-center">
                              <span className="text-primary-foreground font-semibold text-xs">
                                {workspace.display_name[0]?.toUpperCase()}
                              </span>
                            </div>
                            <div>
                              <div className="font-medium">{workspace.display_name}</div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          {workspace.is_current ? (
                            <Badge variant="secondary">Current</Badge>
                          ) : (
                            <Badge variant="outline">Member</Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>

            {/* Passkeys Section */}
            <Card className="mb-6">
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold mb-1">Passkeys</h3>
                    <p className="text-sm text-muted-foreground">
                      Securely sign-in with on-device biometric authentication.
                    </p>
                  </div>
                  <Button variant="outline" onClick={handleAddPasskey}>
                    Manage Passkeys
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile;
