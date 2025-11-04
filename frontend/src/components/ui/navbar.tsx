"use client"

import { useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { MobileMenu } from "@/components/ui/mobile-menu";

import {
    NavigationMenu,
    NavigationMenuItem,
    NavigationMenuLink,
    NavigationMenuList,
    navigationMenuTriggerStyle,
    NavigationMenuIndicator
} from "@/components/ui/navigation-menu"

export function Navbar() {
    const { authenticated, setAuthenticated } = useAuth();
    const navigate = useNavigate();


    const handleLogout = () => {
        fetch("/auth/logout", { method: "POST", credentials: "include" }).then(() => {
            setAuthenticated(false);
            navigate("/login");
        });
    };

    const handleLoginButton = () => {
        navigate("/login");
    }

    const handleProfileButton = () => {
        navigate("/profile");
    }

    const handleHomeButton = () => {
        navigate("/");
    }

    const handleAboutButton = () => {

    }

    return (
        <>
            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center justify-center gap-3">
                <NavigationMenu>
                    <NavigationMenuList>
                        <NavigationMenuItem>
                            <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                                <button
                                    onClick={handleHomeButton}
                                    aria-label="Home"
                                    className="bg-transparent border-none p-0 m-0 cursor-pointer"
                                    type="button"
                                >
                                    Home
                                </button>
                            </NavigationMenuLink>
                        </NavigationMenuItem>
                        <NavigationMenuItem>
                            <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                                <button
                                    onClick={handleAboutButton}
                                    aria-label="About"
                                    className="bg-transparent border-none p-0 m-0 cursor-pointer"
                                    type="button"
                                >
                                    About
                                </button>
                            </NavigationMenuLink>
                        </NavigationMenuItem>
                        {authenticated ?
                            <NavigationMenuItem>
                                <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>

                                    <button
                                        onClick={handleProfileButton}
                                        aria-label="Profile"
                                        className="bg-transparent border-none p-0 m-0 cursor-pointer"
                                        type="button"
                                    >
                                        Profile & Keys
                                    </button>

                                </NavigationMenuLink>
                            </NavigationMenuItem>
                            : <></>
                        }
                        <NavigationMenuItem>
                            <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                                {authenticated ? (<button
                                    onClick={handleLogout}
                                    aria-label="Logout"
                                    className="bg-transparent border-none p-0 m-0 cursor-pointer"
                                    type="button"
                                >
                                    Logout
                                </button>) : (<button
                                    onClick={handleLoginButton}
                                    aria-label="Login"
                                    className="bg-transparent border-none p-0 m-0 cursor-pointer"
                                    type="button"
                                >
                                    Login
                                </button>)
                                }
                            </NavigationMenuLink>
                        </NavigationMenuItem>
                        <NavigationMenuIndicator className="NavigationMenuIndicator" />
                    </NavigationMenuList>
                </NavigationMenu>
            </div>

            {/* Mobile Menu */}
            <MobileMenu />

            {/* Desktop Theme Toggle - Fixed at bottom center */}
            <div className="hidden md:block">
                <ThemeToggle className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 shadow-lg" />
            </div>
        </>
    )
}