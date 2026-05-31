import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import NavMenu from "./components/layout/NavMenu";
import PageTracker from "./components/layout/PageTracker";
import HomePage from "./pages/HomePage";
import ResumeRequestPage from "./pages/ResumeRequestPage";
import DashboardPage from "./pages/DashboardPage";
import NotFoundPage from "./pages/NotFoundPage";

export default function App() {
  // pt-20 clears the fixed NavMenu (h-14); keep it >= the nav height if that changes.
  return (
    <main className="relative min-h-screen text-gray-900 dark:text-white px-8 pb-8 pt-20">
      {/* Ambient aurora background behind everything */}
      <div
        aria-hidden
        className="fixed inset-0 -z-10 overflow-hidden bg-white dark:bg-gray-950"
      >
        <div className="absolute -top-40 -left-32 w-[34rem] h-[34rem] rounded-full bg-brand-gradient opacity-[0.07] dark:opacity-20 blur-[130px] animate-aurora" />
        <div className="absolute bottom-0 right-0 w-[30rem] h-[30rem] rounded-full bg-brand-gradient opacity-[0.07] dark:opacity-20 blur-[130px] animate-aurora [animation-delay:-9s]" />
      </div>

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
