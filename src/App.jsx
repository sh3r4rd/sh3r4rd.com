import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import NavMenu from "./components/layout/NavMenu";
import PageTracker from "./components/layout/PageTracker";
import HomePage from "./pages/HomePage";
import ResumeRequestPage from "./pages/ResumeRequestPage";
import DashboardPage from "./pages/DashboardPage";
import NotFoundPage from "./pages/NotFoundPage";

export default function App() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white px-8 pb-8 pt-20">
      <Router>
        <NavMenu />
        <PageTracker />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/resume" element={<ResumeRequestPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Router>
    </main>
  );
}