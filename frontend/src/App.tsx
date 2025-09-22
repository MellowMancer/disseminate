import { Routes, Route } from "react-router-dom";
import HomePage from "./pages/HomePage";
import { SchedulerPage } from "./pages/SchedulerPage";
import "./App.css";
import LoginPage from "./pages/LoginPage";
import SignUpPage from "./pages/SignUpPage";
import Layout from "./components/ui/layout";


function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/schedule" element={<SchedulerPage />} />
        <Route path="auth/login" element={<LoginPage />} />
        <Route path="auth/signup" element={<SignUpPage />} />
      </Route>
    </Routes>
  );
}

export default App;
