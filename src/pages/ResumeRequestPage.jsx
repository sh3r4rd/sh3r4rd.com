import { useState } from "react";
import { Button } from "../components/ui/button";
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";

export default function ResumeRequestPage() {
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e) => {
    e.preventDefault();
    const form = e.target;
    const nameRegex = /^.{2,}$/;
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    const phoneRegex = /^(\+\d{1,2}\s?)?1?-?\.?\s?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$/;
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