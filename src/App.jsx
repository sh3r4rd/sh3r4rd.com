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
import React, { useState, useEffect } from "react";

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
  const [timeLeft, setTimeLeft] = useState({ days: 0, hours: 0, minutes: 0, seconds: 0 });

  useEffect(() => {
    const targetDate = new Date('2025-08-08T17:30:00-04:00'); // Set target date here: August 8, 2025 at 5:30 PM EDT

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

        <section className="prose max-w-none prose-a:underline">
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
                  When I started my career I worked primarily on the frontend, but I have since transitioned to backend engineering. I'm <a href="https://github.com/sh3r4rd/sh3r4rd.com" target="_blank" rel="noopener noreferrer">working on this portfolio</a> to showcase my skills and projects throughout the full stack.
                  This site is built with React, Tailwind CSS, and Vite. It features a responsive design, dark mode support, and a clean, modern aesthetic. On the backend I'm leveraging AWS services like S3, CloudFront, API Gateway and Lambda.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <h3 className="text-xl font-medium">Leveraging AI</h3>
                <p className="text-gray-700 dark:text-gray-300 mt-2">
                  I'm currently exploring how to leverage AI to enhance my development workflow and improve the user experience of my applications. This includes using AI for code generation, testing, and even automating some aspects of deployment.
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

function SkillsSection() {
  const skills = [
    { 
      name: "Golang", 
      description: `This is my strongest coding language. I write code daily in Golang for backend services. The code I write is idiomatic, performant, and well-tested.
      One of the more interesting issues I've run into recently is needing to set GO_MAX_PROCS to accurately handle CPU resources in a containerized environment.` 
    },
    { 
      name: "Postgres", 
      description: `I've worked with SQL databases extensively, particularly Postgres. I have experience with advanced querying, indexing strategies, and performance 
      optimizations. One pattern I use often in microservices with Postgres is the outbox pattern, which allows me to handle eventual consistency and message delivery guarantees.` 
    },
    { 
      name: "AWS", 
      description: `I've used many of AWS's services, including EC2, Lambda, S3, CloudFront, and RDS to design and implement software solutions. One interesting service that
      I used recently is AWS Device Farm, which allows me to test mobile applications on real devices in the cloud. This is particularly useful for ensuring compatibility 
      across different devices and OS versions.` 
    },
    { 
      name: "Kafka", 
      description: `I've used Kafka for building event-driven architectures at a few jobs. I often use it to decouple services and ensure reliable message delivery. I've used with and
      without a schema registry, depending on the use case.`
    },
    { 
      name: "Docker", 
      description: `Docker has been a crucial part of my developer workflow at every job. I use it to containerize applications, manage dependencies, and ensure consistent environments 
      across development, testing, and production. It's particularly useful for microservices architectures, where each service can run in its own container with its own dependencies.` 
    },
    { 
      name: "GitHub Actions", 
      description: `I've used GitHub Actions for automating CI/CD pipelines, running tests, and deploying applications. It's a powerful tool for ensuring code quality and streamlining the development process.` 
    },
    { 
      name: "New Relic", 
      description: `I've used New Relic for application performance monitoring and observability. It provides valuable insights into application performance, user interactions, and infrastructure health. 
      I've integrated it with various services to track key metrics and troubleshoot issues effectively.` 
    },
    {
      name: "Redis",
      description: `I've used Redis for caching, session management, and real-time data processing. It's an excellent tool for improving application performance and scalability, especially in high-traffic environments. 
      I've implemented Redis in various architectures to handle caching and pub/sub messaging patterns effectively. I currently use Redis to handle caching for complex queries used to determine participant compliance in a study.`
    },
    {
      name: "Agile Methodologies",
      description: `I've worked with Agile methodologies throughout my career, participating in Scrum and Kanban processes. I value iterative development, continuous feedback, and cross-functional collaboration. 
      These principles have helped me deliver high-quality software that meets user needs effectively.`
    },
    {
      name: "API Design",
      description: `I've designed and implemented RESTful APIs and GraphQL endpoints for various applications. I focus on creating intuitive and efficient APIs that meet the needs of both frontend developers and end-users. 
      I also prioritize API documentation and versioning to ensure smooth integration and maintainability. I've used tools like Swagger and Postman to document and test APIs effectively.`
    },
    {
      name: "Microservices Architecture",
      description: `I've designed and implemented microservices architectures for scalable and maintainable applications. I focus on service decomposition, inter-service communication (with experience in gRPC and REST), and data management strategies. 
      I've used technologies like Docker, Kubernetes, and Apache Kafka to manage microservices effectively. I also prioritize observability and monitoring to ensure the health and performance of distributed systems.`
    },
    {
      name: "React",
      description: `I've used React for building dynamic and responsive user interfaces. I am currently more of a backend engineer, but I have experience with React and its ecosystem, including state management with Redux and context API.`
    },
    {
      name: "NoSQL Databases",
      description: `I've worked with NoSQL databases like MongoDB and Neo4j for applications that require flexible data models and horizontal scalability. I have experience designing data schemas, optimizing queries, and integrating these databases with various backend technologies.`
    }
  ];
  const [selected, setSelected] = useState(0);

  return (
    <section className="text-center">
      <h2 className="text-2xl font-bold mb-4">Skills</h2>
      <div className="flex flex-wrap justify-center gap-4 mb-4">
        {skills.map((skill, idx) => (
          <button
            key={idx}
            onClick={() => setSelected(selected === idx ? null : idx)}
            className={`transition-all duration-300 text-white px-4 py-2 rounded-full shadow-md ${selected === idx ? 'text-xl bg-purple-600 scale-110' : 'bg-indigo-500 hover:bg-indigo-600'}`}
          >
            {skill.name}
          </button>
        ))}
      </div>
      {selected !== null && (
        <div className="max-w-xl mx-auto text-gray-100 bg-gray-800 rounded-lg p-4 shadow-inner">
          {skills[selected].description}
        </div>
      )}
    </section>
  );
}

function ResumeRequestPage() {
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e) => {
    e.preventDefault();
    const form = e.target;
    const nameRegex = /^.{2,}$/;
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const phoneRegex = /^(\+\d{1,2}\s?)?1?\-?\.?\s?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$/;
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
      alert("Please enter a valid US phone number");
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
    setSubmitted(true);
  };

  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />

      {!submitted ? (
        <>
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
            <textarea name="description" placeholder="Job Description" required rows={6} className="p-2 border rounded-md text-black resize-y" />
            <Button type="submit">Submit Request</Button>
          </form>
        </>
      ) : (
        <div className="text-lg text-center text-gray-700 dark:text-gray-300">
          <p>Thank you for your consideration — I appreciate your interest.</p>
          <p>I&apos;ve received your request and will review it shortly. You&apos;ll receive a confirmation by email and I&apos;ll be in touch soon.</p>

          <div className="max-w-md mx-auto">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
            <path fill="#FF156D" stroke="#FF156D" stroke-width="15" transform-origin="center" d="m148 84.7 13.8-8-10-17.3-13.8 8a50 50 0 0 0-27.4-15.9v-16h-20v16A50 50 0 0 0 63 67.4l-13.8-8-10 17.3 13.8 8a50 50 0 0 0 0 31.7l-13.8 8 10 17.3 13.8-8a50 50 0 0 0 27.5 15.9v16h20v-16a50 50 0 0 0 27.4-15.9l13.8 8 10-17.3-13.8-8a50 50 0 0 0 0-31.7Zm-47.5 50.8a35 35 0 1 1 0-70 35 35 0 0 1 0 70Z">
              <animateTransform type="rotate" attributeName="transform" calcMode="spline" dur="1.5" values="0;120" keyTimes="0;1" keySplines="0 0 1 1" repeatCount="indefinite"></animateTransform>
            </path>
          </svg>
          </div>
        </div>
      )}
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