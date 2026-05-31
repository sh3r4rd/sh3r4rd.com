import React from "react";

export function Card({ children, className = "" }) {
  return (
    <div
      className={`group relative rounded-2xl border border-white/40 dark:border-white/10 bg-white/70 dark:bg-gray-900/50 backdrop-blur-xl shadow-md transition-all duration-300 hover:-translate-y-1 hover:shadow-brand-glow ${className}`}
    >
      <span
        aria-hidden
        className="absolute inset-x-0 top-0 h-1 rounded-t-2xl bg-brand-gradient opacity-0 group-hover:opacity-100 transition-opacity"
      />
      {children}
    </div>
  );
}

export function CardContent({ children, className = "" }) {
  return (
    <div className={`p-4 ${className}`}>
      {children}
    </div>
  );
}
