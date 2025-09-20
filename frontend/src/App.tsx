import React from "react";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { ProtectedRoute } from "./components/ProtectedRoute";
import HomePage from "./pages/HomePage";
import LocalAccountPage from "./pages/LocalAuth";
import CloudAuthPage from "./pages/CloudAuthPage";
import { SchedulerPage } from "./pages/SchedulerPage";
import "./App.css";


function App() {
  return (
      <Routes>
        <Route path="/" element={ <HomePage />} />
        <Route path="/schedule" element={<SchedulerPage />} />
        <Route path="/auth" element={<CloudAuthPage />} />
        <Route path="/local" element={<LocalAccountPage />} />
      </Routes>
  );
}

export default App;
