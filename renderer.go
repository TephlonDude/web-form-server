package main

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

// renderField is registered as a template FuncMap function.
// It dispatches on field type and returns safe HTML.
func renderField(f Field) template.HTML {
	switch f.Type {
	case "hidden":
		return template.HTML(fmt.Sprintf(
			`<input type="hidden" name="%s" value="%s">`,
			esc(f.ID), esc(f.Value),
		))
	case "textarea":
		return template.HTML(renderTextarea(f))
	case "select":
		return template.HTML(renderSelect(f))
	case "radio":
		return template.HTML(renderRadioGroup(f))
	case "checkbox_group":
		return template.HTML(renderCheckboxGroup(f))
	case "checkbox":
		return template.HTML(renderCheckbox(f))
	case "submit", "reset", "button":
		return template.HTML(renderButton(f))
	default:
		return template.HTML(renderInput(f))
	}
}

// esc escapes a string for use in an HTML attribute value.
func esc(s string) string {
	return html.EscapeString(s)
}

// txt escapes a string for use as HTML text content.
func txt(s string) string {
	return html.EscapeString(s)
}

// ifaceStr converts an interface{} to its string representation.
// Used for Min/Max/Step which may be int, float64, or string in YAML.
func ifaceStr(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// capitalize uppercases the first character of s.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// requiredAsterisk returns the required-asterisk span if the field is required.
func requiredAsterisk(required bool) string {
	if required {
		return `<span class="required" aria-hidden="true">*</span>`
	}
	return ""
}

// fieldWrapper wraps content in a .field-group div with a label.
func fieldWrapper(f Field, inner string) string {
	cssClass := "field-group"
	if f.CSSClass != "" {
		cssClass += " " + f.CSSClass
	}
	return fmt.Sprintf(
		`<div class="%s"><label for="%s">%s%s</label>%s</div>`,
		cssClass, esc(f.ID), txt(f.Label), requiredAsterisk(f.Required), inner,
	)
}

// commonAttrs builds the shared HTML attributes for simple <input> elements.
func commonAttrs(f Field) string {
	var b strings.Builder
	w := func(s string) { b.WriteString(s) }

	if f.Value != "" {
		w(fmt.Sprintf(` value="%s"`, esc(f.Value)))
	}
	if f.Placeholder != "" {
		w(fmt.Sprintf(` placeholder="%s"`, esc(f.Placeholder)))
	}
	if v := ifaceStr(f.Min); v != "" {
		w(fmt.Sprintf(` min="%s"`, esc(v)))
	}
	if v := ifaceStr(f.Max); v != "" {
		w(fmt.Sprintf(` max="%s"`, esc(v)))
	}
	if v := ifaceStr(f.Step); v != "" {
		w(fmt.Sprintf(` step="%s"`, esc(v)))
	}
	if f.Pattern != "" {
		w(fmt.Sprintf(` pattern="%s"`, esc(f.Pattern)))
	}
	if f.Minlength > 0 {
		w(fmt.Sprintf(` minlength="%d"`, f.Minlength))
	}
	if f.Maxlength > 0 {
		w(fmt.Sprintf(` maxlength="%d"`, f.Maxlength))
	}
	if f.Accept != "" {
		w(fmt.Sprintf(` accept="%s"`, esc(f.Accept)))
	}
	if f.Multiple {
		w(` multiple`)
	}
	if f.Required {
		w(` required`)
	}
	if f.Readonly {
		w(` readonly`)
	}
	if f.Disabled {
		w(` disabled`)
	}
	if f.Autocomplete != "" {
		w(fmt.Sprintf(` autocomplete="%s"`, esc(f.Autocomplete)))
	}
	if f.Title != "" {
		w(fmt.Sprintf(` title="%s"`, esc(f.Title)))
	}
	return b.String()
}

// renderInput handles all simple <input> types (text, email, number, date, range, color, file, …).
func renderInput(f Field) string {
	inner := fmt.Sprintf(
		`<input type="%s" id="%s" name="%s"%s>`,
		esc(f.Type), esc(f.ID), esc(f.ID), commonAttrs(f),
	)
	if f.Type == "range" {
		inner += fmt.Sprintf(
			`<output for="%s" id="%s_output">%s</output>`,
			esc(f.ID), esc(f.ID), txt(f.Value),
		)
	}
	return fieldWrapper(f, inner)
}

// renderTextarea handles <textarea> elements.
func renderTextarea(f Field) string {
	var attrs strings.Builder
	w := func(s string) { attrs.WriteString(s) }

	if f.Rows > 0 {
		w(fmt.Sprintf(` rows="%d"`, f.Rows))
	}
	if f.Cols > 0 {
		w(fmt.Sprintf(` cols="%d"`, f.Cols))
	}
	if f.Placeholder != "" {
		w(fmt.Sprintf(` placeholder="%s"`, esc(f.Placeholder)))
	}
	if f.Maxlength > 0 {
		w(fmt.Sprintf(` maxlength="%d"`, f.Maxlength))
	}
	if f.Required {
		w(` required`)
	}
	if f.Readonly {
		w(` readonly`)
	}
	if f.Disabled {
		w(` disabled`)
	}
	if f.Title != "" {
		w(fmt.Sprintf(` title="%s"`, esc(f.Title)))
	}
	if f.Autocomplete != "" {
		w(fmt.Sprintf(` autocomplete="%s"`, esc(f.Autocomplete)))
	}

	inner := fmt.Sprintf(
		`<textarea id="%s" name="%s"%s>%s</textarea>`,
		esc(f.ID), esc(f.ID), attrs.String(), txt(f.Value),
	)
	return fieldWrapper(f, inner)
}

// renderSelect handles <select> elements, both single and multi.
func renderSelect(f Field) string {
	name := f.ID
	if f.Multiple {
		name += "[]"
	}

	var selectAttrs strings.Builder
	if f.Multiple {
		size := f.Size
		if size == 0 {
			size = 4
		}
		selectAttrs.WriteString(fmt.Sprintf(` multiple size="%d"`, size))
	}
	if f.Required {
		selectAttrs.WriteString(` required`)
	}
	if f.Disabled {
		selectAttrs.WriteString(` disabled`)
	}
	if f.Title != "" {
		selectAttrs.WriteString(fmt.Sprintf(` title="%s"`, esc(f.Title)))
	}

	var opts strings.Builder
	if f.Placeholder != "" {
		opts.WriteString(fmt.Sprintf(
			`<option value="" disabled selected>%s</option>`, txt(f.Placeholder),
		))
	}
	for _, opt := range f.Options {
		var flags strings.Builder
		if opt.Selected {
			flags.WriteString(` selected`)
		}
		if opt.Disabled {
			flags.WriteString(` disabled`)
		}
		opts.WriteString(fmt.Sprintf(
			`<option value="%s"%s>%s</option>`,
			esc(opt.Value), flags.String(), txt(opt.Label),
		))
	}

	inner := fmt.Sprintf(
		`<select id="%s" name="%s"%s>%s</select>`,
		esc(f.ID), esc(name), selectAttrs.String(), opts.String(),
	)
	return fieldWrapper(f, inner)
}

// renderRadioGroup renders a set of radio buttons as a <fieldset>.
func renderRadioGroup(f Field) string {
	cssClass := "field-group radio-group"
	if f.CSSClass != "" {
		cssClass += " " + f.CSSClass
	}

	var opts strings.Builder
	for _, opt := range f.Options {
		var flags strings.Builder
		if opt.Checked {
			flags.WriteString(` checked`)
		}
		if opt.Disabled {
			flags.WriteString(` disabled`)
		}
		if f.Required {
			flags.WriteString(` required`)
		}
		opts.WriteString(fmt.Sprintf(
			`<label class="radio-option"><input type="radio" name="%s" value="%s"%s> %s</label>`,
			esc(f.ID), esc(opt.Value), flags.String(), txt(opt.Label),
		))
	}

	return fmt.Sprintf(
		`<fieldset class="%s"><legend>%s%s</legend>%s</fieldset>`,
		cssClass, txt(f.Label), requiredAsterisk(f.Required), opts.String(),
	)
}

// renderCheckboxGroup renders multiple checkboxes that submit as a list (name[]).
func renderCheckboxGroup(f Field) string {
	cssClass := "field-group checkbox-group"
	if f.CSSClass != "" {
		cssClass += " " + f.CSSClass
	}

	var opts strings.Builder
	for _, opt := range f.Options {
		var flags strings.Builder
		if opt.Checked {
			flags.WriteString(` checked`)
		}
		if opt.Disabled {
			flags.WriteString(` disabled`)
		}
		opts.WriteString(fmt.Sprintf(
			`<label class="checkbox-option"><input type="checkbox" name="%s[]" value="%s"%s> %s</label>`,
			esc(f.ID), esc(opt.Value), flags.String(), txt(opt.Label),
		))
	}

	return fmt.Sprintf(
		`<fieldset class="%s"><legend>%s</legend>%s</fieldset>`,
		cssClass, txt(f.Label), opts.String(),
	)
}

// renderCheckbox renders a single boolean checkbox.
func renderCheckbox(f Field) string {
	value := f.Value
	if value == "" {
		value = "on"
	}

	var flags strings.Builder
	if f.Checked {
		flags.WriteString(` checked`)
	}
	if f.Required {
		flags.WriteString(` required`)
	}
	if f.Disabled {
		flags.WriteString(` disabled`)
	}
	if f.Readonly {
		flags.WriteString(` readonly`)
	}
	if f.Title != "" {
		flags.WriteString(fmt.Sprintf(` title="%s"`, esc(f.Title)))
	}

	cssClass := "field-group checkbox-single"
	if f.CSSClass != "" {
		cssClass += " " + f.CSSClass
	}

	return fmt.Sprintf(
		`<div class="%s"><label><input type="checkbox" id="%s" name="%s" value="%s"%s> %s%s</label></div>`,
		cssClass, esc(f.ID), esc(f.ID), esc(value), flags.String(),
		txt(f.Label), requiredAsterisk(f.Required),
	)
}

// renderButton renders a <button> element (submit, reset, or button type).
func renderButton(f Field) string {
	label := f.Value
	if label == "" {
		label = capitalize(f.Type)
	}

	var attrs strings.Builder
	if f.CSSClass != "" {
		attrs.WriteString(fmt.Sprintf(` class="%s"`, esc(f.CSSClass)))
	}
	if f.Disabled {
		attrs.WriteString(` disabled`)
	}
	if f.Title != "" {
		attrs.WriteString(fmt.Sprintf(` title="%s"`, esc(f.Title)))
	}

	return fmt.Sprintf(
		`<button type="%s" id="%s"%s>%s</button>`,
		esc(f.Type), esc(f.ID), attrs.String(), txt(label),
	)
}
