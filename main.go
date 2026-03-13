package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed templates
var templateFiles embed.FS

var tmpl *template.Template

func init() {
	funcMap := template.FuncMap{
		"renderField": renderField,
		"safeCSS":     func(s string) template.CSS { return template.CSS(s) },
		"add1":        func(i int) int { return i + 1 },
	}

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

func main() {
	port := getEnv("PORT", "5000")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /submit", handleSubmit)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /config", handleConfigPage)
	mux.HandleFunc("GET /config/load", handleConfigLoad)
	mux.HandleFunc("POST /config/validate", handleConfigValidate)
	mux.HandleFunc("POST /config/save", handleConfigSave)

	log.Printf("Web Form Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/config/form.yaml")
	form, err := loadForm(formFile)
	if err != nil {
		http.Error(w, "Form configuration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	css := readCSS()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "form.html", map[string]interface{}{
		"Form": form,
		"CSS":  css,
	}); err != nil {
		log.Printf("Template render error: %v", err)
	}
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/config/form.yaml")
	form, err := loadForm(formFile)
	if err != nil {
		http.Error(w, "Form configuration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse up to 32 MB of multipart data (for file uploads); fall back to plain form.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		r.ParseForm() //nolint:errcheck
	}

	record := handleSubmission(r, form)
	css := readCSS()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "submitted.html", map[string]interface{}{
		"Form":   form,
		"CSS":    css,
		"Record": record,
	}); err != nil {
		log.Printf("Template render error: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	formFile := getEnv("FORM_FILE", "/config/form.yaml")
	_, err := os.Stat(formFile)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
		"status":      "ok",
		"form_loaded": err == nil,
	})
}

func readCSS() string {
	cssFile := getEnv("CSS_FILE", "/config/form.css")
	data, err := os.ReadFile(cssFile)
	if err != nil {
		log.Printf("CSS file not found: %s — using empty styles", cssFile)
		return ""
	}
	return string(data)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
