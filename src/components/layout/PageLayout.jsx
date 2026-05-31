import Header from "./Header";

// Shared page shell: constrained width + the profile Header.
// Used by every content page so layout changes live in one place.
export default function PageLayout({ children }) {
  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Header />
      {children}
    </section>
  );
}
