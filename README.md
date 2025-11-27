# Personal Website & Blog (Go)

This is my very own, personal website and blog, built from the ground up with Go (Golang). My main goal here was to create a straightforward, self-hosted solution where I have absolute, 100% control over everything – from my content to how and where it's deployed.

## Features

*   **📝 Markdown Blog**: Write posts in Markdown with full rendering support (via `goldmark`).
*   **🔐 Admin Dashboard**: Secure login system to manage content.
*   **✏️ CRUD Operations**: Create, Read, Update, and Delete (soft delete) posts.
*   **📝 Draft System**: Save posts as drafts and publish them when ready.
*   **🖼️ Media Manager**: Upload and manage images directly from the dashboard.
*   **⚙️ Dynamic Settings**: Edit "About Me" and other site settings without code changes.
*   **🎨 Clean UI**: Minimalist, responsive design with a dark/light neutral theme.
*   **🚀 High Performance**: Built with the Go standard library and `modernc.org/sqlite` (pure Go SQLite, no CGO required).

## Tech Stack

*   **Language**: Go 1.22+
*   **Router**: Standard Library `net/http` (ServeMux)
*   **Database**: SQLite (embedded, pure Go)
*   **Templates**: Go `html/template`
*   **CSS**: Custom minimal CSS (Flexbox/Grid)

## Getting Started

### Prerequisites

*   Go 1.22 or higher installed.

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
│   ├── auth/           # Authentication logic (bcrypt)
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Auth middleware
│   ├── models/         # Data structures
│   └── repository/     # Database access (SQLite)
├── web/
│   ├── static/         # CSS, Images, Uploads
│   └── template/       # HTML Templates
├── data/               # SQLite database file (ignored by Git)
└── go.mod
```

## Development

*   **Run in Dev Mode**: Just use `go run ./cmd/server/`.
*   **Static Files**: CSS and images are served from `web/static/`.
*   **Templates**: HTML files are in `web/template/`. The app caches templates on startup, so restart the server to see changes (or modify `internal/handlers/app.go` to disable caching for dev).

## License

[MIT](LICENSE)

## Acknowledgements

This project was created with the help of Google Gemini.
