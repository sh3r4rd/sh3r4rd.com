import { Button } from "./components/ui/button";
import { Card, CardContent } from "./components/ui/card";
import { Github, Mail, Linkedin } from "lucide-react";

export default function Portfolio() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-8">
      <section className="max-w-4xl mx-auto space-y-12">
        <header className="flex flex-col md:flex-row items-center justify-center md:items-start gap-6 text-center md:text-left mt-8 mb-16">
          {/* Profile Image */}
          <img
            src="https://d241eitbp7g6mq.cloudfront.net/images/sherard-profile.jpg"  // Make sure to put your JPG in the /public folder as profile.jpg
            alt="Sherard Bailey"
            className="w-32 h-32 rounded-full border-4 border-white shadow-md object-cover"
          />

          {/* Text and Links */}
          <div>
            <h1 className="text-4xl font-bold">Sherard Bailey</h1>
            <p className="text-lg mt-2 text-gray-600 dark:text-gray-400">
              Senior Software Engineer —  Microservices | Event-driven Architecture | AWS | Kafka
            </p>
            <div className="flex justify-center md:justify-start gap-4 mt-4">
              <a href="mailto:sherard@example.com" aria-label="Email">
                <Mail />
              </a>
              <a href="https://www.linkedin.com/in/sherardbailey" target="_blank" aria-label="LinkedIn">
                <Linkedin />
              </a>
              <a href="https://github.com/sherardbailey" target="_blank" aria-label="GitHub">
                <Github />
              </a>
            </div>
          </div>
        </header>

        <section>
          <h2 className="text-2xl font-semibold mb-4">About Me</h2>
          <p className="text-gray-700 dark:text-gray-300">
            I'm a senior software engineer with extensive experience in event-driven microservices, software architecture, cloud services and distributed systems. I've worked on large-scale backend systems, optimized complex queries, and built distributed applications that scale.
            I'm seeking opportunities to solve challenging problems and build high-impact systems.
          </p>
        </section>

        <section className="text-center mt-12">
          <Button size="lg">Request Résumé</Button>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Projects</h2>
          <div className="grid gap-6">
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">In-Memory Key-Value Store</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Built a concurrency-safe key-value store in Go with TTL support and background cleanup using goroutines and mutexes.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Kafka Event Consumer Observability</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Designed a Kafka-based observability solution integrated with New Relic to monitor throughput, error rates and latency in consumers.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Sharded Postgres Architecture</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  Implemented a scalable Postgres sharded architecture with global indexing and optimized query routing.
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
            <li>Distributed Systems</li>
            <li>Observability (New Relic)</li>
            <li>Concurrency</li>
            <li>API Design</li>
            <li>Performance Optimization</li>
            <li>Agile Methodologies</li>
          </ul>
        </section>

        
      </section>
    </main>
  );
}
