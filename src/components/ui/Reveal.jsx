import useReveal from "../../hooks/useReveal";

// Wrapper that fades + lifts its children into view on scroll.
// Under prefers-reduced-motion the global CSS neutralizes the transition,
// so the content still appears. Pass `immediate` for above-the-fold content
// so it renders visible right away instead of waiting for the observer.
export function Reveal({ children, className = "", immediate = false, ...props }) {
  const [ref, visible] = useReveal(immediate);

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
