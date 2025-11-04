import { Moon, Sun, Palette } from "lucide-react";
import { useTheme } from "@/context/ThemeContext";
import { Button } from "@/components/ui/button";

interface ThemeToggleProps {
  className?: string;
  variant?: "outline" | "ghost" | "default";
}

export function ThemeToggle({ className, variant = "outline" }: Readonly<ThemeToggleProps>) {
  const { theme, cycleTheme } = useTheme();

  const getIcon = () => {
    switch (theme) {
      case "lavender":
        return <Palette className="h-[1.2rem] w-[1.2rem]" />;
      case "light":
        return <Sun className="h-[1.2rem] w-[1.2rem]" />;
      case "dark":
        return <Moon className="h-[1.2rem] w-[1.2rem]" />;
    }
  };

  const getThemeName = () => {
    switch (theme) {
      case "lavender":
        return "Lavender";
      case "light":
        return "Light";
      case "dark":
        return "Dark";
    }
  };

  return (
    <Button
      variant={variant}
      size="icon"
      onClick={cycleTheme}
      title={`Current theme: ${getThemeName()}. Click to cycle.`}
      className={className}
    >
      {getIcon()}
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}

