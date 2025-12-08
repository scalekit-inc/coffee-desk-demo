import { useState, useEffect } from 'react';
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { AppSidebar } from "@/components/AppSidebar";
import { WorkspaceDropdown } from "@/components/WorkspaceDropdown";
import { useAuth } from "@/hooks/useAuth";
import { config } from "@/config";
import { AlertCircle, Search, User, LogOut, Loader2 } from "lucide-react";
import { getSlackStatus, configureSlack, getGithubStatus, configureGithub } from '../api/connections';
import { useToast } from '@/hooks/use-toast';

export default function Connections() {
  const [slackEnabled, setSlackEnabled] = useState(false);
  const [githubEnabled, setGithubEnabled] = useState(false);
  const [slackConnectedAt, setSlackConnectedAt] = useState<string | null>(null);
  const [githubConnectedAt, setGithubConnectedAt] = useState<string | null>(null);
  const [loadingSlack, setLoadingSlack] = useState(false);
  const [loadingGithub, setLoadingGithub] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const navigate = useNavigate();
  const { user } = useAuth();
  const { toast } = useToast();

  // Load initial connection statuses
  useEffect(() => {
    loadConnectionStatuses();
  }, []);

  const loadConnectionStatuses = async () => {
    try {
      const [slackStatus, githubStatus] = await Promise.all([
        getSlackStatus(),
        getGithubStatus(),
      ]);

      setSlackEnabled(slackStatus.connection?.enabled || false);
      setSlackConnectedAt(slackStatus.connection?.connected_at || null);
      
      setGithubEnabled(githubStatus.connection?.enabled || false);
      setGithubConnectedAt(githubStatus.connection?.connected_at || null);
    } catch (err) {
      console.error('Failed to load connection statuses:', err);
      setError('Failed to load connection statuses. Please refresh the page.');
    } finally {
      setInitialLoading(false);
    }
  };

  const handleSlackToggle = async (enabled: boolean) => {
    setLoadingSlack(true);
    setError(null);
    
    try {
      if (enabled) {
        // Enable Slack - need to handle OAuth flow
        const result = await configureSlack(true);
        
        if (result.authUrl) {
          // Open OAuth popup
          const popup = window.open(
            result.authUrl,
            'slack-oauth',
            'width=600,height=700,left=100,top=100'
          );

          if (!popup) {
            throw new Error('Popup blocked. Please allow popups for this site.');
          }

          // Poll for connection completion
          const checkInterval = setInterval(async () => {
            if (popup.closed) {
              clearInterval(checkInterval);
              
              // Check if connection was successful
              const status = await getSlackStatus();
              const isConnected = status.connection?.enabled || false;
              
              setSlackEnabled(isConnected);
              setSlackConnectedAt(status.connection?.connected_at || null);
              setLoadingSlack(false);
              
              if (isConnected) {
                toast({
                  title: "Success",
                  description: "Slack integration enabled successfully",
                });
              } else {
                toast({
                  title: "Cancelled",
                  description: "Slack connection was not completed",
                  variant: "destructive",
                });
              }
            }
          }, 1000);

          // Timeout after 5 minutes
          setTimeout(() => {
            clearInterval(checkInterval);
            if (!popup.closed) {
              popup.close();
            }
            setLoadingSlack(false);
          }, 300000);
          
          return; // Don't clear loading state yet
        } else {
          // No auth URL returned, update status
          setSlackEnabled(true);
          toast({
            title: "Success",
            description: "Slack integration enabled successfully",
          });
        }
      } else {
        // Disable Slack
        await configureSlack(false);
        setSlackEnabled(false);
        setSlackConnectedAt(null);
        toast({
          title: "Success",
          description: "Slack integration disabled successfully",
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred';
      console.error('Slack toggle error:', err);
      setError(errorMessage);
      setSlackEnabled(!enabled); // Revert toggle
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoadingSlack(false);
    }
  };

  const handleGithubToggle = async (enabled: boolean) => {
    setLoadingGithub(true);
    setError(null);
    
    try {
      if (enabled) {
        // Enable GitHub - need to handle OAuth flow
        const result = await configureGithub(true);
        
        if (result.authUrl) {
          // Open OAuth popup
          const popup = window.open(
            result.authUrl,
            'github-oauth',
            'width=600,height=700,left=100,top=100'
          );

          if (!popup) {
            throw new Error('Popup blocked. Please allow popups for this site.');
          }

          // Poll for connection completion
          const checkInterval = setInterval(async () => {
            if (popup.closed) {
              clearInterval(checkInterval);
              
              // Check if connection was successful
              const status = await getGithubStatus();
              const isConnected = status.connection?.enabled || false;
              
              setGithubEnabled(isConnected);
              setGithubConnectedAt(status.connection?.connected_at || null);
              setLoadingGithub(false);
              
              if (isConnected) {
                toast({
                  title: "Success",
                  description: "GitHub integration enabled successfully",
                });
              } else {
                toast({
                  title: "Cancelled",
                  description: "GitHub connection was not completed",
                  variant: "destructive",
                });
              }
            }
          }, 1000);

          // Timeout after 5 minutes
          setTimeout(() => {
            clearInterval(checkInterval);
            if (!popup.closed) {
              popup.close();
            }
            setLoadingGithub(false);
          }, 300000);
          
          return; // Don't clear loading state yet
        } else {
          // No auth URL returned, update status
          setGithubEnabled(true);
          toast({
            title: "Success",
            description: "GitHub integration enabled successfully",
          });
        }
      } else {
        // Disable GitHub
        await configureGithub(false);
        setGithubEnabled(false);
        setGithubConnectedAt(null);
        toast({
          title: "Success",
          description: "GitHub integration disabled successfully",
        });
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred';
      console.error('GitHub toggle error:', err);
      setError(errorMessage);
      setGithubEnabled(!enabled); // Revert toggle
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoadingGithub(false);
    }
  };

  const handleLogout = () => {
    const redirectUri = window.location.href;
    window.location.href = `${config.backendUrl}/api/logout?redirect_uri=${encodeURIComponent(redirectUri)}`;
    setIsProfileOpen(false);
  };

  const handleProfile = () => {
    navigate("/dashboard/profile");
    setIsProfileOpen(false);
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return null;
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric', 
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Get first letter of email for profile icon
  const emailInitial = user?.email ? user.email[0].toUpperCase() : "U";

  if (initialLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-muted-foreground">Loading connections...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen w-full">
      {/* Top section with logo and workspace dropdown */}
      <div className="flex items-center justify-between p-4 border-b bg-background z-10">
        <div className="flex items-center gap-4">
          {/* App logo in top-left */}
          <img 
            src="/uploads/name_icon.png" 
            alt="Coffeedesk Logo" 
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
          <div className="p-6">
            {/* Page header */}
            <div className="mb-6">
              <h1 className="text-3xl font-bold">Integrations</h1>
              <p className="text-muted-foreground">
                Manage your workspace integrations and connect with third-party services
              </p>
            </div>

            {/* Error Alert */}
            {error && (
              <Alert variant="destructive" className="mb-6">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* Integration Cards */}
            <div className="grid gap-6 max-w-4xl">
              {/* Slack Integration Card */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      {/*<span className="text-2xl">💬</span>*/}
                      Slack
                    </CardTitle>
                    <CardDescription>
                      Connect your workspace to Slack for notifications and updates
                    </CardDescription>
                  </div>
                  <Switch
                    checked={slackEnabled}
                    onCheckedChange={handleSlackToggle}
                    disabled={loadingSlack}
                  />
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <p className="text-sm text-muted-foreground">
                        {slackEnabled
                          ? 'Slack integration is active. Your workspace is connected.'
                          : 'Enable Slack integration to receive notifications and updates in your Slack workspace.'}
                      </p>
                      {loadingSlack && (
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                      )}
                    </div>
                    {slackEnabled && slackConnectedAt && (
                      <p className="text-xs text-muted-foreground">
                        Connected on {formatDate(slackConnectedAt)}
                      </p>
                    )}
                    {slackEnabled && (
                      <Button variant="outline" size="sm" disabled={loadingSlack}>
                        Manage Slack Settings
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* GitHub Integration Card */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      {/*<span className="text-2xl">🐙</span>*/}
                      GitHub
                    </CardTitle>
                    <CardDescription>
                      Connect your workspace to GitHub for code integration and automation
                    </CardDescription>
                  </div>
                  <Switch
                    checked={githubEnabled}
                    onCheckedChange={handleGithubToggle}
                    disabled={loadingGithub}
                  />
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <p className="text-sm text-muted-foreground">
                        {githubEnabled
                          ? 'GitHub integration is active. Your workspace is connected.'
                          : 'Enable GitHub integration to sync repositories and automate workflows.'}
                      </p>
                      {loadingGithub && (
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                      )}
                    </div>
                    {githubEnabled && githubConnectedAt && (
                      <p className="text-xs text-muted-foreground">
                        Connected on {formatDate(githubConnectedAt)}
                      </p>
                    )}
                    {githubEnabled && (
                      <Button variant="outline" size="sm" disabled={loadingGithub}>
                        Manage GitHub Settings
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Refresh Button */}
            <div className="mt-6 max-w-4xl">
              <Button 
                variant="outline" 
                size="sm"
                onClick={loadConnectionStatuses}
                disabled={initialLoading}
              >
                🔄 Refresh Status
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}