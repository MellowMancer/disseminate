"use client"

import { useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

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
        <div className="flex items-center justify-center">
        <NavigationMenu>
            <NavigationMenuList>
                <NavigationMenuItem>
                    <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                        <div onClick={handleHomeButton} aria-label="Home" role="button">
                            Home
                        </div>
                    </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuItem>
                    <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                        <div onClick={handleAboutButton} aria-label="About" role="button">
                            About
                        </div>
                    </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuItem>
                    <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                        <div onClick={handleProfileButton} aria-label="Profile" role="button">
                            Profile & Keys
                        </div>
                    </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuItem>
                    <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
                        {authenticated ? (<div onClick={handleLogout} aria-label="Logout" role="button">
                            Logout
                        </div>) : (<div onClick={handleLoginButton} aria-label="Login" role="button">
                            Login
                        </div>)
                        }
                    </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuIndicator className="NavigationMenuIndicator" />
            </NavigationMenuList>
        </NavigationMenu>
        </div>
    )
}