import { Link } from "react-router-dom";

export default function NotFoundPage() {
  return (
    <section className="text-center mt-20">
      <h1 className="text-4xl font-bold mb-4">404 - Page Not Found</h1>
      <p className="text-lg text-gray-600 dark:text-gray-400 mb-6">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link
        to="/"
        className="text-indigo-600 hover:underline font-medium"
      >
        ← Go back home
      </Link>
    </section>
  );
}