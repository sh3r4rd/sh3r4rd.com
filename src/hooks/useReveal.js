import { useEffect, useRef, useState } from "react";

// Reveal-on-scroll hook. Returns [ref, visible].
// Defaults to hidden, but if IntersectionObserver is unsupported it falls back
// to visible so content is never permanently hidden.
export default function useReveal(immediate = false) {
  const ref = useRef(null);
  const [visible, setVisible] = useState(immediate);

  useEffect(() => {
    if (immediate) return;
    if (typeof IntersectionObserver === "undefined") {
      setVisible(true);
      return;
    }

    const node = ref.current;
    if (!node) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { rootMargin: "-80px" }
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [immediate]);

  return [ref, visible];
}
