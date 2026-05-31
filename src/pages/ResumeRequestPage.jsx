import { useState } from "react";
import { Link } from "react-router-dom";
import { CheckCircle2 } from "lucide-react";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import PageLayout from "../components/layout/PageLayout";

const inputClass =
  "w-full rounded-xl px-4 py-3 bg-white/70 dark:bg-white/5 backdrop-blur border border-gray-300 dark:border-white/10 text-gray-900 dark:text-gray-100 placeholder-gray-400 transition focus:border-purple-500 focus:ring-2 focus:ring-purple-500/40";

function FieldError({ id, message }) {
  if (!message) return null;
  return (
    <p id={id} className="mt-1 text-xs text-red-600 dark:text-red-400">
      {message}
    </p>
  );
}

export default function ResumeRequestPage() {
  const [submitted, setSubmitted] = useState(false);
  const [errors, setErrors] = useState({});
  const [wordCount, setWordCount] = useState(0);

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

    const newErrors = {};
    const setError = (field, message) => {
      newErrors[field] = message;
    };

    for (let id of fields) {
      const el = form.elements[id];
      if (!el || !el.value.trim()) {
        setError(id, `Please fill in the ${id} field.`);
        setErrors(newErrors);
        if (el) el.focus();
        return;
      }
    }

    if (form.zip.value) {
      // Honeypot field filled, likely a bot submission
      setSubmitted(true);
      return;
    }

    if (!nameRegex.test(form.firstName.value)) {
      setError("firstName", "First name must be at least 2 characters.");
      setErrors(newErrors);
      form.firstName.focus();
      return;
    }

    if (!nameRegex.test(form.lastName.value)) {
      setError("lastName", "Last name must be at least 2 characters.");
      setErrors(newErrors);
      form.lastName.focus();
      return;
    }

    if (!emailRegex.test(form.email.value)) {
      setError("email", "Please enter a valid email address.");
      setErrors(newErrors);
      form.email.focus();
      return;
    }

    if (!phoneRegex.test(form.phone.value)) {
      setError("phone", "Please enter a valid US phone number");
      setErrors(newErrors);
      form.phone.focus();
      return;
    }

    if (form.jobTitle.value.trim().length < jobTitleMinLength) {
      setError("jobTitle", "Job title must be at least 10 characters long.");
      setErrors(newErrors);
      form.jobTitle.focus();
      return;
    }

    const wc = form.description.value.trim().split(/\s+/).length;
    if (wc < descriptionMinWords) {
      setError("description", "Job description must be at least 50 words.");
      setErrors(newErrors);
      form.description.focus();
      return;
    }

    setErrors({});

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

  const handleDescriptionChange = (e) => {
    const value = e.target.value.trim();
    setWordCount(value ? value.split(/\s+/).length : 0);
  };

  const descriptionMet = wordCount >= 50;

  return (
    <PageLayout variant="slim">
      {!submitted ? (
        <>
          <p className="text-lg text-gray-700 dark:text-gray-300 mb-8">
            Thank you for your interest! Please fill out the form below to request a copy of my resume.
          </p>
          <Card>
            <CardContent className="p-6">
              <form className="grid gap-8" onSubmit={handleSubmit} noValidate>
                {/* About you */}
                <fieldset className="grid gap-4">
                  <legend className="text-xs font-semibold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-2">
                    About you
                  </legend>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label htmlFor="firstName" className="sr-only">First Name</label>
                      <input
                        id="firstName"
                        name="firstName"
                        type="text"
                        placeholder="First Name"
                        className={inputClass}
                        aria-invalid={!!errors.firstName}
                        aria-describedby={errors.firstName ? "firstName-error" : undefined}
                      />
                      <FieldError id="firstName-error" message={errors.firstName} />
                    </div>
                    <div>
                      <label htmlFor="lastName" className="sr-only">Last Name</label>
                      <input
                        id="lastName"
                        name="lastName"
                        type="text"
                        placeholder="Last Name"
                        className={inputClass}
                        aria-invalid={!!errors.lastName}
                        aria-describedby={errors.lastName ? "lastName-error" : undefined}
                      />
                      <FieldError id="lastName-error" message={errors.lastName} />
                    </div>
                  </div>
                  <div>
                    <label htmlFor="email" className="sr-only">Email</label>
                    <input
                      id="email"
                      name="email"
                      type="email"
                      placeholder="Email"
                      className={inputClass}
                      aria-invalid={!!errors.email}
                      aria-describedby={errors.email ? "email-error" : undefined}
                    />
                    <FieldError id="email-error" message={errors.email} />
                  </div>
                  <div>
                    <label htmlFor="phone" className="sr-only">Phone Number</label>
                    <input
                      id="phone"
                      name="phone"
                      type="tel"
                      placeholder="Phone Number"
                      className={inputClass}
                      aria-invalid={!!errors.phone}
                      aria-describedby={errors.phone ? "phone-error" : undefined}
                    />
                    <FieldError id="phone-error" message={errors.phone} />
                  </div>
                </fieldset>

                {/* About the role */}
                <fieldset className="grid gap-4">
                  <legend className="text-xs font-semibold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-2">
                    About the role
                  </legend>
                  <div>
                    <label htmlFor="company" className="sr-only">Company</label>
                    <input
                      id="company"
                      name="company"
                      type="text"
                      placeholder="Company"
                      className={inputClass}
                      aria-invalid={!!errors.company}
                      aria-describedby={errors.company ? "company-error" : undefined}
                    />
                    <FieldError id="company-error" message={errors.company} />
                  </div>
                  <div>
                    <label htmlFor="jobTitle" className="sr-only">Job Title</label>
                    <input
                      id="jobTitle"
                      name="jobTitle"
                      type="text"
                      placeholder="Job Title"
                      className={inputClass}
                      aria-invalid={!!errors.jobTitle}
                      aria-describedby={errors.jobTitle ? "jobTitle-error" : undefined}
                    />
                    <FieldError id="jobTitle-error" message={errors.jobTitle} />
                  </div>
                  <div>
                    <label htmlFor="description" className="sr-only">Job Description</label>
                    <textarea
                      id="description"
                      name="description"
                      placeholder="Job Description"
                      rows={6}
                      onChange={handleDescriptionChange}
                      className={`${inputClass} resize-y`}
                      aria-invalid={!!errors.description}
                      aria-describedby={errors.description ? "description-error" : undefined}
                    />
                    <p
                      className={`mt-1 text-xs ${descriptionMet ? "text-teal-600 dark:text-teal-400" : "text-gray-500 dark:text-gray-400"}`}
                    >
                      {wordCount} / 50 words
                    </p>
                    <FieldError id="description-error" message={errors.description} />
                  </div>
                </fieldset>

                <input name="zip" style={{ display: "none" }} type="text" />
                <Button type="submit" variant="primary" size="lg" className="w-full">
                  Submit Request
                </Button>
              </form>
            </CardContent>
          </Card>
        </>
      ) : (
        <div className="text-lg text-center text-gray-700 dark:text-gray-300">
          <div className="flex justify-center mb-4">
            <CheckCircle2 className="w-12 h-12 text-teal-500" />
          </div>
          <p>Thank you for your consideration — I appreciate your interest.</p>
          <p>I&apos;ve received your request and will review it shortly. I&apos;ll be in touch soon.</p>

          <div className="max-w-md mx-auto">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
              <path
                fill="#9333ea"
                stroke="#9333ea"
                strokeWidth="15"
                style={{ transformOrigin: "center" }}
                d="m148 84.7 13.8-8-10-17.3-13.8 8a50 50 0 0 0-27.4-15.9v-16h-20v16A50 50 0 0 0 63 67.4l-13.8-8-10 17.3 13.8 8a50 50 0 0 0 0 31.7l-13.8 8 10 17.3 13.8-8a50 50 0 0 0 27.5 15.9v16h20v-16a50 50 0 0 0 27.4-15.9l13.8 8 10-17.3-13.8-8a50 50 0 0 0 0-31.7Zm-47.5 50.8a35 35 0 1 1 0-70 35 35 0 0 1 0 70Z"
              >
                <animateTransform type="rotate" attributeName="transform" calcMode="spline" dur="1.5" values="0;120" keyTimes="0;1" keySplines="0 0 1 1" repeatCount="indefinite"></animateTransform>
              </path>
            </svg>
          </div>

          <Link to="/" className="inline-block mt-4 font-medium text-teal-600 dark:text-teal-400 hover:underline">
            ← Back to portfolio
          </Link>
        </div>
      )}
    </PageLayout>
  );
}
