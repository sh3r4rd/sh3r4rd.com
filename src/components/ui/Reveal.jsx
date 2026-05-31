import useReveal from "../../hooks/useReveal";

// Wrapper that fades + lifts its children into view on scroll.
// Under prefers-reduced-motion the global CSS neutralizes the transition,
// so the content still appears.
export function Reveal({ children, className = "", ...props }) {
  const [ref, visible] = useReveal();

  return (
    <div
      ref={ref}
      className={`transition-all duration-500 ease-out ${
        visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"
      } ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}
