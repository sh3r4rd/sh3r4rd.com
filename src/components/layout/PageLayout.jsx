import Header from "./Header";

// Shared page shell: constrained width + the profile Header.
// Used by every content page so layout changes live in one place.
// variant "hero" renders the full-bleed hero header; "slim" renders a compact one.
export default function PageLayout({ children, variant = "hero" }) {
  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Header variant={variant} />
      {children}
    </section>
  );
}
