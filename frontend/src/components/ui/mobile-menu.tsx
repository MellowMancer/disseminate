"use client"

import { useState } from "react";
import { Menu, X } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { cn } from "@/lib/utils";

export function MobileMenu() {
    const [isOpen, setIsOpen] = useState(false);
    const { authenticated, setAuthenticated } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        fetch("/auth/logout", { method: "POST", credentials: "include" }).then(() => {
            setAuthenticated(false);
            navigate("/login");
            setIsOpen(false);
        });
    };

    const handleNavigation = (path: string) => {
        navigate(path);
        setIsOpen(false);
    };

    return (
        <>
            {/* Hamburger Button - Fixed at bottom center */}
            <Button
                variant="outline"
                size="icon"
                onClick={() => setIsOpen(!isOpen)}
                className="md:hidden fixed bottom-4 left-1/2 -translate-x-1/2 z-50 shadow-lg"
                aria-label="Toggle menu"
            >
                {isOpen ? <X className="h-[1.2rem] w-[1.2rem]" /> : <Menu className="h-[1.2rem] w-[1.2rem]" />}
            </Button>

            {/* Overlay */}
            {isOpen && (
                <div
                    className="md:hidden fixed inset-0 bg-black/50 z-40"
                    onClick={() => setIsOpen(false)}
                />
            )}

            {/* Menu Panel */}
            <div
                className={cn(
                    "md:hidden fixed bottom-20 left-1/2 -translate-x-1/2 z-40 w-[90vw] max-w-sm bg-card border-2 border-border rounded-lg shadow-lg transition-all duration-300",
                    isOpen ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4 pointer-events-none"
                )}
            >
                <div className="p-6 space-y-2">
                    {/* Navigation Items */}
                    <Button
                        variant="ghost"
                        className="w-full justify-start text-base h-12 text-foreground"
                        onClick={() => handleNavigation("/")}
                    >
                        Home
                    </Button>
                    
                    <Button
                        variant="ghost"
                        className="w-full justify-start text-base h-12 text-foreground"
                        onClick={() => {}}
                    >
                        About
                    </Button>

                    {authenticated && (
                        <Button
                            variant="ghost"
                            className="w-full justify-start text-base h-12 text-foreground"
                            onClick={() => handleNavigation("/profile")}
                        >
                            Profile & Keys
                        </Button>
                    )}

                    <Button
                        variant="ghost"
                        className="w-full justify-start text-base h-12 text-foreground"
                        onClick={authenticated ? handleLogout : () => handleNavigation("/login")}
                    >
                        {authenticated ? "Logout" : "Login"}
                    </Button>

                    {/* Theme Toggle */}
                    <div className="pt-4 mt-2 border-t border-border flex items-center justify-between">
                        <span className="text-sm font-medium text-primary">Theme</span>
                        <ThemeToggle variant="ghost" />
                    </div>
                </div>
            </div>
        </>
    );
}

