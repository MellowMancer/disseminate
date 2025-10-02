import React from "react";
import { Outlet } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { Navbar } from "./navbar";

const Layout: React.FC = () => {
  const { authenticated } = useAuth();

  return (
    <div className="h-screen w-screen overflow-auto bg-background pt-28 py-16 px-8 md:py-8 md:px-24 mx-auto ">
      <Navbar />
      <Outlet context={{ authenticated }} />

    </div>
  );
};

export default Layout;
