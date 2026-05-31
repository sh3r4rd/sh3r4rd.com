import { useState, useEffect } from "react";
import { Cloud, Code2, Sparkles } from "lucide-react";
import { Card, CardContent } from "../components/ui/card";
import { Reveal } from "../components/ui/Reveal";
import PageLayout from "../components/layout/PageLayout";
import SkillsSection from "../components/sections/SkillsSection";

const projects = [
  {
    icon: Cloud,
    title: "AWS Solutions Architect",
    summary: "AWS Solutions Architect – Associate certified, with hands-on experience across core AWS services.",
    tags: ["EC2", "Lambda", "S3", "CloudFront", "RDS"],
    body: (
      <>
        A couple years ago, I passed the <a href="https://aws.amazon.com/certification/certified-solutions-architect-associate/" target="_blank" rel="noopener noreferrer">AWS Solutions Architect - Associate</a> exam. I have experience with many AWS services, including EC2, Lambda, S3, CloudFront, and RDS. But this
        certification is a testament to my understanding of AWS architecture and best practices. I am currently working on the Professional level certification to deepen my expertise.
      </>
    ),
  },
  {
    icon: Code2,
    title: "Building Portfolio",
    summary: "A full-stack portfolio built with React and Tailwind on the frontend and AWS serverless on the backend.",
    tags: ["React", "Tailwind", "Vite", "AWS"],
    body: (
      <>
        When I started my career I worked primarily on the frontend but I have since transitioned to backend engineering. I'm <a href="https://github.com/sh3r4rd/sh3r4rd.com" target="_blank" rel="noopener noreferrer">working on this portfolio</a> to showcase my skills and projects throughout the full stack.
        This site is built with React, Tailwind CSS, and Vite. It features a responsive design, dark mode support, and a clean, modern aesthetic. On the backend I'm leveraging AWS services like Route 53, S3, CloudFront, SES, API Gateway and Lambda.
      </>
    ),
  },
  {
    icon: Sparkles,
    title: "Leveraging AI",
    summary: "Using AI to accelerate development — from code generation to reviews, testing, and deployment.",
    tags: ["Claude Code", "GitHub Copilot"],
    body: (
      <>
        I'm currently exploring how to leverage AI to enhance my development workflow and improve the user experience of my applications. This includes using AI for code generation, testing, and even automating some aspects of deployment via <span className="font-semibold dark:text-white">Claude Code</span> and <span className="font-semibold dark:text-white">Github Copilot</span>.
        I'm particularly interested in how AI can help with code reviews, bug detection, and optimizing performance. I believe AI has the potential to revolutionize software development, and I'm excited to be at the forefront of this change.
      </>
    ),
  },
];

const countdownLabels = [
  ["days", "Days"],
  ["hours", "Hrs"],
  ["minutes", "Min"],
  ["seconds", "Sec"],
];

export default function HomePage() {
  const [timeLeft, setTimeLeft] = useState({ days: 0, hours: 0, minutes: 0, seconds: 0 });

  useEffect(() => {
    let targetDate;
    if (typeof window !== "undefined" && window.localStorage) {
      targetDate = new Date(localStorage.getItem("targetDate") || 0);
    } else {
      targetDate = new Date(Date.now() + 21 * 24 * 60 * 60 * 1000);
    }
    const now = new Date();
    // If no date is set, or less than 5 days away, set to 3 weeks from now
    if (!targetDate || isNaN(targetDate) || (targetDate - now) < 5 * 24 * 60 * 60 * 1000) {
      targetDate = new Date(now.getTime() + 21 * 24 * 60 * 60 * 1000);
      localStorage.setItem("targetDate", targetDate.toISOString());
    }

    const interval = setInterval(() => {
      const now = new Date();
      const difference = targetDate - now;

      const days = Math.floor(difference / (1000 * 60 * 60 * 24));
      const hours = Math.floor((difference / (1000 * 60 * 60)) % 24);
      const minutes = Math.floor((difference / 1000 / 60) % 60);
      const seconds = Math.floor((difference / 1000) % 60);

      setTimeLeft({ days, hours, minutes, seconds });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  return (
    <PageLayout>
      <Reveal>
        <section>
          <h2 className="text-2xl font-semibold mb-4">About Me</h2>
          <p className="text-gray-700 dark:text-gray-300">
            I'm a software engineer with extensive experience in event-driven microservices, software architecture, cloud services and distributed systems. I've worked on large-scale backend systems, optimized complex queries, and built distributed applications that scale.
            I'm seeking opportunities to solve challenging problems and build high-impact systems.
          </p>
        </section>
      </Reveal>

      <Reveal>
        <section className="relative overflow-hidden rounded-3xl p-8 text-white bg-brand-gradient bg-[length:200%_200%] animate-gradient-pan shadow-brand-glow text-center">
          <span className="inline-flex items-center gap-2 rounded-full bg-white/15 backdrop-blur px-3 py-1 text-sm font-semibold">
            <Sparkles className="w-4 h-4" />
            AI-Powered · Launching soon
          </span>
          <h2 className="text-2xl font-bold mt-3 mb-2">Coming Soon</h2>
          <p className="text-lg">
            Let me save you time! I'm building a new feature with the help of AI that will automatically analyze your job descriptions and let you know if I'm a good fit — instantly.
            This tool is currently in development and will be launching soon.
          </p>

          <span className="sr-only">Launching soon</span>
          <div
            aria-live="off"
            className="mt-6 grid grid-cols-4 gap-3 max-w-md mx-auto"
          >
            {countdownLabels.map(([key, label]) => (
              <div
                key={key}
                className="rounded-2xl bg-white/15 backdrop-blur border border-white/20 py-3"
              >
                <div className="text-3xl font-bold tabular-nums">
                  {String(timeLeft[key]).padStart(2, "0")}
                </div>
                <div className="text-xs font-semibold uppercase tracking-wide text-white/80">
                  {label}
                </div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

      <Reveal>
        <section className="max-w-none prose-a:underline">
          <h2 className="text-2xl font-semibold mb-4">Projects</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <Card key={project.title}>
                <CardContent className="p-4 flex flex-col h-full">
                  <span className="inline-flex items-center justify-center w-11 h-11 rounded-xl text-white bg-brand-gradient shadow-brand-glow">
                    <project.icon className="w-5 h-5" />
                  </span>
                  <h3 className="text-xl font-medium mt-3">{project.title}</h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">{project.summary}</p>

                  <div className="flex flex-wrap gap-1.5 mt-3">
                    {project.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded-full px-2 py-0.5 text-xs font-medium text-purple-700 dark:text-purple-200 bg-purple-50 dark:bg-white/5 border border-purple-200/60 dark:border-white/10"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>

                  <details className="mt-4 group/details">
                    <summary className="cursor-pointer text-sm font-semibold text-purple-700 dark:text-purple-300 list-none [&::-webkit-details-marker]:hidden">
                      Read more
                    </summary>
                    <p className="text-gray-700 dark:text-gray-300 mt-2 text-sm">
                      {project.body}
                    </p>
                  </details>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      </Reveal>

      <Reveal>
        <SkillsSection />
      </Reveal>
    </PageLayout>
  );
}
