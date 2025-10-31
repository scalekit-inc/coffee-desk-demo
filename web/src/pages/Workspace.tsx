import { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { AppSidebar } from "@/components/AppSidebar";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog";
import { InviteUserForm } from "@/components/InviteUserForm";
import { Plus, MoreHorizontal, Trash2, Search, User, LogOut, Mail, Info, Save, Globe, MoreVertical } from "lucide-react";
import { config } from '@/config';
import { toast } from "sonner";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { RBACGuard } from "@/components/RBACGuard";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface UserProfile {
  custom_attributes: any;
  email_verified: boolean;
  first_name: string;
  id: string;
  last_name: string;
  locale: string;
  metadata: any;
  name: string;
  phone_number: string;
}

interface Role {
  id: string;
  name: string;
}

interface Membership {
  membership_status: string;
  metadata: any;
  name: string;
  organization_id: string;
  primary_identity_provider: string;
  roles: Role[];
  expires_at?: string;
}

interface User {
  create_time: string;
  email: string;
  environment_id: string;
  external_id: string;
  id: string;
  memberships: Membership[];
  metadata: any;
  update_time: string;
  user_profile: UserProfile;
}

interface MembersResponse {
  next_page_token: string;
  prev_page_token: string;
  total_size: number;
  users: User[];
}

interface Domain {
  id: string;
  name: string;
  type: string;
  created_at?: string | {
    seconds: number;
    nanos: number;
  };
}

interface DomainsResponse {
  domains: Domain[];
}

export default function Workspace() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState<User | null>(null);
  const [inviteFormOpen, setInviteFormOpen] = useState(false);
  const [portalUrl, setPortalUrl] = useState<string | null>(null);
  const [portalLoading, setPortalLoading] = useState(false);
  const [portalError, setPortalError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [hasFetchedPortal, setHasFetchedPortal] = useState(false);
  const MAX_RETRIES = 3;
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [openDropdownId, setOpenDropdownId] = useState<string | null>(null);
  const [deletingUserId, setDeletingUserId] = useState<string | null>(null);
  const [resendingInviteUserId, setResendingInviteUserId] = useState<string | null>(null);
  const [workspaceName, setWorkspaceName] = useState<string>("");
  const [isSaving, setIsSaving] = useState(false);
  const [allowedDomains, setAllowedDomains] = useState<string>("");
  const [isSavingDomains, setIsSavingDomains] = useState(false);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [domainsLoading, setDomainsLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, currentWorkspace, workspaces, loading: authLoading } = useAuth();


  // Get current workspace from workspaces array if currentWorkspace is not available
  const getCurrentWorkspace = () => {
    if (currentWorkspace) return currentWorkspace;
    if (workspaces && workspaces.length > 0) {
      return workspaces.find(ws => ws.is_current) || workspaces[0];
    }
    return null;
  };

  const workspaceData = getCurrentWorkspace();

  // Initialize workspace name when workspace data changes
  useEffect(() => {
    if (workspaceData?.display_name) {
      setWorkspaceName(workspaceData.display_name);
    }
  }, [workspaceData]);

  // Determine the active tab based on URL
  const getActiveTab = () => {
    if (location.pathname === '/dashboard/workspace/members') {
      return 'members';
    }
    if (location.pathname === '/dashboard/workspace/security') {
      return 'security';
    }
    if (location.pathname === '/dashboard/workspace/settings') {
      return 'advanced';
    }
    return 'general'; // default tab
  };

  const handleTabChange = (value: string) => {
    switch (value) {
      case 'members':
        navigate('/dashboard/workspace/members');
        break;
      case 'security':
        navigate('/dashboard/workspace/security');
        break;
      case 'advanced':
        navigate('/dashboard/workspace/settings');
        break;
      default:
        navigate('/dashboard/workspace');
    }
  };

  const fetchMembers = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(`${config.backendUrl}/api/workspace/members`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      
      if (response.ok) {
        const data: MembersResponse = await response.json();
        setUsers(data.users || []);
      } else {
        setError(`Failed to fetch members: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to fetch members:", error);
      setError(error instanceof Error ? error.message : 'Failed to fetch members');
    } finally {
      setLoading(false);
    }
  };

  const fetchDomains = async () => {
    try {
      setDomainsLoading(true);
      
      const response = await fetch(`${config.backendUrl}/api/workspace/domains`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data: DomainsResponse = await response.json();
        setDomains(data.domains || []);
      } else {
        console.error(`Failed to fetch domains: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to fetch domains:", error);
    } finally {
      setDomainsLoading(false);
    }
  };

  const fetchPortalLink = async () => {
    try {
      setPortalLoading(true);
      setPortalError(null);
      
      const response = await fetch(`${config.backendUrl}/api/portal/link`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setPortalUrl(data.link);
        setRetryCount(0); // Reset retry count on success
        setHasFetchedPortal(true); 
      } else {
        setPortalError(`Failed to fetch portal link: ${response.status}`);
        setRetryCount(prev => prev + 1);
      }
    } catch (error) {
      setPortalError(error instanceof Error ? error.message : 'Failed to fetch portal link');
      setRetryCount(prev => prev + 1);
    } finally {
      setPortalLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
    fetchDomains();
  }, []);

  // Clean up dropdown states when users list changes
  useEffect(() => {
    // If the user that had an open dropdown is no longer in the list, close the dropdown
    if (openDropdownId && !users.find(user => user.id === openDropdownId)) {
      setOpenDropdownId(null);
    }
  }, [users, openDropdownId]);

  // Auto-fetch portal link when needed
  useEffect(() => {
    // Always fetch a fresh portal link when on Advanced tab
    if (location.pathname === '/dashboard/workspace/settings' && !portalLoading && !hasFetchedPortal && retryCount < MAX_RETRIES) {
      fetchPortalLink();
    } else if (location.pathname !== '/dashboard/workspace/settings') {
      // Reset flags when navigating away from Advanced tab
      setRetryCount(0);
      setHasFetchedPortal(false);
    }
  }, [location.pathname, portalLoading, hasFetchedPortal, retryCount]);

  // Auto-retry on error
  useEffect(() => {
    if (portalError && location.pathname === '/dashboard/workspace/settings' && !portalLoading && retryCount < MAX_RETRIES) {
      // Retry after 3 seconds
      const timer = setTimeout(() => {
        setPortalError(null);
        setHasFetchedPortal(false); // Reset flag to allow retry
        fetchPortalLink();
      }, 3000);
      
      return () => clearTimeout(timer);
    }
  }, [portalError, location.pathname, portalLoading, retryCount]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      setOpenDropdownId(null);
      setDeleteDialogOpen(false);
      setUserToDelete(null);
      setDeletingUserId(null);
      setResendingInviteUserId(null);
    };
  }, []);

  const handleDeleteUser = async (user: User) => {
    // Prevent multiple delete operations
    if (deletingUserId) {
      return;
    }

    try {
      
      // Set deleting state
      setDeletingUserId(user.id);
      
      // Close any open dropdowns first
      setOpenDropdownId(null);
      
      const response = await fetch(`${config.backendUrl}/api/workspace/members/${user.id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        toast.success('User deleted successfully');
        // Optimistically update the UI by removing the user from the list
        setUsers(prevUsers => prevUsers.filter(u => u.id !== user.id));
      } else {
        toast.error(`Failed to delete user: ${response.status}`);
        // Refresh the list to ensure consistency
        await fetchMembers();
      }
    } catch (error) {
      console.error("Failed to delete user:", error);
      toast.error(error instanceof Error ? error.message : 'Failed to delete user');
      // Refresh the list to ensure consistency
      await fetchMembers();
    } finally {
      setDeleteDialogOpen(false);
      setUserToDelete(null);
      setDeletingUserId(null);
    }
  };

  const handleDeleteClick = (user: User) => {
    setUserToDelete(user);
    setDeleteDialogOpen(true);
    setOpenDropdownId(null); // Close the dropdown
  };

  const handleResendInvite = async (user: User) => {
    // Prevent multiple resend operations
    if (resendingInviteUserId) {
      return;
    }

    try {
      
      // Set resending state
      setResendingInviteUserId(user.id);
      
      // Close any open dropdowns first
      setOpenDropdownId(null);
      
      const response = await fetch(`${config.backendUrl}/api/workspace/members/resend-invite`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: user.id,
          organization_id: workspaceData?.id
        }),
      });

      if (response.ok) {
        toast.success('Invite resent successfully');
      } else {
        toast.error(`Failed to resend invite: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to resend invite:", error);
      toast.error(error instanceof Error ? error.message : 'Failed to resend invite');
    } finally {
      setResendingInviteUserId(null);
    }
  };

  const handleResendInviteClick = (user: User) => {
    handleResendInvite(user);
    setOpenDropdownId(null); // Close the dropdown
  };

  const getInitials = (user: User) => {
    const name = user.user_profile.name || `${user.user_profile.first_name} ${user.user_profile.last_name}`.trim();
    if (name) {
      return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase();
    }
    return user.email.slice(0, 2).toUpperCase();
  };

  const getDisplayName = (user: User) => {
    const profile = user.user_profile;
    if (profile.name) return profile.name;
    if (profile.first_name || profile.last_name) {
      return `${profile.first_name} ${profile.last_name}`.trim();
    }
    return user.email.split('@')[0];
  };

  const getRole = (user: User) => {
    const membership = user.memberships[0];
    if (membership?.roles?.length > 0) {
      return membership.roles.map(role => role.name).join(', ');
    }
    return 'Member';
  };

  const getStatus = (user: User) => {
    const membership = user.memberships[0];
    if (membership?.membership_status) {
      return membership.membership_status;
    }
    return 'Unknown';
  };

  const formatExpirationDate = (expiresAt: string) => {
    try {
      const date = new Date(expiresAt);
      const day = String(date.getDate()).padStart(2, '0');
      const monthShort = date.toLocaleString('en-US', { month: 'short' });
      const year = date.getFullYear();
      let hours = date.getHours();
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const ampm = hours >= 12 ? 'PM' : 'AM';
      hours = hours % 12;
      if (hours === 0) hours = 12;
      const hourStr = String(hours).padStart(2, '0');
      return `${day}-${monthShort}-${year} ${hourStr}:${minutes}${ampm}`;
    } catch (error) {
      console.error('Error formatting expiration date:', error);
      return 'Invalid date';
    }
  };

  const getExpirationDate = (user: User) => {
    const membership = user.memberships[0];
    return membership?.expires_at;
  };

  const renderStatusBadge = (user: User) => {
    const status = getStatus(user);
    const expirationDate = getExpirationDate(user);
    
    if (status === 'PENDING_INVITE') {
      return (
        <span className="inline-flex items-center gap-1">
          <span className={getStatusBadge(status)}>{status}</span>
          <Tooltip>
            <TooltipTrigger asChild>
              <span>
                <Info className="h-3.5 w-3.5 text-muted-foreground" />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>
                {expirationDate
                  ? `This invitation will expire on ${formatExpirationDate(expirationDate)}`
                  : 'This invitation has no expiration date available'}
              </p>
            </TooltipContent>
          </Tooltip>
        </span>
      );
    }
    
    return (
      <span className={getStatusBadge(status)}>
        {status}
      </span>
    );
  };

  const getStatusBadge = (status: string) => {
    const baseClasses = "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium";
    
    // Ensure status is a string and handle null/undefined cases
    const statusString = status || 'Unknown';
    
    switch (statusString.toUpperCase()) {
      case 'ACTIVE':
        return `${baseClasses} bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300`;
      case 'PENDING':
      case 'PENDING_INVITE':
        return `${baseClasses} bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300`;
      case 'INACTIVE':
        return `${baseClasses} bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300`;
      case 'SUSPENDED':
        return `${baseClasses} bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300`;
      default:
        return `${baseClasses} bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300`;
    }
  };

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

  const handleSaveWorkspace = async () => {
    if (!workspaceData?.id || !workspaceName.trim()) {
      toast.error("Organization name cannot be empty");
      return;
    }

    try {
      setIsSaving(true);
      
      const response = await fetch(`${config.backendUrl}/api/workspace`, {
        method: 'PUT',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          display_name: workspaceName.trim()
        }),
      });

      if (response.ok) {
        toast.success('Organization name updated successfully');
        // You might want to refresh the workspace data here
        // or update the local state if you have a way to refresh workspaces
      } else {
        toast.error(`Failed to update organization: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to update organization:", error);
      toast.error(error instanceof Error ? error.message : 'Failed to update organization');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancelEdit = () => {
    // Reset to original workspace name
    if (workspaceData?.display_name) {
      setWorkspaceName(workspaceData.display_name);
    }
  };

  const handleSaveDomains = async () => {
    if (!allowedDomains.trim()) {
      toast.error("Please enter a domain");
      return;
    }

    try {
      setIsSavingDomains(true);
      
      // Clean the domain (remove @ symbol if present)
      const cleanDomain = allowedDomains.trim().replace(/^@/, '');

      if (cleanDomain.length === 0) {
        toast.error("Please enter a valid domain");
        return;
      }

      const response = await fetch(`${config.backendUrl}/api/workspace/domains`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          domain: cleanDomain
        }),
      });

      if (response.ok) {
        toast.success('Allowed email domain added successfully');
        // Refresh the domains list to show the newly created domain
        await fetchDomains();
        // Clear the input field
        setAllowedDomains('');
      } else {
        const errorData = await response.text();
        console.error("Domain save error:", errorData);
        toast.error(`Failed to save domain: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to save domain:", error);
      toast.error(error instanceof Error ? error.message : 'Failed to save domain');
    } finally {
      setIsSavingDomains(false);
    }
  };

  const handleDeleteDomain = async (domainId: string) => {
    try {
      console.log('Attempting to delete domain:', domainId);
      const response = await fetch(`${config.backendUrl}/api/workspace/domains/${domainId}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      console.log('Delete response status:', response.status);
      
      if (response.ok) {
        const responseData = await response.json();
        console.log('Delete response data:', responseData);
        toast.success('Domain removed successfully');
        // Refresh the domains list
        await fetchDomains();
      } else {
        const errorData = await response.text();
        console.error('Delete failed:', response.status, errorData);
        toast.error(`Failed to remove domain: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to remove domain:", error);
      toast.error(error instanceof Error ? error.message : 'Failed to remove domain');
    }
  };

  const formatDomainTime = (createdAt?: string | { seconds: number; nanos: number }) => {
    if (!createdAt) {
      return "Added recently";
    }
    
    try {
      let date: Date;
      
      // Handle protobuf timestamp format
      if (typeof createdAt === 'object' && createdAt.seconds) {
        // Convert protobuf timestamp (seconds + nanos) to JavaScript Date
        date = new Date(createdAt.seconds * 1000 + Math.floor(createdAt.nanos / 1000000));
      } else if (typeof createdAt === 'string') {
        // Handle string timestamp
        if (createdAt === 'null' || createdAt === 'undefined') {
          return "Added recently";
        }
        date = new Date(createdAt);
      } else {
        return "Added recently";
      }
      
      // Check if the date is valid
      if (isNaN(date.getTime())) {
        console.warn('Invalid date format:', createdAt);
        return "Added recently";
      }
      
      const now = new Date();
      const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
      
      // Handle negative time differences (future dates)
      if (diffInMinutes < 0) {
        return "Added just now";
      }
      
      if (diffInMinutes < 1) {
        return "Added just now";
      } else if (diffInMinutes < 60) {
        return `Added ${diffInMinutes} minute${diffInMinutes === 1 ? '' : 's'} ago`;
      } else if (diffInMinutes < 1440) { // 24 hours
        const hours = Math.floor(diffInMinutes / 60);
        return `Added ${hours} hour${hours === 1 ? '' : 's'} ago`;
      } else {
        const days = Math.floor(diffInMinutes / 1440);
        return `Added ${days} day${days === 1 ? '' : 's'} ago`;
      }
    } catch (error) {
      console.error('Error formatting domain time:', error, 'Input:', createdAt);
      return "Added recently";
    }
  };

  const displayName = user?.name || user?.first_name || "User";
  const initials = user?.first_name && user?.last_name 
    ? `${user.first_name[0]}${user.last_name[0]}` 
    : user?.name 
    ? user.name.split(' ').map(n => n[0]).join('').slice(0, 2)
    : "U";

  // Get first letter of email for profile icon
  const emailInitial = user?.email ? user.email[0].toUpperCase() : "U";

  return (
    <RBACGuard requirePermission="workspace:admin">
      <TooltipProvider delayDuration={0}>
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

      {/* Main content area with sidebar and organization content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="w-64 flex-shrink-0 border-r bg-background">
          <AppSidebar />
        </div>

        {/* Main content area */}
        <div className="flex-1 overflow-auto">
          <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="text-2xl font-semibold">Organization Settings</h1>
            </div>

            <Tabs defaultValue={getActiveTab()} className="space-y-6" value={getActiveTab()} onValueChange={handleTabChange}>
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="general">General</TabsTrigger>
                <TabsTrigger value="members">Members</TabsTrigger>
                <TabsTrigger value="security">Security</TabsTrigger>
                <TabsTrigger value="advanced">Advanced</TabsTrigger>
              </TabsList>

              <TabsContent value="general" className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Organization Settings</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <h3 className="text-sm font-medium">Organization Logo</h3>
                          <p className="text-xs text-muted-foreground">Recommended image size: 200x200px</p>
                        </div>
                        <div className="w-12 h-12 bg-primary rounded-lg flex items-center justify-center">
                          <span className="text-primary-foreground font-bold">
                            {workspaceData?.display_name?.[0]?.toUpperCase() || 'W'}
                          </span>
                        </div>
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium">Organization name</label>
                        <input 
                          type="text" 
                          value={workspaceName}
                          onChange={(e) => setWorkspaceName(e.target.value)}
                          className="w-full px-3 py-2 border rounded-md"
                          placeholder="Enter organization name"
                        />
                      </div>

                      <div className="space-y-2">
                        <label className="text-sm font-medium">Organization ID</label>
                        <input 
                          type="text" 
                          value={workspaceData?.id || ""}
                          className="w-full px-3 py-2 border rounded-md bg-muted"
                          readOnly
                        />
                      </div>



                      <div className="flex gap-2">
                        <Button 
                          onClick={handleSaveWorkspace}
                          disabled={isSaving || !workspaceName.trim() || workspaceName === workspaceData?.display_name}
                        >
                          {isSaving ? 'Saving...' : 'Save'}
                        </Button>
                        <Button 
                          variant="outline" 
                          onClick={handleCancelEdit}
                          disabled={isSaving}
                        >
                          Cancel
                        </Button>
                      </div>
                    </div>

                    {/* <div className="pt-6 border-t">
                      <div className="space-y-4">
                        <h3 className="text-sm font-medium text-destructive">Delete Workspace</h3>
                        <p className="text-sm text-muted-foreground">This is a permanent and irreversible action</p>
                        <Button variant="destructive">Delete Workspace</Button>
                      </div>
                    </div> */}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="members" className="space-y-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between">
                    <div>
                      <CardTitle>Members</CardTitle>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <input
                          type="text"
                          placeholder="Search"
                          className="pl-3 pr-4 py-2 border rounded-md text-sm w-64"
                        />
                      </div>
                      <Button onClick={() => setInviteFormOpen(true)}>
                        <Plus className="h-4 w-4 mr-2" />
                        Invite
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {loading ? (
                      <div className="flex items-center justify-center py-8">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                      </div>
                    ) : error ? (
                      <div className="text-center py-8">
                        <p className="text-destructive">{error}</p>
                        <Button onClick={fetchMembers} className="mt-2">Retry</Button>
                      </div>
                    ) : (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Name and email</TableHead>
                            <TableHead>Role</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead className="w-[50px]"></TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {users.map((user) => (
                            <TableRow key={`user-${user.id}`}>
                              <TableCell>
                                <div className="flex items-center gap-3">
                                  <Avatar className="h-8 w-8">
                                    <AvatarFallback>{getInitials(user)}</AvatarFallback>
                                  </Avatar>
                                  <div>
                                    <div className="font-medium">{getDisplayName(user)}</div>
                                    <div className="text-sm text-muted-foreground">{user.email}</div>
                                  </div>
                                </div>
                              </TableCell>
                              <TableCell>{getRole(user)}</TableCell>
                              <TableCell>
                                {renderStatusBadge(user)}
                              </TableCell>
                              <TableCell>
                                <DropdownMenu 
                                  open={openDropdownId === user.id}
                                  onOpenChange={(open) => {
                                    setOpenDropdownId(open ? user.id : null);
                                  }}
                                >
                                  <DropdownMenuTrigger asChild>
                                    <Button 
                                      variant="ghost" 
                                      size="sm"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                      }}
                                    >
                                      <MoreHorizontal className="h-4 w-4" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent className="w-48 bg-white dark:bg-gray-800" align="end">
                                    {getStatus(user) === 'PENDING_INVITE' && (
                                      <DropdownMenuItem 
                                        onClick={(e) => {
                                          e.stopPropagation();
                                          handleResendInviteClick(user);
                                        }}
                                        className="cursor-pointer"
                                        disabled={!!resendingInviteUserId}
                                      >
                                        <Mail className="h-4 w-4 mr-2" />
                                        {resendingInviteUserId === user.id ? 'Resending...' : 'Resend invite'}
                                      </DropdownMenuItem>
                                    )}
                                    <DropdownMenuItem 
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleDeleteClick(user);
                                      }}
                                      className="text-destructive focus:text-destructive cursor-pointer"
                                      disabled={!!deletingUserId}
                                    >
                                      <Trash2 className="h-4 w-4 mr-2" />
                                      {deletingUserId === user.id ? 'Deleting...' : 'Delete user'}
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
              </TabsContent>

              <TabsContent value="security" className="space-y-6">
                {/* Allowed Email Domains Section */}
                <Card>
                  <CardHeader>
                    <CardTitle>Allowed Email Domains</CardTitle>
                    <CardDescription>
                      Anyone with email addresses at these domains can automatically join this organization.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    {/* Add Domain Input */}
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={allowedDomains}
                        onChange={(e) => setAllowedDomains(e.target.value)}
                        placeholder="Enter email domain"
                        className="flex-1 px-3 py-2 border rounded-md"
                      />
                      <Button 
                        onClick={handleSaveDomains}
                        disabled={isSavingDomains || !allowedDomains.trim()}
                      >
                        <Plus className="h-4 w-4 mr-2" />
                        Add
                      </Button>
                    </div>

                    {/* Current Domains List */}
                    {domainsLoading ? (
                      <div className="flex items-center justify-center py-4">
                        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                      </div>
                    ) : domains.length > 0 ? (
                      <div className="space-y-3">
                        {domains.map((domain) => (
                          <div key={domain.id} className="flex items-center justify-between p-3 border rounded-lg">
                            <div className="flex items-center gap-3">
                              <Globe className="h-4 w-4 text-muted-foreground" />
                              <div>
                                <div className="font-medium">{domain.name}</div>
                                <div className="text-sm text-muted-foreground">
                                  {domain.created_at ? formatDomainTime(domain.created_at) : "Added recently"}
                                </div>
                              </div>
                            </div>
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm">
                                  <MoreVertical className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem 
                                  className="text-destructive"
                                  onClick={() => handleDeleteDomain(domain.id)}
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  Remove
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-center py-8 text-muted-foreground">
                        <Globe className="h-12 w-12 mx-auto mb-4 opacity-50" />
                        <p>No allowed email domains configured yet.</p>
                        <p className="text-sm">Add a domain to get started.</p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="advanced" className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Advanced Settings</CardTitle>
                    <CardDescription>Configure advanced organization settings including SSO and SCIM</CardDescription>
                  </CardHeader>
                  <CardContent>
                    {portalLoading ? (
                      <div className="flex items-center justify-center py-8">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                        <span className="ml-2">Loading configuration...</span>
                      </div>
                    ) : portalError ? (
                      <div className="text-center py-8">
                        <p className="text-destructive">{portalError}</p>
                        {retryCount < MAX_RETRIES ? (
                          <>
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mt-4"></div>
                            <p className="text-sm text-muted-foreground mt-2">
                              Retrying... ({retryCount + 1}/{MAX_RETRIES})
                            </p>
                          </>
                        ) : (
                          <>
                            <p className="text-sm text-muted-foreground mt-2">
                              Failed to load after {MAX_RETRIES} attempts
                            </p>
                            <Button 
                              onClick={() => {
                                setRetryCount(0);
                                setPortalError(null);
                                setHasFetchedPortal(false);
                                fetchPortalLink();
                              }} 
                              className="mt-2"
                            >
                              Try Again
                            </Button>
                          </>
                        )}
                      </div>
                    ) : portalUrl ? (
                      <div className="w-full">
                        <iframe
                          src={portalUrl}
                          className="w-full h-[600px] border rounded-lg"
                          title="Advanced Configuration"
                          sandbox="allow-same-origin allow-scripts allow-forms allow-popups"
                        />
                      </div>
                    ) : (
                      <div className="text-center py-8">
                        <p className="text-muted-foreground">
                          No portal URL available.
                        </p>
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mt-4"></div>
                        <p className="text-sm text-muted-foreground mt-2">Loading configuration...</p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>
        </div>
      </div>

      {/* Delete confirmation dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete {userToDelete?.email}? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={!!deletingUserId}>Cancel</AlertDialogCancel>
            <AlertDialogAction 
              onClick={() => userToDelete && handleDeleteUser(userToDelete)}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={!!deletingUserId}
            >
              {deletingUserId ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Invite user form */}
      <InviteUserForm 
        open={inviteFormOpen}
        onOpenChange={setInviteFormOpen}
        onSuccess={fetchMembers}
      />
    </div>
    </TooltipProvider>
    </RBACGuard>
  );
}
