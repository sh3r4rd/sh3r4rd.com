import { useEffect, useState } from "react";
import { NavLink } from "react-router-dom";

const links = [
  { to: "/", label: "Home", end: true },
  { to: "/resume", label: "Resume" },
  // `rel: nofollow` pairs with public/robots.txt and the noindex meta tag in
  // DashboardPage — this is the only in-site link pointing at /dashboard.
  { to: "/dashboard", label: "Dashboard", badge: "New", rel: "nofollow" },
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
        {links.map(({ to, label, end, badge, rel }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            rel={rel}
            className={({ isActive }) =>
              `relative text-sm font-medium transition-colors hover:text-teal-600 dark:hover:text-teal-400 ${
                isActive
                  ? "text-gray-900 dark:text-white after:absolute after:-bottom-1 after:left-0 after:right-0 after:h-0.5 after:rounded-full after:bg-brand-gradient"
                  : "text-gray-600 dark:text-gray-300"
              }`
            }
          >
            {label}
            {/* Superscripted via `relative -top-*` rather than `align-super`:
                it lifts the pill visually without growing the line box, which
                would otherwise throw off the h-14 bar's vertical centering.
                `aria-hidden` keeps the link's accessible name "Dashboard" —
                without it a screen reader announces "Dashboard New, link". */}
            {badge && (
              <span
                aria-hidden
                className="relative -top-1.5 ml-1 inline-flex items-center rounded-full bg-brand-gradient px-1.5 py-px text-[9px] font-semibold uppercase tracking-wider leading-none text-white"
              >
                {badge}
              </span>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  );
}
