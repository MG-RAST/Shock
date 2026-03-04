import { NavLink } from "react-router-dom";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";
import {
  Database,
  Upload,
  Settings,
  Shield,
  MapPin,
  Lock,
  Activity,
  LayoutDashboard,
  PanelLeftClose,
  PanelLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
    isActive
      ? "bg-accent text-accent-foreground"
      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
  );

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { isAdmin } = useAuth();

  return (
    <aside
      className={cn(
        "flex h-full flex-col border-r bg-card transition-all duration-200",
        collapsed ? "w-14" : "w-56"
      )}
    >
      <div className="flex h-14 items-center border-b px-3">
        {!collapsed && (
          <span className="text-lg font-bold tracking-tight">Shock</span>
        )}
        <Button
          variant="ghost"
          size="icon"
          className={cn("ml-auto", collapsed && "mx-auto")}
          onClick={onToggle}
        >
          {collapsed ? <PanelLeft className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
        </Button>
      </div>

      <nav className="flex-1 space-y-1 p-2">
        <NavLink to="/ui/nodes" className={navLinkClass}>
          <Database className="h-4 w-4 shrink-0" />
          {!collapsed && <span>Nodes</span>}
        </NavLink>
        <NavLink to="/ui/upload" className={navLinkClass}>
          <Upload className="h-4 w-4 shrink-0" />
          {!collapsed && <span>Upload</span>}
        </NavLink>
        <NavLink to="/ui/settings" className={navLinkClass}>
          <Settings className="h-4 w-4 shrink-0" />
          {!collapsed && <span>Settings</span>}
        </NavLink>

        {isAdmin && (
          <>
            <div className={cn("my-3 border-t", collapsed && "mx-1")} />
            {!collapsed && (
              <p className="px-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Admin
              </p>
            )}
            <NavLink to="/ui/admin" end className={navLinkClass}>
              <LayoutDashboard className="h-4 w-4 shrink-0" />
              {!collapsed && <span>Dashboard</span>}
            </NavLink>
            <NavLink to="/ui/admin/locations" className={navLinkClass}>
              <MapPin className="h-4 w-4 shrink-0" />
              {!collapsed && <span>Locations</span>}
            </NavLink>
            <NavLink to="/ui/admin/locker" className={navLinkClass}>
              <Lock className="h-4 w-4 shrink-0" />
              {!collapsed && <span>Locker</span>}
            </NavLink>
            <NavLink to="/ui/admin/trace" className={navLinkClass}>
              <Activity className="h-4 w-4 shrink-0" />
              {!collapsed && <span>Trace</span>}
            </NavLink>
          </>
        )}
      </nav>

      {!collapsed && (
        <div className="border-t p-3">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Shield className="h-3 w-3" />
            <span>{isAdmin ? "Admin" : "User"}</span>
          </div>
        </div>
      )}
    </aside>
  );
}
