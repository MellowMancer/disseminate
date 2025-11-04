import React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

interface PublicRouteProps {
  children: React.ReactElement;
}

/**
 * PublicRoute redirects authenticated users away from public pages like /login and /signup
 * to the home page. Unauthenticated users can access these pages normally.
 */
export const PublicRoute = ({ children }: PublicRouteProps) => {
  const { authenticated } = useAuth();

  if (authenticated) {
    return <Navigate to="/" replace />;
  }

  return children;
};

