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
import { AlertCircle, Search, User, LogOut, Loader2, CheckCircle2 } from "lucide-react";
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

  // 🆕 Check if we just returned from OAuth
  useEffect(() => {
    checkOAuthReturn();
    loadConnectionStatuses();
  }, []);

  // 🆕 Check if user just completed OAuth flow
  const checkOAuthReturn = () => {
    const pendingProvider = localStorage.getItem('oauth_pending');
    if (pendingProvider) {
      console.log(`Returned from ${pendingProvider} OAuth`);
      localStorage.removeItem('oauth_pending');
      
      // Show success message
      toast({
        title: "Connection Successful",
        description: `${pendingProvider} integration enabled successfully`,
      });
    }
  };

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

  // ✅ FIXED: Clean redirect-based OAuth flow
  const handleSlackToggle = async (enabled: boolean) => {
    setLoadingSlack(true);
    setError(null);
    
    try {
      if (enabled) {
        // Enable Slack - open magic link in new tab
        const result = await configureSlack(true);
        
        if (result.link) {
          console.log('🚀 Opening Slack OAuth in new tab:', result.link);
          
          // Store that we're doing Slack OAuth
          localStorage.setItem('oauth_pending', 'Slack');
          
          // Open magic link in new tab
          window.open(result.link, '_blank', 'noopener,noreferrer');
          
          // Clear loading state since we're not redirecting
          setLoadingSlack(false);
          
          toast({
            title: "OAuth Link Opened",
            description: "Please complete the authorization in the new tab",
          });
          return;
        } else {
          // No link (shouldn't happen for enable)
          console.warn('No link returned from backend');
          setSlackEnabled(true);
          toast({
            title: "Success",
            description: "Slack integration enabled",
          });
        }
      } else {
        // Disable Slack
        const result = await configureSlack(false);
        
        if (result.success) {
          setSlackEnabled(false);
          setSlackConnectedAt(null);
          toast({
            title: "Success",
            description: "Slack integration disabled successfully",
          });
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred';
      console.error('❌ Slack toggle error:', err);
      setError(errorMessage);
      
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoadingSlack(false);
    }
  };

  //Clean redirect-based OAuth flow
  const handleGithubToggle = async (enabled: boolean) => {
    setLoadingGithub(true);
    setError(null);
    
    try {
      if (enabled) {
        // Enable GitHub - open magic link in new tab
        const result = await configureGithub(true);
        
        if (result.link) {
          console.log('🚀 Opening GitHub OAuth in new tab:', result.link);
          
          // Store that we're doing GitHub OAuth
          localStorage.setItem('oauth_pending', 'GitHub');
          
          // Open magic link in new tab
          window.open(result.link, '_blank', 'noopener,noreferrer');
          
          // Clear loading state since we're not redirecting
          setLoadingGithub(false);
          
          toast({
            title: "OAuth Link Opened",
            description: "Please complete the authorization in the new tab",
          });
          return;
        } else {
          // No link (shouldn't happen for enable)
          console.warn('No link returned from backend');
          setGithubEnabled(true);
          toast({
            title: "Success",
            description: "GitHub integration enabled",
          });
        }
      } else {
        // Disable GitHub
        const result = await configureGithub(false);
        
        if (result.success) {
          setGithubEnabled(false);
          setGithubConnectedAt(null);
          toast({
            title: "Success",
            description: "GitHub integration disabled successfully",
          });
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred';
      console.error('❌ GitHub toggle error:', err);
      setError(errorMessage);
      
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

  const emailInitial = user?.email ? user.email[0].toUpperCase() : "U";

  if (initialLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-muted-foreground">Loading integrations...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen w-full">
      {/* Top section with logo and workspace dropdown */}
      <div className="flex items-center justify-between p-4 border-b bg-background z-10">
        <div className="flex items-center gap-4">
          <img 
            src="/uploads/name_icon.png" 
            alt="Coffeedesk Logo" 
            className="h-8 w-auto"
          />
          <WorkspaceDropdown />
        </div>
        
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

      <div className="flex flex-1 overflow-hidden">
        <div className="w-64 flex-shrink-0 border-r bg-background">
          <AppSidebar />
        </div>

        <div className="flex-1 overflow-auto">
          <div className="p-6">
            <div className="mb-6">
              <h1 className="text-3xl font-bold">Integrations</h1>
              <p className="text-muted-foreground">
                Manage your workspace integrations and connect with third-party services
              </p>
            </div>

            {error && (
              <Alert variant="destructive" className="mb-6">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="grid gap-6 max-w-4xl">
              {/* Slack Integration Card */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                  <div className="space-y-1">
                    <CardTitle className="flex items-center gap-2">
                      Slack
                      {slackEnabled && <CheckCircle2 className="h-5 w-5 text-green-500" />}
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
                  </div>
                </CardContent>
              </Card>

              {/* GitHub Integration Card */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                  <div className="space-y-1">
                    <CardTitle className="flex items-center gap-2">
                      GitHub
                      {githubEnabled && <CheckCircle2 className="h-5 w-5 text-green-500" />}
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
                  </div>
                </CardContent>
              </Card>
            </div>

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