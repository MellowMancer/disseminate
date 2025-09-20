import React from "react";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { ProtectedRoute } from "./components/ProtectedRoute";
import HomePage from "./pages/HomePage";
import { SchedulerPage } from "./pages/SchedulerPage";
import "./App.css";


function App() {
  return (
      <Routes>
        <Route path="/" element={ <HomePage />} />
        <Route path="/schedule" element={<SchedulerPage />} />
      </Routes>
  );
}

export default App;
