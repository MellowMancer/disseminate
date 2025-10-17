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
    )
}