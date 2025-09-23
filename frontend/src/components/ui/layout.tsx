import React from "react";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { LogOut, LogInIcon } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

const Layout: React.FC = () => {
  const { authenticated, setAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    fetch("/auth/logout", { method: "POST", credentials: "include" }).then(() => {
      setAuthenticated(false);
      navigate("/login");
    });
  };

  const handleLoginButton = () => {
    navigate("/login");
  }

  return (
    <div className="h-screen w-screen overflow-auto bg-background py-16 px-8 md:py-8 md:px-24 mx-auto ">
      {authenticated && (
        <div className="fixed top-3 left-3 z-1000">
          <Button variant="default" size="lg" onClick={handleLogout} aria-label="Logout">
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </Button>
        </div>
      )}
      {!authenticated && location.pathname != "/login" && location.pathname != "/signup" && (
        <div className="fixed top-3 left-3 z-1000">
          <Button variant="default" size="lg" onClick={handleLoginButton} aria-label="toLogin">
            <LogInIcon className="mr-2 h-4 w-4" />
            Login
          </Button>
        </div>
      )}
      <Outlet context={{ authenticated }} />
    </div>
  );
};

export default Layout;
