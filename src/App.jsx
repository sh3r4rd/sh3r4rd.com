import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import PageTracker from "./components/layout/PageTracker";
import HomePage from "./pages/HomePage";
import ResumeRequestPage from "./pages/ResumeRequestPage";
import NotFoundPage from "./pages/NotFoundPage";

export default function App() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-8">
      <Router>
        <PageTracker />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/resume" element={<ResumeRequestPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Router>
    </main>
  );
}