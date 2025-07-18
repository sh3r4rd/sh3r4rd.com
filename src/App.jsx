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

function PageTracker() {
  const location = useLocation();

  useEffect(() => {
    if (window.gtag) {
      window.gtag("event", "page_view", {
        page_path: location.pathname,
        page_title: document.title,
      });
    }
  }, [location]);

  return null;
}

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
                  I'm currently exploring how to leverage AI to enhance my development workflow and improve the user experience of my applications. This includes using AI for code generation, testing, and even automating some aspects of deployment via <span className="font-semibold dark:text-white">Github Copilot + ChatGPT</span>.
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
      description: (<>This is my most proficient coding language. I write code daily in <span className="font-semibold dark:text-white">Golang</span> for backend services. The code I write is idiomatic, performant, and well-tested.
      One of the more interesting issues I've run into recently is needing to set <span className="inline-block bg-blue-100 text-blue-800 text-xs font-semibold px-1.5 py-0.5 rounded">GO_MAX_PROCS</span> to accurately handle CPU resources in a containerized environment.</>) 
    },
    { 
      name: "Postgres", 
      description: (<>I've worked with SQL databases extensively. I have experience with advanced querying, indexing strategies, and performance optimizations. One pattern I use often in microservices with <span className="font-semibold dark:text-white">Postgres</span> is the outbox pattern, which allows me to handle eventual consistency and message delivery guarantees.</>)
    },
    { 
      name: "AWS", 
      description: (<>I've used many of <span className="font-semibold dark:text-white">AWS</span>'s services, including EC2, Lambda, S3, CloudFront, Route 53 and RDS to design and implement software solutions. One interesting service that
      I used recently is AWS Device Farm, which allows me to test mobile applications on real devices in the cloud. This is particularly useful for ensuring compatibility 
      across different devices and OS versions.</>)
    },
    { 
      name: "Kafka", 
      description: (<>I've used <span className="font-semibold dark:text-white">Kafka</span> for building event-driven architectures at a few jobs. I often use it to decouple services and ensure reliable message delivery. I've used it with and
      without a schema registry. I've also used sink and source connectors for streaming database updates from <span className="font-semibold dark:text-white">Mongo</span> and <span className="font-semibold dark:text-white">Postgres</span> into data pipelines.</>)
    },
    { 
      name: "Docker", 
      description: (<><span className="font-semibold dark:text-white">Docker</span> has been a crucial part of my developer workflow at every job. I use it to containerize applications, manage dependencies, and ensure consistent environments 
      across development, testing, and production. It's particularly useful for microservices architectures, where each service can run in its own container with its own dependencies.</>)
    },
    { 
      name: "GitHub Actions", 
      description: (<>I've used <span className="font-semibold dark:text-white">GitHub Actions</span> for automating CI/CD pipelines, running tests, and deploying applications. It's a powerful tool for ensuring code quality and streamlining the development process.</>)
    },
    { 
      name: "New Relic", 
      description: (<>I've used <span className="font-semibold dark:text-white">New Relic</span> for application performance monitoring and observability. It provides valuable insights into application performance, user interactions, and infrastructure health. 
      I've integrated it with various services to track key metrics and troubleshoot issues effectively. One of my favorite features of <span className="font-semibold dark:text-white">New Relic</span> is the ability to set up custom dashboards and alerts based on SLAs/SLOs, 
      which helps me stay on top of application performance and quickly identify bottlenecks.</>)
    },
    {
      name: "Redis",
      description: (<>I've used <span className="font-semibold dark:text-white">Redis</span> for caching, session management, and real-time data processing. It's an excellent tool for improving application performance and scalability, especially in high-traffic environments. 
      I've used it in various architectures to handle caching and pub/sub messaging patterns effectively. I currently use <span className="font-semibold dark:text-white">Redis</span> to handle caching for complex queries used to determine participant compliance in a study.</>)
    },
    {
      name: "Agile Methodologies",
      description: (<>I've worked with <span className="font-semibold dark:text-white">Agile methodologies</span> throughout my career, participating in <span className="font-semibold dark:text-white">Scrum</span> and <span className="font-semibold dark:text-white">Kanban</span> processes. I value iterative development, continuous feedback, and cross-functional collaboration. 
      These principles have helped me deliver high-quality software that meets user needs effectively. I really enjoy having regular retrospectives to reflect on the team's progress and identify areas for improvement but also to celebrate successes 🎉</>)
    },
    {
      name: "API Design",
      description: (<>I've designed and implemented <span className="font-semibold dark:text-white">RESTful APIs</span> and <span className="font-semibold dark:text-white">GraphQL</span> endpoints for various applications. I focus on creating intuitive and efficient APIs that meet the needs of both frontend developers and end-users. 
      I also prioritize API documentation and versioning to ensure smooth integration and maintainability. I've used tools like <span className="font-semibold dark:text-white">Swagger</span> and <span className="font-semibold dark:text-white">Postman</span> to document and test APIs effectively.</>)
    },
    {
      name: "Microservice Architecture",
      description: (<>I've designed and implemented <span className="font-semibold dark:text-white">microservice architectures</span> for scalable and maintainable applications. I focus on service decomposition, inter-service communication (with experience in <span className="font-semibold dark:text-white">gRPC</span> and <span className="font-semibold dark:text-white">REST</span>), and data management strategies. 
      I've used tools like <span className="font-semibold dark:text-white">Docker</span> and <span className="font-semibold dark:text-white">Kubernetes</span> to manage microservices effectively. I also prioritize observability and monitoring to ensure the health and performance of distributed systems.</>)
    },
    {
      name: "React",
      description: (<>I've used <span className="font-semibold dark:text-white">React</span> for building dynamic and responsive user interfaces. This webpage is built using <span className="font-semibold dark:text-white">React</span>. I am currently work primarily on the backend but I have experience with <span className="font-semibold dark:text-white">React</span> and its ecosystem, including state management with <span className="font-semibold dark:text-white">Redux</span> and <span className="font-semibold dark:text-white">context API</span>.</>)
    },
    {
      name: "NoSQL Databases",
      description: (<>I've worked with a couple <span className="font-semibold dark:text-white">NoSQL databases</span>. I've used <span className="font-semibold dark:text-white">MongoDB</span> to store participant compliance data and audit records because those services didn't need transactions or complex queries but rather high availability. I've used <span className="font-semibold dark:text-white">Neo4j</span> to model relationships 
      between organizations, sponsors, employees and clients as well as cases and referrals that move from one organization to another. These are applications that require flexible data models and horizontal scalability. I have experience designing data schemas, optimizing queries, and integrating these databases with various backend technologies.</>)
    },
    {
      name: "Testing and Quality Assurance",
      description: (<>I prioritize testing and quality assurance in my development process. I have experience with unit testing, integration testing, and end-to-end testing using tools like <span className="font-semibold dark:text-white">Jest</span>, <span className="font-semibold dark:text-white">Mocha</span>, <span className="font-semibold dark:text-white">Cucumber</span> and <span className="font-semibold dark:text-white">RSpec</span> as well as Go's testing and benchmarking package. 
      I believe that thorough testing is essential for delivering reliable software and maintaining code quality.</>)
    },
    {
      name: "TypeScript",
      description: (<>Currently, I use <span className="font-semibold dark:text-white">TypeScript</span> on two backend services. I appreciate TypeScript's static typing, which helps catch errors early in the development process, and strong community support. 
      While I do enjoy <span className="font-semibold dark:text-white">Golang</span>, it lacks the same level of ecosystem maturity. I use <span className="font-semibold dark:text-white">TypeScript</span> with frameworks like <span className="font-semibold dark:text-white">React</span> and <span className="font-semibold dark:text-white">Node.js</span> to enhance code quality and developer productivity.</>)
    },
  ];
  const [selected, setSelected] = useState(0);

  return (
    <section>
      <h2 className="text-2xl font-bold mb-6">Skills</h2>
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
        <div className="max-w-xl mx-auto text-center text-gray-700 bg-gray-800 rounded-lg p-4 shadow-inner rounded-2xl shadow-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 dark:text-gray-300">
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

    if (form.zip.value) {
      // Honeypot field filled, likely a bot submission
      setSubmitted(true)
      return;
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

    const payload = {
      firstName: form.firstName.value,
      lastName: form.lastName.value,
      email: form.email.value,
      phone: form.phone.value,
      company: form.company.value,
      jobTitle: form.jobTitle.value,
      description: form.description.value
    };

    fetch("https://api.sh3r4rd.com/requests", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).catch((err) => console.error("Submission error:", err));
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
              <input name="firstName" type="text" placeholder="First Name" className="p-2 border rounded-md text-black" />
              <input name="lastName" type="text" placeholder="Last Name" className="p-2 border rounded-md text-black" />
            </div>
            <input name="email" type="email" placeholder="Email" className="p-2 border rounded-md text-black" />
            <input name="phone" type="tel" placeholder="Phone Number" className="p-2 border rounded-md text-black" />
            <input name="company" type="text" placeholder="Company" className="p-2 border rounded-md text-black" />
            <input name="jobTitle" type="text" placeholder="Job Title" className="p-2 border rounded-md text-black" />
            <textarea name="description" placeholder="Job Description" rows={6} className="p-2 border rounded-md text-black resize-y" />
            <input name="zip" style={{ display: 'none' }} type="text" />
            <Button type="submit">Submit Request</Button>
          </form>
        </>
      ) : (
        <div className="text-lg text-center text-gray-700 dark:text-gray-300">
          <p>Thank you for your consideration — I appreciate your interest.</p>
          <p>I&apos;ve received your request and will review it shortly. I&apos;ll be in touch soon.</p>

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

function NotFoundPage() {
  return (
    <section className="text-center mt-20">
      <h1 className="text-4xl font-bold mb-4">404 - Page Not Found</h1>
      <p className="text-lg text-gray-600 dark:text-gray-400 mb-6">
        The page you’re looking for doesn’t exist or has been moved.
      </p>
      <Link
        to="/"
        className="text-indigo-600 hover:underline font-medium"
      >
        ← Go back home
      </Link>
    </section>
  );
}

export default function App() {
  return (
    <main className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white p-8">
      <Router>
        <PageTracker />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/resume" element={<ResumeRequestPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Router>
    </main>
  );
}