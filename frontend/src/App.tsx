import { Routes, Route } from "react-router-dom";
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
