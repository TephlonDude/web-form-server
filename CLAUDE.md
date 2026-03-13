# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Docker-containerized web form server written in Go. Reads a TOML form definition file and a CSS file, then dynamically generates and serves a fully functional HTML form. No code changes needed to deploy a new form — edit two config files and restart the container.

## Technology Stack

- **Language:** Go 1.22
- **Web framework:** `net/http` (stdlib ServeMux with method-prefixed patterns)
- **Template engine:** `html/template` with custom FuncMap
- **Config format:** TOML via `github.com/BurntSushi/toml v1.4.0`
- **Container:** Multi-stage Docker build — `golang:1.22-alpine` builder → `alpine:3.19` runtime (~15 MB image)

## Essential Commands

### Docker Operations (Primary)
```bash
docker compose up -d --build      # Build and start container in background
docker compose down               # Stop and remove container
docker compose logs -f            # Follow container logs
docker compose ps                 # Check container status and health
./deploy.sh                       # Build locally + deploy to remote server
./deploy.sh --rebuild             # Force full rebuild (no cache) + deploy
```

### Local Development (requires Go 1.22+)
```bash
go mod tidy                       # Fetch/sync dependencies and generate go.sum
go build -o web-form-server .     # Build binary locally
go vet ./...                      # Static analysis
```

## Architecture

### File Structure
- `main.go` — HTTP server, route registration, CSS reader, `getEnv` helper
- `form_loader.go` — TOML struct definitions, `loadForm()`, `loadFormFromString()`, validation, defaults
- `renderer.go` — `renderField(f Field) template.HTML` dispatches all HTML input rendering
- `submission.go` — Submission record building, stdout logging, optional JSON file output
- `config_handler.go` — `/config` page, `/config/load`, `/config/save` route handlers
- `templates/form.html` — Multi-page and single-page form rendering with JS navigation
- `templates/submitted.html` — Post-submission confirmation page
- `templates/config.html` — Visual form builder + TOML editor admin page

### Routes
| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Render form from TOML config |
| `POST` | `/submit` | Handle form submission |
| `GET` | `/health` | Health check (used by Docker) |
| `GET` | `/config` | Visual form builder + TOML editor |
| `GET` | `/config/load` | Returns raw TOML file content |
| `POST` | `/config/save` | Validates and writes TOML file |

### Environment Variables
| Variable | Default | Description |
|---|---|---|
| `FORM_FILE` | `/web-config/form.toml` | TOML form definition path |
| `CSS_FILE` | `/web-config/form.css` | CSS stylesheet path |
| `PORT` | `5000` | HTTP listen port |
| `SAVE_SUBMISSIONS` | `false` | Write JSON files if `true` |
| `SUBMISSIONS_DIR` | `/data/submissions` | JSON output directory |

### Volume Mounts
- `./web-config:/web-config:rw` — Form TOML + CSS (editable via `/config` page without rebuilding)
- `./data/submissions:/data/submissions:rw` — Persisted submission JSON files

### Key Implementation Details
- Templates compiled into binary via `//go:embed templates`
- `Field.Min/Max/Step` are `interface{}` to handle both numeric and string TOML values
- `renderField()` returns `template.HTML` (not `string`) to prevent double-escaping
- `safeCSS` FuncMap wraps CSS in `template.CSS` to bypass context-aware escaping
- Multi-page forms use client-side JS with HTML5 `checkValidity()` for per-page validation
- Single-page forms use `[[form.sections]]`; multi-page use `[[form.pages]]` (mutually exclusive)

## Development Workflow

1. Edit `web-config/example.form.toml` or use the `/config` page
2. CSS changes: edit `web-config/example.form.css` — takes effect on next page load
3. Code changes: `docker compose up -d --build`
4. Test form at `http://localhost:8237`
5. Deploy to server: `./deploy.sh`

## Server Access & Credentials

### Stored Credentials
Server credentials are stored in `.server-config` (git-ignored). See that file for the server IP, username, and deploy directory.

### SSH Key Authentication
Passwordless SSH is configured using:
- **Key Location:** `~/.ssh/t5n_api_server`
- **SSH Alias:** `t5n-api-server`

**Quick SSH Access:**
```bash
ssh t5n-api-server
ssh -i ~/.ssh/t5n_api_server jason@192.168.202.206
```

### Deployment Script

Use `deploy.sh` for all deployments:
```bash
./deploy.sh             # Build + deploy (uses build cache)
./deploy.sh --rebuild   # Full rebuild from scratch + deploy
```

### Manual Server Commands
```bash
# SSH with key
ssh -i ~/.ssh/t5n_api_server jason@192.168.202.206 "command"

# SCP with key
scp -i ~/.ssh/t5n_api_server file.txt jason@192.168.202.206:/path/

# Check running container
ssh -i ~/.ssh/t5n_api_server jason@192.168.202.206 "cd ~/web-form-server && docker compose ps"

# Follow logs on server
ssh -i ~/.ssh/t5n_api_server jason@192.168.202.206 "cd ~/web-form-server && docker compose logs -f"
```

## Important Notes

- `go.sum` is git-ignored — generated at build time by `go mod tidy` in the Dockerfile
- `data/submissions/*.json` is git-ignored — submission data persists on the host only
- `.server-config` is git-ignored — never commit server credentials
- The `/config` save endpoint validates TOML before writing — invalid TOML is rejected with an error
- `web-config` volume is mounted `rw` so the config page can write updates to the TOML file
