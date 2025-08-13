import { Github, Linkedin } from "lucide-react";

export default function Header() {
  return (
    <header className="flex flex-col md:flex-row items-center justify-center md:items-start gap-6 text-center md:text-left mt-8 mb-16">
      <img
        src="https://d241eitbp7g6mq.cloudfront.net/images/sherard-profile.jpg"
        alt="Sherard Bailey"
        className="w-32 h-32 rounded-full border-4 border-white shadow-md object-cover"
      />

      <div>
        <h1 className="text-4xl font-bold">Sherard Bailey</h1>
        <p className="text-lg mt-2 text-gray-600 dark:text-gray-400">
          Senior Software Engineer —  Microservices | Event-driven Architecture | AWS | Kafka
        </p>
        <div className="flex justify-center md:justify-start gap-4 mt-4">
          <a href="https://www.linkedin.com/in/sherardbailey" target="_blank" aria-label="LinkedIn">
            <Linkedin />
          </a>
          <a href="https://github.com/sh3r4rd" target="_blank" aria-label="GitHub">
            <Github />
          </a>
        </div>
      </div>
    </header>
  );
}