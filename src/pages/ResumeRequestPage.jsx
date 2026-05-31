import { useState } from "react";
import { Link } from "react-router-dom";
import { CheckCircle2 } from "lucide-react";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import PageLayout from "../components/layout/PageLayout";

// Single source of truth for validation rules shared by the form, the live
// word counter, and the error messages.
const MIN_DESCRIPTION_WORDS = 50;
const JOB_TITLE_MIN_LENGTH = 10;
const nameRegex = /^.{2,}$/;
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const phoneRegex = /^(\+\d{1,2}\s?)?1?-?\.?\s?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$/;

const countWords = (value) => {
  const trimmed = value.trim();
  return trimmed ? trimmed.split(/\s+/).length : 0;
};

// Field definitions drive both rendering and validation so the two can't drift.
// `validate` runs only after the required-field check passes; it returns an
// error message or null.
const FIELD_GROUPS = [
  {
    legend: "About you",
    fields: [
      { id: "firstName", label: "First Name", type: "text", half: true, validate: (v) => (nameRegex.test(v) ? null : "First name must be at least 2 characters.") },
      { id: "lastName", label: "Last Name", type: "text", half: true, validate: (v) => (nameRegex.test(v) ? null : "Last name must be at least 2 characters.") },
      { id: "email", label: "Email", type: "email", validate: (v) => (emailRegex.test(v) ? null : "Please enter a valid email address.") },
      { id: "phone", label: "Phone Number", type: "tel", validate: (v) => (phoneRegex.test(v) ? null : "Please enter a valid US phone number.") },
    ],
  },
  {
    legend: "About the role",
    fields: [
      { id: "company", label: "Company", type: "text" },
      { id: "jobTitle", label: "Job Title", type: "text", validate: (v) => (v.trim().length >= JOB_TITLE_MIN_LENGTH ? null : "Job title must be at least 10 characters long.") },
      { id: "description", label: "Job Description", type: "textarea", validate: (v) => (countWords(v) >= MIN_DESCRIPTION_WORDS ? null : `Job description must be at least ${MIN_DESCRIPTION_WORDS} words.`) },
    ],
  },
];

const ALL_FIELDS = FIELD_GROUPS.flatMap((group) => group.fields);

const inputClass =
  "w-full rounded-xl px-4 py-3 bg-white/70 dark:bg-white/5 backdrop-blur border border-gray-300 dark:border-white/10 text-gray-900 dark:text-gray-100 placeholder-gray-400 transition focus:border-purple-500 focus:ring-2 focus:ring-purple-500/40 aria-invalid:border-red-500 dark:aria-invalid:border-red-400";

function FieldError({ id, message }) {
  if (!message) return null;
  return (
    <p id={id} className="mt-1 text-xs text-red-600 dark:text-red-400">
      {message}
    </p>
  );
}

function Field({ field, error }) {
  const { id, label, type } = field;
  return (
    <div>
      <label htmlFor={id} className="sr-only">{label}</label>
      <input
        id={id}
        name={id}
        type={type}
        placeholder={label}
        className={inputClass}
        aria-invalid={!!error}
        aria-describedby={error ? `${id}-error` : undefined}
      />
      <FieldError id={`${id}-error`} message={error} />
    </div>
  );
}

export default function ResumeRequestPage() {
  const [submitted, setSubmitted] = useState(false);
  const [errors, setErrors] = useState({});
  const [wordCount, setWordCount] = useState(0);

  const handleSubmit = (e) => {
    e.preventDefault();
    const form = e.target;

    // Honeypot: a filled hidden field means a bot — show success without sending.
    if (form.zip.value) {
      setSubmitted(true);
      return;
    }

    // Collect every field's error in one pass so all of them surface at once.
    const newErrors = {};
    for (const field of ALL_FIELDS) {
      const value = form.elements[field.id]?.value ?? "";
      if (!value.trim()) {
        newErrors[field.id] = `Please fill in the ${field.label} field.`;
      } else if (field.validate) {
        const message = field.validate(value);
        if (message) newErrors[field.id] = message;
      }
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      const firstInvalid = ALL_FIELDS.find((field) => newErrors[field.id]);
      form.elements[firstInvalid.id]?.focus();
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
      description: form.description.value,
    };

    fetch("https://api.sh3r4rd.com/requests", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).catch((err) => console.error("Submission error:", err));
    setSubmitted(true);
  };

  const handleDescriptionChange = (e) => {
    setWordCount(countWords(e.target.value));
  };

  const descriptionMet = wordCount >= MIN_DESCRIPTION_WORDS;

  const renderField = (field) => {
    if (field.type === "textarea") {
      const error = errors[field.id];
      return (
        <div key={field.id}>
          <label htmlFor={field.id} className="sr-only">{field.label}</label>
          <textarea
            id={field.id}
            name={field.id}
            placeholder={field.label}
            rows={6}
            onChange={handleDescriptionChange}
            className={`${inputClass} resize-y`}
            aria-invalid={!!error}
            aria-describedby={error ? `${field.id}-error` : undefined}
          />
          <p
            className={`mt-1 text-xs ${descriptionMet ? "text-teal-600 dark:text-teal-400" : "text-gray-500 dark:text-gray-400"}`}
          >
            {wordCount} / {MIN_DESCRIPTION_WORDS} words
          </p>
          <FieldError id={`${field.id}-error`} message={error} />
        </div>
      );
    }
    return <Field key={field.id} field={field} error={errors[field.id]} />;
  };

  // Render fields in order, pairing consecutive `half` fields into one row.
  const renderGroupFields = (fields) => {
    const out = [];
    for (let i = 0; i < fields.length; i++) {
      const field = fields[i];
      const next = fields[i + 1];
      if (field.half && next?.half) {
        out.push(
          <div key={field.id} className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {renderField(field)}
            {renderField(next)}
          </div>
        );
        i++;
      } else {
        out.push(renderField(field));
      }
    }
    return out;
  };

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
                {FIELD_GROUPS.map((group) => (
                  <fieldset key={group.legend} className="grid gap-4">
                    <legend className="text-xs font-semibold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-2">
                      {group.legend}
                    </legend>
                    {renderGroupFields(group.fields)}
                  </fieldset>
                ))}

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
