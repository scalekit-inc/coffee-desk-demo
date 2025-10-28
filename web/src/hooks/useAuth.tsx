
import { useState, useEffect, useRef } from 'react';
import { config } from '@/config';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  name: string;
  roles: string[];
  is_admin: boolean;
  permissions: string[];
}

interface Workspace {
  display_name: string;
  id: string;
  is_current: boolean;
}

interface AuthResponse {
  authenticated: boolean;
  user?: User;
  workspaces?: Workspace[];
}

interface AuthState {
  isAuthenticated: boolean | null;
  user: User | null;
  workspaces: Workspace[];
  currentWorkspace: Workspace | null;
  loading: boolean;
  error: string | null;
}

export function useAuth() {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: null,
    user: null,
    workspaces: [],
    currentWorkspace: null,
    loading: true,
    error: null,
  });

  const hasCheckedSession = useRef(false);

  const checkSession = async () => {
    // Prevent multiple simultaneous calls
    if (hasCheckedSession.current) {
      return;
    }
    
    hasCheckedSession.current = true;

    try {
      setAuthState(prev => ({ ...prev, loading: true, error: null }));
      
      const response = await fetch(`${config.backendUrl}/api/session`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (response.ok) {
        const data: AuthResponse = await response.json();
        
        const currentWorkspace = data.workspaces?.find(ws => ws.is_current) || null;
        
        setAuthState({
          isAuthenticated: data.authenticated,
          user: data.user || null,
          workspaces: data.workspaces || [],
          currentWorkspace,
          loading: false,
          error: null,
        });
      } else {
        setAuthState({
          isAuthenticated: false,
          user: null,
          workspaces: [],
          currentWorkspace: null,
          loading: false,
          error: `Session check failed with status: ${response.status}`,
        });
      }
    } catch (error) {
      console.error("Session check failed:", error);
      setAuthState({
        isAuthenticated: false,
        user: null,
        workspaces: [],
        currentWorkspace: null,
        loading: false,
        error: error instanceof Error ? error.message : 'Session check failed',
      });
    }
  };

  const refetchSession = () => {
    hasCheckedSession.current = false;
    checkSession();
  };

  // Only check session once when hook is first used
  useEffect(() => {
    if (!hasCheckedSession.current) {
      checkSession();
    }
  }, []);

  return {
    ...authState,
    refetchSession,
  };
}
