import { useEffect, useState } from "react";
import { NavLink } from "react-router-dom";

const links = [
  { to: "/", label: "Home", end: true },
  { to: "/resume", label: "Resume" },
];

export default function NavMenu() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <nav
      className={`fixed top-0 left-0 right-0 z-50 transition-colors duration-300 ${
        scrolled
          ? "bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border-b border-white/40 dark:border-white/10 shadow-sm"
          : "bg-transparent border-b border-transparent"
      }`}
    >
      <div className="max-w-4xl mx-auto flex items-center gap-6 px-8 h-14">
        {links.map(({ to, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `relative text-sm font-medium transition-colors hover:text-indigo-600 dark:hover:text-indigo-400 ${
                isActive
                  ? "text-indigo-600 dark:text-indigo-400 after:absolute after:-bottom-1 after:left-0 after:right-0 after:h-0.5 after:rounded-full after:bg-brand-gradient"
                  : "text-gray-600 dark:text-gray-300"
              }`
            }
          >
            {label}
          </NavLink>
        ))}
      </div>
    </nav>
  );
}
