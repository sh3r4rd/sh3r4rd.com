import { Button } from "./components/ui/button";
import { Card, CardContent } from "./components/ui/card";
import { Github, Mail, Linkedin } from "lucide-react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
} from "react-router-dom";
import React from "react";

function Breadcrumbs() {
  const location = useLocation();
  const crumbs = location.pathname.split("/").filter(Boolean);

  return (
    <nav className="mb-6 text-sm text-gray-500 dark:text-gray-400">
      <Link to="/" className="hover:underline">Home</Link>
      {crumbs.map((crumb, idx) => (
        <span key={idx}>
          {' / '}
          <span className="capitalize">
            {idx === crumbs.length - 1 ? crumb : <Link to={`/${crumb}`} className="hover:underline">{crumb}</Link>}
          </span>
        </span>
      ))}
    </nav>
  );
}

function Header() {
  return (
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
          <a href="mailto:dev@sh3r4rd.com" aria-label="Email">
            <Mail />
          </a>
          <a href="https://www.linkedin.com/in/sherardbailey" target="_blank" aria-label="LinkedIn">
            <Linkedin />
          </a>
          <a href="https://github.com/sh3r4rd" target="_blank" aria-label="GitHub">
            <Github />
          </a>
        </div>
      </div>
    </header>
  )
}

function HomePage() {
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
  );
}

function ResumeRequestPage() {
  const handleSubmit = (e) => {
    e.preventDefault();
    const form = e.target;
    const nameRegex = /^.{2,}$/;
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const phoneRegex = /^(\([0-9]{3}\) |[0-9]{3}-)[0-9]{3}-[0-9]{4}$/;
    const jobTitleMinLength = 10;
    const descriptionMinWords = 50;

    const fields = [
      "firstName",
      "lastName",
      "email",
      "phone",
      "company",
      "jobTitle",
      "description",
    ];

    for (let id of fields) {
      const el = form.elements[id];
      if (!el || !el.value.trim()) {
        alert(`Please fill in the ${id} field.`);
        el.focus();
        return;
      }
    }

    if (!nameRegex.test(form.firstName.value)) {
      alert("First name must be at least 2 characters.");
      form.firstName.focus();
      return;
    }

    if (!nameRegex.test(form.lastName.value)) {
      alert("Last name must be at least 2 characters.");
      form.lastName.focus();
      return;
    }

    if (!emailRegex.test(form.email.value)) {
      alert("Please enter a valid email address.");
      form.email.focus();
      return;
    }
    
    if (!phoneRegex.test(form.phone.value)) {
      console.log('phone', form.phone.value);
      alert("Please enter a valid US phone number.");
      form.phone.focus();
      return;
    }

    if (form.jobTitle.value.trim().length < jobTitleMinLength) {
      alert("Job title must be at least 10 characters long.");
      form.jobTitle.focus();
      return;
    }

    const wordCount = form.description.value.trim().split(/\s+/).length;
    if (wordCount < descriptionMinWords) {
      alert("Job description must be at least 50 words.");
      form.description.focus();
      return;
    }

    alert("Form submitted successfully!");
    form.reset();
  };

  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />

      <p className="text-lg text-gray-700 dark:text-gray-300 mb-8">
        Thank you for your interest! Please fill out the form below to request a copy of my résumé.
      </p>
      <form className="grid gap-4" onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <input name="firstName" type="text" placeholder="First Name" required className="p-2 border rounded-md text-black" />
          <input name="lastName" type="text" placeholder="Last Name" required className="p-2 border rounded-md text-black" />
        </div>
        <input name="email" type="email" placeholder="Email" required className="p-2 border rounded-md text-black" />
        <input name="phone" type="tel" placeholder="Phone Number" required className="p-2 border rounded-md text-black" />
        <input name="company" type="text" placeholder="Company" required className="p-2 border rounded-md text-black" />
        <input name="jobTitle" type="text" placeholder="Job Title" required className="p-2 border rounded-md text-black" />
        <textarea name="description" placeholder="Job Description" required rows={6} className="p-2 border rounded-md resize-y text-black" />
        <Button type="submit">Submit Request</Button>
      </form>
    </section>
  );
}

export default function App() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-8">
      <Router>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/resume" element={<ResumeRequestPage />} />
        </Routes>
      </Router>
    </main>
  );
}