
import { ReactNode, useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useAuth } from '@/hooks/useAuth';

interface AuthWrapperProps {
  children: ReactNode;
  requireAuth?: boolean;
}

export const AuthWrapper = ({ children, requireAuth = false }: AuthWrapperProps) => {
  const { isAuthenticated, user, loading, error } = useAuth();
  const location = useLocation();
  const [isPostLogout, setIsPostLogout] = useState(false);

  // Check if we're potentially coming from a logout redirect
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const wasLoggedOut = urlParams.has('logged_out') || sessionStorage.getItem('logout_redirect');
    
    if (wasLoggedOut) {
      setIsPostLogout(true);
      sessionStorage.removeItem('logout_redirect');
      
      // Small delay to allow backend to process logout
      setTimeout(() => {
        setIsPostLogout(false);
      }, 500);
    }
  }, [location.pathname]);

  if (loading || isPostLogout) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">
            {isPostLogout ? "Processing logout..." : "Checking session..."}
          </p>
        </div>
      </div>
    );
  }

  if (requireAuth && !isAuthenticated) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle>Access Denied</CardTitle>
            <CardDescription>
              You need to be authenticated to access this page.
              {error && <div className="text-destructive text-sm mt-2">{error}</div>}
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <Button onClick={() => window.location.href = '/'}>
              Go to Home
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return <>{children}</>;
};
