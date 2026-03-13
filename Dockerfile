# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy module definition and all Go source files
COPY go.mod *.go ./
# Copy templates (embedded into the binary via //go:embed)
COPY templates/ ./templates/

# Fetch dependencies and generate go.sum, then compile a fully static binary
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o web-form-server .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

# curl is required for the Docker healthcheck
RUN apk --no-cache add curl

# Non-root user for container security
RUN adduser -D -h /home/appuser appuser

WORKDIR /app
COPY --from=builder /build/web-form-server .

USER appuser

# Default environment variables (all overridable at runtime)
ENV FORM_FILE=/web-config/form.toml \
    CSS_FILE=/web-config/form.css \
    PORT=5000 \
    SAVE_SUBMISSIONS=false \
    SUBMISSIONS_DIR=/data/submissions

EXPOSE 5000

# Templates are embedded in the binary — no extra files needed at runtime
CMD ["./web-form-server"]
