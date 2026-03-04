package models

// ExtractionResult matches the OpenAI structured output JSON schema
// used for recruiter data extraction from email body text.
type ExtractionResult struct {
	RecruiterName  string  `json:"recruiter_name"`
	RecruiterEmail string  `json:"recruiter_email"`
	Company        string  `json:"company"`
	JobTitle       string  `json:"job_title"`
	Phone          string  `json:"phone"`
	Confidence     float64 `json:"confidence"`
}

// IsEmpty returns true if all string fields are empty or "Unknown".
func (r ExtractionResult) IsEmpty() bool {
	return (r.RecruiterName == "" || r.RecruiterName == "Unknown") &&
		(r.RecruiterEmail == "" || r.RecruiterEmail == "Unknown") &&
		(r.Company == "" || r.Company == "Unknown") &&
		(r.JobTitle == "" || r.JobTitle == "Unknown")
}

// UnknownResult returns an ExtractionResult with all fields set to "Unknown"
// and confidence 0.0, used as a fallback when extraction fails.
func UnknownResult() ExtractionResult {
	return ExtractionResult{
		RecruiterName:  "Unknown",
		RecruiterEmail: "Unknown",
		Company:        "Unknown",
		JobTitle:       "Unknown",
		Phone:          "Unknown",
		Confidence:     0.0,
	}
}
