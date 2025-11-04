import { createContext, useContext, useState, useEffect, type ReactNode, useMemo } from "react";

export type Theme = "lavender" | "light" | "dark";

interface ThemeContextType {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  cycleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState<Theme>(() => {
    // Get theme from localStorage or default to lavender
    const savedTheme = localStorage.getItem("theme") as Theme | null;
    return savedTheme || "lavender";
  });

  useEffect(() => {
    // Apply theme to document root
    const root = document.documentElement;
    
    // Remove all theme attributes
    delete root.dataset.theme;
    
    // Apply the current theme (lavender is default, doesn't need data-theme)
    if (theme === "light") {
      root.dataset.theme = "light";
    } else if (theme === "dark") {
      root.dataset.theme = "dark";
    }
    
    // Save to localStorage
    localStorage.setItem("theme", theme);
  }, [theme]);

  const cycleTheme = () => {
    setTheme((currentTheme) => {
      if (currentTheme === "lavender") return "light";
      if (currentTheme === "light") return "dark";
      return "lavender";
    });
  };

  const value = useMemo(
    () => ({ theme, setTheme, cycleTheme }),
    [theme]
  );

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
};

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
};

