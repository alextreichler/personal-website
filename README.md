# Personal Website & Blog (Go)

This is my very own, personal website and blog, built from the ground up with Go (Golang). My main goal here was to create a straightforward, self-hosted solution where I have absolute, 100% control over everything – from my content to how and where it's deployed.

## Features

*   **📝 Markdown Blog**: Write posts in Markdown with full rendering support (via `goldmark`). Features syntax highlighting and HTML sanitization.
*   **🔐 Admin Dashboard**: Secure login system to manage content.
*   **✏️ CRUD Operations**: Create, Read, Update, and Delete (soft delete) posts.
*   **📝 Draft System**: Save posts as drafts and publish them when ready.
*   **🖼️ Media Manager**: Upload and manage images with automatic optimization.
*   **⚙️ Dynamic Settings**: Edit "About Me" and other site settings without code changes.
*   **📈 Metrics & Health**: Built-in Prometheus metrics and Kubernetes health checks.
*   **🎨 Clean UI**: Minimalist, responsive design with Dark/Light/Retro modes.
*   **🚀 High Performance**: Built with the Go standard library and `modernc.org/sqlite` (pure Go SQLite, no CGO required).

## Tech Stack

*   **Language**: Go 1.25+
*   **Router**: Standard Library `net/http` (ServeMux)
*   **Database**: SQLite (embedded, pure Go via `modernc.org/sqlite`)
*   **Templates**: Go `html/template`
*   **Markdown**: `goldmark` with syntax highlighting
*   **Security**: `bluemonday` for HTML sanitization and custom security middleware
*   **Monitoring**: `prometheus/client_golang`
*   **Image Processing**: `disintegration/imaging`
*   **CSS**: Custom minimal CSS (Flexbox/Grid)

## Getting Started

### Prerequisites

*   Go 1.25 or higher installed.
*   (Optional) [Task](https://taskfile.dev/) for easier development commands.

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/alextreichler/personal-website.git
    cd personal-website
    ```

2.  **Create an Admin User**:
    The database is created automatically on the first run. Use the CLI tool to create your admin account:
    ```bash
    go run cmd/admin/main.go -user admin -pass securepassword
    ```

3.  **Run the Server**:
    Using Task:
    ```bash
    task run
    ```
    Or using Go directly:
    ```bash
    go run ./cmd/server/
    ```

4.  **Visit the Site**:
    *   Public Site: [http://localhost:6060](http://localhost:6060)
    *   Admin Login: [http://localhost:6060/admin](http://localhost:6060/admin)

## Project Structure

```text
/
├── cmd/
│   ├── server/         # Main web server entry point
│   └── admin/          # CLI tool for user management
├── internal/           # Application code
│   ├── auth/           # Authentication and session logic
│   ├── config/         # Environment-based configuration
│   ├── handlers/       # HTTP handlers and template rendering
│   ├── middleware/     # Auth, Gzip, Security, Metrics, CSRF, ETag
│   ├── models/         # Data structures
│   └── repository/     # Database access and migrations
├── migrations/         # SQL migration files
├── web/
│   ├── static/         # CSS, JS, Favicon, Uploads
│   └── template/       # HTML Templates (Base + Pages)
├── data/               # SQLite database file (ignored by Git)
├── Taskfile.yaml       # Automation tasks
└── go.mod              # Dependencies
```

## Development

*   **Build**: `task build` builds the binary in the `bin/` directory.
*   **Clean**: `task clean` removes build artifacts.
*   **Docker**: `task image` builds a production-ready container image.
*   **Static Files**: CSS and images are served from `web/static/`.
*   **Templates**: The app caches templates on startup for performance.

## License

[MIT](LICENSE)

## Acknowledgements

This project was created with the help of Google Gemini.
