import React from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { LogOut } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

const Layout: React.FC = () => {
  const { authenticated, setAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    fetch("/auth/logout", { method: "POST", credentials: "include" }).then(() => {
      setAuthenticated(false);
      navigate("/login");
    });
  };

  return (
    <div>
      {authenticated && (
        <div style={{ position: "fixed", top: 12, right: 12, zIndex: 1000 }}>
          <Button size="sm" onClick={handleLogout} aria-label="Logout">
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </Button>
        </div>
      )}
      <Outlet context={{ authenticated }} />
    </div>
  );
};

export default Layout;
