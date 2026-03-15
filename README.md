# Web Form Server

A Docker-containerized web server that renders fully functional HTML forms from a JSON configuration file. Deploy any form by editing two files — no code changes needed.

- **Config-driven** — define your entire form in a `.json` file
- **Multi-page support** — step indicator, progress bar, per-page validation
- **All HTML input types** — text, email, date, file upload, select, radio, checkbox groups, range sliders, color pickers, and more
- **Visual editor** — built-in `/config` page with a JSON text editor and point-and-click form builder
- **Tiny footprint** — ~15 MB Alpine-based Docker image, single static binary
- **Submission persistence** — logs to stdout and optionally saves timestamped JSON files

---

## Quick Start

```bash
git clone https://github.com/TephlonDude/web-form-server.git
cd web-form-server
docker compose up -d
```

Open [http://localhost:8237](http://localhost:8237) to see the demo form.
Open [http://localhost:8237/config](http://localhost:8237/config) to edit it.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 — stdlib only (`net/http`, `html/template`, `encoding/json`, `embed`) |
| Container | Docker + Docker Compose — multi-stage build, ~15 MB Alpine image |
| Config editor | CodeMirror 5 (v5.65.16, cdnjs) — syntax highlighting, line numbers, code folding |
| Client-side | Vanilla JavaScript — Visual Builder, multi-page navigation |
| Form validation | HTML5 `checkValidity()` API — per-page validation in multi-step forms |

---

## Configuration

The form is defined in a JSON file (default: `web-config/example.form.json`). Point to your own file with the `FORM_FILE` environment variable.

### Form Structure

Use **one** of two layouts:

| Layout | When to use | Top-level key |
|---|---|---|
| Single-page | One scrollable page | `"sections"` |
| Multi-page | Wizard with steps | `"pages"` |

#### Single-page

```json
{
  "form": {
    "title": "Contact Us",
    "submit_label": "Send Message",
    "sections": [
      {
        "id": "contact",
        "title": "Your Details",
        "fields": [
          {
            "id": "full_name",
            "type": "text",
            "label": "Full Name",
            "required": true
          }
        ]
      }
    ]
  }
}
```

#### Multi-page

```json
{
  "form": {
    "title": "Registration",
    "submit_label": "Complete Registration",
    "pages": [
      {
        "id": "step1",
        "title": "Personal Info",
        "description": "Tell us about yourself.",
        "sections": [
          {
            "id": "basics",
            "title": "Basic Details",
            "fields": [
              {
                "id": "full_name",
                "type": "text",
                "label": "Full Name"
              }
            ]
          }
        ]
      },
      {
        "id": "step2",
        "title": "Preferences",
        "sections": [
          {
            "id": "prefs",
            "title": "Your Preferences",
            "fields": [
              {
                "id": "newsletter",
                "type": "checkbox",
                "label": "Subscribe to newsletter",
                "value": "yes"
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### Form-level Keys

| Key | Required | Description |
|---|---|---|
| `title` | Yes | Page `<title>` and form heading |
| `description` | | Subtitle shown below the heading |
| `action` | | Form action URL (default: `/submit`) |
| `method` | | `post` or `get` (default: `post`) |
| `enctype` | | Set to `multipart/form-data` when using file fields |
| `submit_label` | | Submit button text (default: `Submit`) |
| `reset_label` | | Reset button text; omit to hide the button |

### Section Keys

| Key | Required | Description |
|---|---|---|
| `id` | Yes | Unique identifier, used as HTML `id` |
| `title` | | Section heading rendered as `<legend>` |
| `description` | | Optional paragraph below the heading |

### Universal Field Keys

Every field type supports these keys:

| Key | Type | Description |
|---|---|---|
| `id` | string | **Required.** HTML `name` and `id` attribute |
| `type` | string | **Required.** See field types below |
| `label` | string | **Required** (except `hidden`). Visible label text |
| `required` | bool | Browser validation — blocks submission if empty |
| `disabled` | bool | Greys out the field; value not submitted |
| `readonly` | bool | Prevents editing; value is submitted |
| `value` | string | Pre-populated default value |
| `placeholder` | string | Ghost text shown when empty |
| `autocomplete` | string | HTML `autocomplete` attribute (e.g. `name`, `email`, `off`) |
| `title` | string | Tooltip shown on hover |
| `css_class` | string | Extra CSS class added to the field wrapper |

---

## Field Types

### Text Inputs

`text` `password` `email` `url` `tel` `search`

```json
{
  "id": "full_name",
  "type": "text",
  "label": "Full Name",
  "required": true,
  "placeholder": "Jane Doe",
  "pattern": "[A-Za-z ]+",
  "minlength": 2,
  "maxlength": 100,
  "autocomplete": "name"
}
```

| Extra key | Description |
|---|---|
| `pattern` | Regex the value must match |
| `minlength` | Minimum character count |
| `maxlength` | Maximum character count |

---

### Numeric

`number` `range`

```json
{
  "id": "age",
  "type": "number",
  "label": "Age",
  "min": 0,
  "max": 120,
  "step": 1
}
```

```json
{
  "id": "rating",
  "type": "range",
  "label": "Satisfaction (1–10)",
  "min": 1,
  "max": 10,
  "step": 1,
  "value": "5"
}
```

Range fields render a live `<output>` display alongside the slider.

| Extra key | Description |
|---|---|
| `min` | Minimum value |
| `max` | Maximum value |
| `step` | Increment size |

---

### Date & Time

`date` `time` `datetime-local` `month` `week`

```json
{
  "id": "birth_date",
  "type": "date",
  "label": "Date of Birth",
  "min": "1900-01-01",
  "max": "2010-12-31"
}
```

```json
{
  "id": "appt_time",
  "type": "time",
  "label": "Appointment Time",
  "min": "08:00",
  "max": "17:00",
  "step": 900
}
```

```json
{
  "id": "meeting",
  "type": "datetime-local",
  "label": "Meeting Date & Time",
  "min": "2026-01-01T08:00",
  "max": "2027-12-31T17:00"
}
```

`min` and `max` accept the format matching the field type:
- `date` → `YYYY-MM-DD`
- `time` → `HH:MM`
- `datetime-local` → `YYYY-MM-DDTHH:MM`
- `month` → `YYYY-MM`
- `week` → `YYYY-Www`

`step` for `time` is in seconds (e.g. `900` = 15-minute intervals).

---

### Color

```json
{
  "id": "brand_color",
  "type": "color",
  "label": "Brand Color",
  "value": "#3498db"
}
```

---

### Textarea

```json
{
  "id": "comments",
  "type": "textarea",
  "label": "Additional Comments",
  "placeholder": "Enter your thoughts here…",
  "rows": 6,
  "cols": 60,
  "maxlength": 2000
}
```

| Extra key | Description |
|---|---|
| `rows` | Visible row count |
| `cols` | Visible column count |
| `maxlength` | Maximum character count |

---

### Select (single and multi)

```json
{
  "id": "country",
  "type": "select",
  "label": "Country",
  "required": true,
  "placeholder": "-- Select a country --",
  "options": [
    { "label": "United States", "value": "us" },
    { "label": "Canada",        "value": "ca", "selected": true },
    { "label": "Other",         "value": "other" }
  ]
}
```

```json
{
  "id": "languages",
  "type": "select",
  "label": "Languages Spoken",
  "multiple": true,
  "size": 5,
  "options": [
    { "label": "English", "value": "en", "selected": true },
    { "label": "French",  "value": "fr" },
    { "label": "Spanish", "value": "es" }
  ]
}
```

| Extra key | Description |
|---|---|
| `placeholder` | Disabled first option (single select only) |
| `multiple` | Allow multiple selections; submits as a list |
| `size` | Number of visible options (multi-select) |

**Option keys:** `label` (required), `value` (required), `selected`, `disabled`

---

### Radio Group

```json
{
  "id": "plan",
  "type": "radio",
  "label": "Choose Plan",
  "required": true,
  "options": [
    { "label": "Free",       "value": "free" },
    { "label": "Pro",        "value": "pro",        "checked": true },
    { "label": "Enterprise", "value": "enterprise", "disabled": true }
  ]
}
```

**Option keys:** `label` (required), `value` (required), `checked`, `disabled`

---

### Checkbox (single)

```json
{
  "id": "agree_terms",
  "type": "checkbox",
  "label": "I agree to the Terms of Service",
  "required": true,
  "value": "yes"
}
```

`value` is the submitted value when checked (default: `"on"`).

---

### Checkbox Group

Submits as `fieldname[]` so the server receives a list.

```json
{
  "id": "interests",
  "type": "checkbox_group",
  "label": "Interests",
  "options": [
    { "label": "Technology", "value": "tech",    "checked": true },
    { "label": "Sports",     "value": "sports" },
    { "label": "Music",      "value": "music",   "disabled": true }
  ]
}
```

---

### File Upload

```json
{
  "id": "resume",
  "type": "file",
  "label": "Upload Resume",
  "accept": ".pdf,.doc,.docx"
}
```

```json
{
  "id": "photos",
  "type": "file",
  "label": "Upload Photos",
  "accept": "image/*",
  "multiple": true
}
```

When any file field is present, set `"enctype": "multipart/form-data"` on the form.

---

### Hidden

```json
{
  "id": "form_version",
  "type": "hidden",
  "value": "2.0.0"
}
```

---

### Button

```json
{
  "id": "preview_btn",
  "type": "button",
  "value": "Preview",
  "css_class": "btn-secondary",
  "title": "Preview before submitting"
}
```

---

## Visual Config Editor

Navigate to `/config` while the container is running.

**JSON Editor tab** — Edit the raw JSON file directly with Load and Save buttons. The save endpoint validates the JSON before writing; invalid configs are rejected with an error message.

**Visual Builder tab** — Point-and-click form construction:
- Tree panel shows the full page → section → field hierarchy
- Click any node to edit its properties in the right panel
- Type-aware field editor shows only the relevant properties for each field type
- Options table for `select`, `radio`, and `checkbox_group` fields
- "Apply to JSON Editor" generates the JSON and switches tabs for review before saving

---

## Submissions

Every submission is:

1. **Logged to stdout** as a JSON line:
   ```json
   {"timestamp":"2026-03-12T14:30:00Z","form_title":"Contact Us","fields":{"full_name":"Jane Doe","email":"jane@example.com"},"files":[]}
   ```

2. **Optionally saved** to a JSON file when `SAVE_SUBMISSIONS=true`:
   ```
   data/submissions/20260312_143000_000000.json
   ```
   Files are named with a UTC timestamp (sorts chronologically). The directory is volume-mounted so files persist across container restarts.

---

## Styling

The CSS file (default: `web-config/example.form.css`) is loaded at startup and injected inline — one HTTP request returns a fully self-contained page.

Key CSS hook points:

```css
/* Page wrapper */
body, .form-container {}

/* Section grouping */
fieldset, legend {}

/* Per-field wrapper (all types) */
.field-group, .field-group label {}

/* Input elements */
input[type="text"], input[type="email"], textarea, select {}

/* Choice groups */
.radio-group .radio-option {}
.checkbox-group .checkbox-option {}
.checkbox-single {}

/* Range slider + live value */
input[type="range"], output {}

/* Required asterisk */
.required {}

/* Submit area */
.form-actions {}
button[type="submit"], button[type="reset"] {}

/* Browser validation states */
input:invalid, input:valid {}

/* Multi-page navigation */
.page-steps, .step-item, .step-bubble, .step-label {}
.progress-bar, .progress-fill {}
.page-actions, .btn-next, .btn-prev {}
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `FORM_FILE` | `/web-config/form.json` | Path to the JSON form definition |
| `CSS_FILE` | `/web-config/form.css` | Path to the CSS stylesheet |
| `PORT` | `5000` | Container listen port |
| `SAVE_SUBMISSIONS` | `false` | Write JSON files to `SUBMISSIONS_DIR` when `true` |
| `SUBMISSIONS_DIR` | `/data/submissions` | Directory for saved submission JSON files |
| `CONFIG_USER` | *(unset)* | Username to protect `/config` with HTTP Basic Auth |
| `CONFIG_PASS` | *(unset)* | Password to protect `/config` with HTTP Basic Auth |

> **Security:** Always set `CONFIG_USER` and `CONFIG_PASS` on any internet-facing deployment to protect the form editor.

Override any variable in `docker-compose.yml`:

```yaml
environment:
  FORM_FILE: /web-config/my-form.json
  CSS_FILE:  /web-config/my-form.css
  SAVE_SUBMISSIONS: "true"
```

---

## Deployment

### Local (Docker Compose)

```bash
docker compose up -d --build   # Build and start
docker compose logs -f         # Follow logs
docker compose down            # Stop
```

The host port defaults to `8237`. Override with `HOST_PORT`:

```bash
HOST_PORT=9000 docker compose up -d
```

### Remote Server

```bash
./deploy.sh             # Build + transfer image + restart on server
./deploy.sh --rebuild   # Force full rebuild (no cache)
```

The script builds the image locally, saves it as a tarball, copies it via SCP, loads it on the remote host, and restarts the container. Requires Docker on the remote host and SSH key authentication. Copy `.server-config.example` to `.server-config` and fill in your server details before running.

---

## Project Structure

```
Web Form Server/
├── main.go               # HTTP server, route registration
├── form_loader.go        # JSON parsing, validation, struct definitions
├── renderer.go           # HTML rendering for all field types
├── submission.go         # Submission logging and JSON file output
├── config_handler.go     # /config route handlers
├── go.mod                # Go module definition
├── Dockerfile            # Multi-stage build (builder + alpine runtime)
├── docker-compose.yml    # Service definition with volume mounts
├── deploy.sh             # Build-and-deploy script for remote server
├── CLAUDE.md             # Claude Code project context
├── templates/
│   ├── form.html         # Form rendering template (multi + single page)
│   ├── submitted.html    # Post-submission confirmation page
│   └── config.html       # Visual config editor page
├── web-config/
│   ├── example.form.json # Demo form covering all field types
│   └── example.form.css  # Default stylesheet
└── data/
    └── submissions/      # Saved submission JSON files (volume-mounted)
```

## Development

Requires Go 1.22+.

```bash
go mod tidy                      # Fetch dependencies
go build -o web-form-server .    # Build binary
go vet ./...                     # Static analysis
./web-form-server                # Run (set FORM_FILE and CSS_FILE env vars)
```

Templates are embedded into the binary at compile time via `//go:embed`. Changes to template files require a rebuild.

Config and CSS files are read from disk at request time — edit them and refresh the browser with no rebuild needed.
