package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// FormDefinition is the top-level TOML structure.
type FormDefinition struct {
	Form Form `toml:"form"`
}

// Form holds all form-level metadata.
// Use either Pages (multi-page) or Sections (single-page) — not both.
type Form struct {
	Title       string    `toml:"title"`
	Description string    `toml:"description"`
	Action      string    `toml:"action"`
	Method      string    `toml:"method"`
	Enctype     string    `toml:"enctype"`
	SubmitLabel string    `toml:"submit_label"`
	ResetLabel  string    `toml:"reset_label"`
	Pages       []Page    `toml:"pages"`    // multi-page forms
	Sections    []Section `toml:"sections"` // single-page forms (backward-compatible)
}

// Page groups sections into a discrete step of a multi-page form.
type Page struct {
	ID          string    `toml:"id"`
	Title       string    `toml:"title"`
	Description string    `toml:"description"`
	Sections    []Section `toml:"sections"`
}

// Section groups a set of related fields under an optional heading.
type Section struct {
	ID          string  `toml:"id"`
	Title       string  `toml:"title"`
	Description string  `toml:"description"`
	Fields      []Field `toml:"fields"`
}

// Field represents a single form input of any supported type.
type Field struct {
	ID           string      `toml:"id"`
	Type         string      `toml:"type"`
	Label        string      `toml:"label"`
	Required     bool        `toml:"required"`
	Disabled     bool        `toml:"disabled"`
	Readonly     bool        `toml:"readonly"`
	Value        string      `toml:"value"`
	Placeholder  string      `toml:"placeholder"`
	Min          interface{} `toml:"min"`  // string for dates, int/float for numbers
	Max          interface{} `toml:"max"`
	Step         interface{} `toml:"step"`
	Pattern      string      `toml:"pattern"`
	Minlength    int         `toml:"minlength"`
	Maxlength    int         `toml:"maxlength"`
	Accept       string      `toml:"accept"`
	Multiple     bool        `toml:"multiple"`
	Size         int         `toml:"size"`
	Rows         int         `toml:"rows"`
	Cols         int         `toml:"cols"`
	Checked      bool        `toml:"checked"`
	Autocomplete string      `toml:"autocomplete"`
	Title        string      `toml:"title"` // HTML title attribute (tooltip)
	CSSClass     string      `toml:"css_class"`
	Options      []Option    `toml:"options"`
}

// Option is a selectable item for select, radio, and checkbox_group fields.
type Option struct {
	Label    string `toml:"label"`
	Value    string `toml:"value"`
	Selected bool   `toml:"selected"`
	Checked  bool   `toml:"checked"`
	Disabled bool   `toml:"disabled"`
}

var validTypes = map[string]bool{
	"text": true, "password": true, "email": true, "url": true, "tel": true, "search": true,
	"number": true, "range": true,
	"date": true, "time": true, "datetime-local": true, "month": true, "week": true,
	"color":          true,
	"checkbox":       true,
	"checkbox_group": true,
	"radio":          true,
	"select":         true,
	"textarea":       true,
	"file":           true,
	"hidden":         true,
	"submit":         true,
	"reset":          true,
	"button":         true,
}

var optionTypes = map[string]bool{
	"checkbox_group": true,
	"radio":          true,
	"select":         true,
}

// loadForm reads, parses, and validates a TOML form definition file.
func loadForm(path string) (*Form, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("form file not found: %s", path)
	}
	return loadFormFromString(string(data))
}

// loadFormFromString parses and validates a TOML form definition from a string.
func loadFormFromString(tomlText string) (*Form, error) {
	var def FormDefinition
	if _, err := toml.Decode(tomlText, &def); err != nil {
		return nil, fmt.Errorf("TOML parse error: %w", err)
	}

	form := &def.Form
	if err := validateForm(form); err != nil {
		return nil, err
	}
	applyDefaults(form)
	return form, nil
}

func validateForm(f *Form) error {
	if f.Title == "" {
		return fmt.Errorf("form.title is required")
	}

	hasPaged := len(f.Pages) > 0
	hasFlat := len(f.Sections) > 0

	if hasPaged && hasFlat {
		return fmt.Errorf("form cannot have both 'pages' and 'sections' — use one or the other")
	}
	if !hasPaged && !hasFlat {
		return fmt.Errorf("form requires either 'pages' (multi-page) or 'sections' (single-page)")
	}

	if hasPaged {
		for pi, page := range f.Pages {
			if page.ID == "" {
				return fmt.Errorf("pages[%d] missing required key: id", pi)
			}
			if len(page.Sections) == 0 {
				return fmt.Errorf("pages[%d] must have at least one section", pi)
			}
			for si, s := range page.Sections {
				path := fmt.Sprintf("pages[%d].sections[%d]", pi, si)
				if err := validateSection(s, path); err != nil {
					return err
				}
			}
		}
	} else {
		for si, s := range f.Sections {
			path := fmt.Sprintf("sections[%d]", si)
			if err := validateSection(s, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateSection(s Section, path string) error {
	if s.ID == "" {
		return fmt.Errorf("%s missing required key: id", path)
	}
	if len(s.Fields) == 0 {
		return fmt.Errorf("%s must have at least one field", path)
	}
	for fi, field := range s.Fields {
		fieldPath := fmt.Sprintf("%s.fields[%d]", path, fi)
		if err := validateField(field, fieldPath); err != nil {
			return err
		}
	}
	return nil
}

func validateField(f Field, path string) error {
	if f.ID == "" {
		return fmt.Errorf("%s missing required key: id", path)
	}
	if f.Type == "" {
		return fmt.Errorf("%s missing required key: type", path)
	}
	if f.Type != "hidden" && f.Label == "" {
		return fmt.Errorf("%s missing required key: label", path)
	}
	if !validTypes[f.Type] {
		return fmt.Errorf("%s has unknown type %q", path, f.Type)
	}
	if optionTypes[f.Type] && len(f.Options) == 0 {
		return fmt.Errorf("%s (type=%q) requires an options list", path, f.Type)
	}
	for oi, opt := range f.Options {
		if opt.Value == "" {
			return fmt.Errorf("%s.options[%d] missing required key: value", path, oi)
		}
		if opt.Label == "" {
			return fmt.Errorf("%s.options[%d] missing required key: label", path, oi)
		}
	}
	return nil
}

func applyDefaults(f *Form) {
	if f.Method == "" {
		f.Method = "post"
	}
	if f.Action == "" {
		f.Action = "/submit"
	}
	if f.SubmitLabel == "" {
		f.SubmitLabel = "Submit"
	}
}
