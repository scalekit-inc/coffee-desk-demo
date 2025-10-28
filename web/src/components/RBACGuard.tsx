import { ReactNode } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

// Utility function to check if user has a specific permission
export const hasPermission = (userPermissions: string[] | undefined, permission: string): boolean => {
  return userPermissions?.includes(permission) || false;
};

interface RBACGuardProps {
  children: ReactNode;
  requireAdmin?: boolean;
  requirePermission?: string | string[];
  fallback?: ReactNode;
}

export function RBACGuard({ children, requireAdmin = false, requirePermission, fallback }: RBACGuardProps) {
  const { user, loading } = useAuth();

  // Wait for auth to finish loading to avoid access denied flicker
  if (loading) {
    return (
      <div className="w-full flex flex-col items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <p className="mt-3 text-sm text-muted-foreground">Loading… Please wait.</p>
      </div>
    );
  }

  if (!user) {
    return fallback || (
      <Card className="w-full max-w-md mx-auto mt-8">
        <CardHeader>
          <CardTitle>Access Denied</CardTitle>
          <CardDescription>
            You need to be authenticated to access this content.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  // Check for specific permission(s)
  if (requirePermission) {
    const hasRequiredPermission = Array.isArray(requirePermission) 
      ? requirePermission.some(permission => hasPermission(user.permissions, permission))
      : hasPermission(user.permissions, requirePermission);
    
    if (!hasRequiredPermission) {
      const permissionText = Array.isArray(requirePermission) 
        ? requirePermission.join('" or "')
        : requirePermission;
      
      return fallback || (
        <Card className="w-full max-w-md mx-auto mt-8">
          <CardHeader>
            <CardTitle>Access Denied</CardTitle>
            <CardDescription>
              You need the "{permissionText}" permission to access this content.
            </CardDescription>
          </CardHeader>
        </Card>
      );
    }
  }

  // Legacy admin check (for backward compatibility)
  if (requireAdmin && !user.is_admin) {
    return fallback || (
      <Card className="w-full max-w-md mx-auto mt-8">
        <CardHeader>
          <CardTitle>Access Denied</CardTitle>
          <CardDescription>
            You need administrator privileges to access this content.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return <>{children}</>;
} 