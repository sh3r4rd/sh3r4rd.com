import { Link } from "react-router-dom";
import { Github, Linkedin, ArrowRight } from "lucide-react";

const PROFILE_IMG = "https://d241eitbp7g6mq.cloudfront.net/images/sherard-profile.jpg";

const specialties = ["Microservices", "System Design", "Event-Driven Architecture", "Agentic AI"];

const proofStats = [
  ["8+", "Years shipping backends"],
  ["AWS", "SA-Associate certified"],
  ["Lead", "Engineer & mentor"],
];

function SocialLinks({ className = "" }) {
  return (
    <div className={`flex gap-3 ${className}`}>
      <a
        href="https://www.linkedin.com/in/sherardbailey"
        target="_blank"
        rel="noopener noreferrer"
        aria-label="LinkedIn"
        className="inline-flex items-center justify-center w-11 h-11 rounded-full text-gray-600 dark:text-gray-300 bg-white/70 dark:bg-white/10 border border-gray-200 dark:border-white/10 transition hover:-translate-y-0.5 hover:text-purple-600 dark:hover:text-purple-300 hover:border-purple-300"
      >
        <Linkedin className="w-5 h-5" />
      </a>
      <a
        href="https://github.com/sh3r4rd"
        target="_blank"
        rel="noopener noreferrer"
        aria-label="GitHub"
        className="inline-flex items-center justify-center w-11 h-11 rounded-full text-gray-600 dark:text-gray-300 bg-white/70 dark:bg-white/10 border border-gray-200 dark:border-white/10 transition hover:-translate-y-0.5 hover:text-purple-600 dark:hover:text-purple-300 hover:border-purple-300"
      >
        <Github className="w-5 h-5" />
      </a>
    </div>
  );
}

function SlimHeader() {
  return (
    <header className="flex items-center gap-4 mt-2 mb-6">
      <img
        src={PROFILE_IMG}
        alt="Sherard Bailey"
        className="w-14 h-14 rounded-full ring-2 ring-white dark:ring-gray-900 shadow object-cover"
      />
      <div className="flex-1">
        <h1 className="text-2xl font-extrabold tracking-tight">Sherard Bailey</h1>
        <p className="text-sm font-semibold text-purple-600 dark:text-purple-300">
          Lead Software Engineer
        </p>
      </div>
      <SocialLinks />
    </header>
  );
}

function HeroHeader() {
  return (
    <header className="relative pt-6 pb-12">
      <div className="relative flex flex-col md:flex-row items-center md:items-start gap-8 md:gap-12 text-center md:text-left">
        {/* Avatar with subtle gradient ring */}
        <div className="relative w-36 h-36 shrink-0">
          <span
            aria-hidden
            className="absolute -inset-0.5 rounded-full bg-brand-gradient opacity-70"
          />
          <img
            src={PROFILE_IMG}
            alt="Sherard Bailey"
            className="relative w-36 h-36 rounded-full ring-4 ring-white dark:ring-gray-900 object-cover"
          />
        </div>

        <div className="flex-1">
          <p className="text-xs font-semibold uppercase tracking-widest text-purple-600 dark:text-purple-300">
            Lead Software Engineer
          </p>
          <h1 className="mt-2 text-5xl md:text-6xl font-extrabold tracking-tight text-gray-900 dark:text-white">
            Sherard Bailey
          </h1>
          <p className="mt-3 text-lg text-gray-600 dark:text-gray-300 max-w-2xl">
            I build event-driven, distributed systems composed of microservices that scale — with a growing focus on agentic AI in the software development lifecycle.
          </p>

          {/* Specialty chips */}
          <div className="mt-4 flex flex-wrap justify-center md:justify-start gap-2">
            {specialties.map((tag) => (
              <span
                key={tag}
                className="rounded-full px-3 py-1 text-sm font-medium text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-white/10 border border-gray-200 dark:border-white/10"
              >
                {tag}
              </span>
            ))}
          </div>

          {/* CTA cluster */}
          <div className="mt-6 flex flex-wrap items-center justify-center md:justify-start gap-4">
            <Link
              to="/resume"
              className="group inline-flex items-center justify-center gap-2 rounded-2xl px-7 py-3.5 text-lg font-semibold text-white bg-brand-gradient bg-[length:200%_200%] animate-gradient-pan shadow-brand-glow transition-all duration-200 hover:-translate-y-0.5"
            >
              Request Resume
              <ArrowRight className="w-5 h-5 transition-transform group-hover:translate-x-1" />
            </Link>
            <SocialLinks />
          </div>
        </div>
      </div>

      {/* Seniority proof bar */}
      <dl className="relative mt-10 grid grid-cols-3 gap-3 sm:gap-4">
        {proofStats.map(([value, label]) => (
          <div
            key={label}
            className="rounded-2xl px-4 py-4 text-center bg-white/70 dark:bg-gray-900/50 backdrop-blur-xl border border-gray-200 dark:border-white/10"
          >
            <dt className="sr-only">{label}</dt>
            <dd>
              <span className="block text-2xl sm:text-3xl font-extrabold text-gray-900 dark:text-white">
                {value}
              </span>
              <span className="mt-1 block text-xs sm:text-sm text-gray-600 dark:text-gray-400">
                {label}
              </span>
            </dd>
          </div>
        ))}
      </dl>
    </header>
  );
}

export default function Header({ variant = "hero" }) {
  return variant === "slim" ? <SlimHeader /> : <HeroHeader />;
}
