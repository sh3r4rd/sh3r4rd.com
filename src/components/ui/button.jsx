import React from "react";

export function Button({ children, size = "md", variant = "primary", className = "", ...props }) {
  const base =
    "inline-flex items-center justify-center gap-2 rounded-2xl font-semibold transition-all duration-200 active:translate-y-0";

  const sizes = {
    sm: "px-3 py-1.5 text-sm",
    md: "px-5 py-2.5 text-base",
    lg: "px-7 py-3.5 text-lg",
  };

  const variants = {
    primary:
      "text-white bg-brand-gradient bg-[length:200%_200%] animate-gradient-pan shadow-brand-glow hover:-translate-y-0.5",
    glass:
      "text-teal-700 dark:text-teal-200 bg-white/60 dark:bg-white/10 backdrop-blur border border-teal-200/60 dark:border-white/10 hover:bg-white/80",
    ghost:
      "text-teal-700 dark:text-teal-300 hover:bg-teal-50 dark:hover:bg-white/5",
  };

  return (
    <button
      className={`${base} ${sizes[size]} ${variants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
