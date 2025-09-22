import { Routes, Route } from "react-router-dom";
import HomePage from "./pages/HomePage";
import { SchedulerPage } from "./pages/SchedulerPage";
import "./App.css";
import LoginPage from "./pages/LoginPage";
import SignUpPage from "./pages/SignUpPage";
import Layout from "./components/ui/layout";
// import { ProtectedRoute } from "./components/ui/ProtectedRoute";


function App() {
  return (
    <Routes>
      <Route element={<Layout />}>

        <Route
          path="/"
          element={
              <HomePage />
          }
        />
        <Route
          path="/schedule"
          element={
              <SchedulerPage />
          }
        />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignUpPage />} />

      </Route>
    </Routes>
  );
}

export default App;
