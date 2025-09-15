import React, { createContext, useContext, useState, useEffect } from "react";

interface AuthContextType {
  isAuthenticated: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    const localLogin = localStorage.getItem("localLoggedIn") === "true";
    setAuthenticated(localLogin /* || additional cloud login check */);
  }, []);

  const login = () => setAuthenticated(true);
  const logout = () => {
    setAuthenticated(false);
    localStorage.removeItem("localLoggedIn");
    // TODO: Backend Logout
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
