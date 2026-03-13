package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func handleConfigPage(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/web-config/form.toml")
	tomlContent := ""
	data, err := os.ReadFile(formFile)
	if err != nil {
		log.Printf("Config page: could not read form file %s: %v", formFile, err)
	} else {
		tomlContent = string(data)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "config.html", map[string]interface{}{
		"FormFile":    formFile,
		"TOMLContent": tomlContent,
	}); err != nil {
		log.Printf("Config template render error: %v", err)
	}
}

func handleConfigLoad(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/web-config/form.toml")
	data, err := os.ReadFile(formFile)
	if err != nil {
		http.Error(w, "Could not read form file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data) //nolint:errcheck
}

func handleConfigSave(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/web-config/form.toml")

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read request body"})
		return
	}
	defer r.Body.Close()

	tomlText := string(body)

	// Validate by attempting to parse — reject if invalid.
	if _, err := loadFormFromString(tomlText); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.WriteFile(formFile, []byte(tomlText), 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to write file: " + err.Error()})
		return
	}

	log.Printf("Config saved to %s (%d bytes)", formFile, len(tomlText))
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
