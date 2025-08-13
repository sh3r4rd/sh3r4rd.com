import { useState } from "react";

export default function SkillsSection() {
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