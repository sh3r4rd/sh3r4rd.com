import { Button } from "./components/ui/button";
import { Card, CardContent } from "./components/ui/card";
import { Github, Mail, Linkedin } from "lucide-react";

export default function Portfolio() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-8">
      <section className="max-w-4xl mx-auto space-y-12">
        <header className="text-center">
          <h1 className="text-4xl font-bold">Sherard Bailey</h1>
          <p className="text-lg mt-2 text-gray-600 dark:text-gray-400">
            Senior Software Engineer — Microservices | Event-driven architecture | AWS | Kafka
          </p>
          <div className="flex justify-center gap-4 mt-4">
            <a href="mailto:sherard@example.com" aria-label="Email">
              <Mail />
            </a>
            <a href="https://www.linkedin.com/in/sherardbailey" target="_blank" aria-label="LinkedIn">
              <Linkedin />
            </a>
            <a href="https://github.com/sh3r4rd" target="_blank" aria-label="GitHub">
              <Github />
            </a>
          </div>
        </header>

        <section>
          <h2 className="text-2xl font-semibold mb-4">About Me</h2>
          <p className="text-gray-700 dark:text-gray-300">
            I'm a senior software engineer with extensive experience in Golang, PostgreSQL, system architecture, and event-driven microservices. I've worked on large-scale backend systems, optimized complex queries, and built distributed applications that scale.
            I'm currently seeking opportunities to solve challenging problems and build high-impact systems.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Projects</h2>
          <div className="grid gap-6">
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">In-Memory Key-Value Store (Golang)</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Built a concurrency-safe key-value store in Go with TTL support and background cleanup using goroutines and mutexes.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Kafka Microservice Observability</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Designed a Kafka-based observability solution integrated with New Relic to monitor throughput and version message schemas.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Sharded SQL Architecture</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Implemented a scalable Postgres architecture with sharding, global indexing, and optimized query routing.
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Skills</h2>
          <ul className="grid grid-cols-2 gap-2 text-gray-700 dark:text-gray-300">
            <li>Golang</li>
            <li>PostgreSQL</li>
            <li>Kafka</li>
            <li>Microservices</li>
            <li>AWS</li>
            <li>System Design</li>
            <li>CI/CD (GitHub Actions)</li>
            <li>Event-driven Architecture</li>
          </ul>
        </section>

        <section className="text-center mt-12">
          <Button size="lg">Request Resume</Button>
        </section>
      </section>
    </main>
  );
}
