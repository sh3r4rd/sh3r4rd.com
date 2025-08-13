import { Link, useLocation } from "react-router-dom";

export default function Breadcrumbs() {
  const location = useLocation();
  const crumbs = location.pathname.split("/").filter(Boolean);

  return (
    <nav className="mb-6 text-sm text-gray-500 dark:text-gray-400">
      <Link to="/" className="hover:underline">Home</Link>
      {crumbs.map((crumb, idx) => (
        <span key={idx}>
          {' / '}
          <span className="capitalize">
            {idx === crumbs.length - 1 ? crumb : <Link to={`/${crumb}`} className="hover:underline">{crumb}</Link>}
          </span>
        </span>
      ))}
    </nav>
  );
}