import { Routes, Route } from "react-router-dom";
import HomePage from "@/pages/HomePage";
import { SchedulerPage } from "@/pages/SchedulerPage";
import "@/App.css";
import LoginPage from "@/pages/LoginPage";
import SignUpPage from "@/pages/SignUpPage";
import Layout from "@/components/ui/layout";
import Profile from "@/pages/Profile";
import { ProtectedRoute } from "@/components/ui/ProtectedRoute";
import { PublicRoute } from "@/components/ui/PublicRoute";


function App() {
  return (
    <Routes>
      <Route element={<Layout />}>

        <Route path="/" element={<HomePage />} />
        <Route path="/schedule" element={<ProtectedRoute><SchedulerPage /></ProtectedRoute>} />
        <Route path="/login" element={<PublicRoute><LoginPage /></PublicRoute>} />
        <Route path="/signup" element={<PublicRoute><SignUpPage /></PublicRoute>} />
        <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />

      </Route>
    </Routes>
  );
}

export default App;
