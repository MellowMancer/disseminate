import { createContext, useContext, useState, type ReactNode, useEffect, useMemo } from "react";

interface AuthContextType {
  authenticated: boolean;
  setAuthenticated: (value: boolean) => void;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [authenticated, setAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true); // New loading state

  useEffect(() => {
    fetch("/auth/status", { credentials: "include" })
      .then((res) => res.json())
      .then((data) => {
          setAuthenticated(data.authenticated);
          setLoading(false);
      })
      .catch(() => {
        setAuthenticated(false);
        setLoading(false); // Done loading even on error
      });
  }, []);

  const value = useMemo(
    () => ({ authenticated, setAuthenticated, loading }),
    [authenticated, loading]
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};


export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
