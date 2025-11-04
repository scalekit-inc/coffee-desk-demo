import { LayoutDashboard, Users, Settings, CreditCard, FolderOpen, CheckSquare } from "lucide-react";
import { NavLink, useLocation, Link } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { hasPermission } from "./RBACGuard";

const navigationItems = [
  { title: "Overview", url: "/dashboard", icon: LayoutDashboard },
  { title: "Projects", url: "/dashboard/projects", icon: FolderOpen },
  { title: "Tasks", url: "/dashboard/tasks", icon: CheckSquare },
];

export function AppSidebar() {
  const { user } = useAuth();
  
  // RBAC: Only show workspace section for users with organization:settings permission
  const canAccessWorkspace = hasPermission(user?.permissions, "organization:settings");
  
  // RBAC: Show billing section for users with organization:settings or workspace:billing permission
  const canAccessBilling = hasPermission(user?.permissions, "organization:settings") || hasPermission(user?.permissions, "workspace:billing");

  const location = useLocation();
  const currentPath = location.pathname;

  const isActive = (path: string) => {
    if (path === "/dashboard") {
      return currentPath === "/dashboard";
    }
    return currentPath.startsWith(path);
  };

  const getNavClasses = (path: string) => {
    return isActive(path) 
      ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium" 
      : "hover:bg-sidebar-accent/50 hover:text-sidebar-accent-foreground";
  };

  return (
    <nav className="flex flex-col gap-2 p-4">
      <div className="space-y-2">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          Overview
        </h3>
        <Link
          to="/dashboard"
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          <LayoutDashboard className="h-4 w-4" />
          Dashboard
        </Link>
      </div>

      {/* Navigation Items */}
      <div>
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
          Navigation
        </h3>
        <nav className="space-y-1">
          {navigationItems.map((item) => (
            <NavLink 
              key={item.title}
              to={item.url} 
              className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors text-sm ${getNavClasses(item.url)}`}
            >
              <item.icon className="h-4 w-4" />
              <span>{item.title}</span>
            </NavLink>
          ))}
        </nav>
      </div>

      {/* Bottom Settings Section */}
      <div className="mt-auto">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
          Settings
        </h3>
        <nav className="space-y-1">
          {canAccessWorkspace && (
            <Link
              to="/dashboard/workspace"
              className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors text-sm ${getNavClasses("/dashboard/workspace")}`}
            >
              <Settings className="h-4 w-4" />
              <span>Organization</span>
            </Link>
          )}
          {canAccessBilling && (
            <Link
              to="/dashboard/billing"
              className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors text-sm ${getNavClasses("/dashboard/billing")}`}
            >
              <CreditCard className="h-4 w-4" />
              <span>Billing</span>
            </Link>
          )}
          <NavLink 
            to="/dashboard/profile" 
            className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors text-sm ${getNavClasses("/dashboard/profile")}`}
          >
            <Users className="h-4 w-4" />
            <span>My Profile</span>
          </NavLink>
        </nav>
      </div>
    </nav>
  );
}
