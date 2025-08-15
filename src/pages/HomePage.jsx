import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";
import SkillsSection from "../components/sections/SkillsSection";

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
      <section className="max-w-4xl mx-auto space-y-12">
        <Breadcrumbs />
        <Header /> 
        <section>
          <h2 className="text-2xl font-semibold mb-4">About Me</h2>
          <p className="text-gray-700 dark:text-gray-300">
            I'm a senior software engineer with extensive experience in event-driven microservices, software architecture, cloud services and distributed systems. I've worked on large-scale backend systems, optimized complex queries, and built distributed applications that scale.
            I'm seeking opportunities to solve challenging problems and build high-impact systems.
          </p>
        </section>

        <section className="text-center mt-12">
          <Link to="/resume">
            <Button size="lg">Request Résumé</Button>
          </Link>
        </section>

        <section className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white p-6 rounded-2xl shadow-lg text-center mt-8">
          <h2 className="text-2xl font-bold mb-2">🚀 Coming Soon</h2>
          <p className="text-lg">
            Let me save you time! I'm building a new feature with the help of AI that will automatically analyze your job descriptions and let you know if I'm a good fit — instantly.
            This tool is currently in development and will be launching soon.
          </p>
          <div className="text-xl font-mono mt-2">
            {timeLeft.days}d {timeLeft.hours}h {timeLeft.minutes}m {timeLeft.seconds}s
          </div>
        </section>

        <section className="max-w-none prose-a:underline">
          <h2 className="text-2xl font-semibold mb-4">Projects</h2>
          <div className="grid gap-6">
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">AWS Solutions Architect</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  A couple years ago, I passed the <a href="https://aws.amazon.com/certification/certified-solutions-architect-associate/" target="_blank" rel="noopener noreferrer">AWS Solutions Architect - Associate</a> exam. I have experience with many AWS services, including EC2, Lambda, S3, CloudFront, and RDS. But this
                  certification is a testament to my understanding of AWS architecture and best practices. I am currently working on the Professional level certification to deepen my expertise.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Building Portfolio</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  When I started my career I worked primarily on the frontend but I have since transitioned to backend engineering. I'm <a href="https://github.com/sh3r4rd/sh3r4rd.com" target="_blank" rel="noopener noreferrer">working on this portfolio</a> to showcase my skills and projects throughout the full stack.
                  This site is built with React, Tailwind CSS, and Vite. It features a responsive design, dark mode support, and a clean, modern aesthetic. On the backend I'm leveraging AWS services like Route 53, S3, CloudFront, SES, API Gateway and Lambda.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Leveraging AI</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  I'm currently exploring how to leverage AI to enhance my development workflow and improve the user experience of my applications. This includes using AI for code generation, testing, and even automating some aspects of deployment via <span className="font-semibold dark:text-white">Claude Code</span> and <span className="font-semibold dark:text-white">Github Copilot</span>.
                  I'm particularly interested in how AI can help with code reviews, bug detection, and optimizing performance. I believe AI has the potential to revolutionize software development, and I'm excited to be at the forefront of this change.
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        <SkillsSection />
      </section>
  );
}