import React from "react";
import { Outlet } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { Navbar } from "./navbar";
import DotMatrixBackground from "./dot-matrix-bg";

const Layout: React.FC = () => {
  const { authenticated } = useAuth();

  return (
    <div className="h-screen w-screen overflow-auto pt-28 py-16 px-8 md:py-8 md:px-24 mx-auto ">
      <DotMatrixBackground />
      <Navbar/>
      <Outlet context={{ authenticated }} />

    </div>
  );
};

export default Layout;
