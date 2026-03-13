package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// SubmissionRecord is the structure saved to disk and logged to stdout.
type SubmissionRecord struct {
	Timestamp string                 `json:"timestamp"`
	FormTitle string                 `json:"form_title"`
	Fields    map[string]interface{} `json:"fields"`
	Files     []string               `json:"files"`
}

// handleSubmission builds a record from the request, logs it, and optionally saves it.
func handleSubmission(r *http.Request, form *Form) SubmissionRecord {
	fields := make(map[string]interface{})
	for key, values := range r.Form {
		if len(values) == 1 {
			fields[key] = values[0]
		} else {
			fields[key] = values
		}
	}

	var fileNames []string
	if r.MultipartForm != nil {
		for _, fileHeaders := range r.MultipartForm.File {
			for _, fh := range fileHeaders {
				if fh.Filename != "" {
					fileNames = append(fileNames, fh.Filename)
				}
			}
		}
	}

	record := SubmissionRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		FormTitle: form.Title,
		Fields:    fields,
		Files:     fileNames,
	}

	data, _ := json.Marshal(record)
	log.Printf("FORM_SUBMISSION: %s", data)

	if os.Getenv("SAVE_SUBMISSIONS") == "true" {
		saveSubmission(record)
	}

	return record
}

// saveSubmission writes the record to a timestamped JSON file.
func saveSubmission(record SubmissionRecord) {
	dir := getEnv("SUBMISSIONS_DIR", "/data/submissions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create submissions directory: %v", err)
		return
	}

	filename := fmt.Sprintf("%s.json", time.Now().UTC().Format("20060102_150405_000000"))
	path := dir + "/" + filename

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal submission: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to write submission file: %v", err)
		return
	}

	log.Printf("Submission saved to: %s", path)
}
