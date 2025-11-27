# Project Specifications: Personal Website in Go

## Overview
A personal website and blog built with Go, featuring a public-facing portfolio/blog and a secure admin dashboard for content management. The design emphasizes simplicity and performance.

## Features

### Public Facing
1.  **Home Page**:
    -   **About Me**: Brief introduction.
    -   **Social Links**: GitHub, LinkedIn, etc.
    -   **Blog Feed**: List of recent posts including:
        -   Title
        -   Header/Summary
        -   Timestamp
    -   Clicking a post navigates to the full article.
2.  **Blog Post View**:
    -   Renders Markdown content as HTML.
    -   Clean, readable typography.

### Admin Dashboard
1.  **Authentication**: Secure login page for the site owner.
2.  **Content Management**:
    -   Interface to write/edit blog posts.
    -   Markdown editor support.
    -   CRUD operations (Create, Read, Update, Delete) for posts.

## Tech Stack
-   **Language**: Go (Golang)
-   **Web Framework/Router**: Go Standard Library (`net/http` with `ServeMux`)
-   **Database**: SQLite (Pure Go via `modernc.org/sqlite`)
-   **Templating**: Go standard `html/template`
-   **Styling**: Minimal CSS (custom or lightweight class-less framework)
-   **Markdown Engine**: `gomarkdown/markdown` or similar

## Project Structure
Adhering to [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```text
/
├── cmd/
│   └── server/         # Main application entry point
├── internal/           # Private application and library code
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   ├── repository/     # Database access layer
│   └── auth/           # Authentication logic
├── web/                # Web assets
│   ├── template/       # HTML templates
│   └── static/         # CSS, JS, Images
├── configs/            # Configuration files
├── migrations/         # Database migrations
├── data/               # SQLite database file
├── go.mod
└── README.md
```

## Development Log & Status
-   [x] **Initialization**: Project setup, `go.mod`, directory structure.
-   [x] **Database**: SQLite setup and schema migration.
-   [x] **Backend Core**: Basic HTTP server and routing.
-   [x] **Frontend**: Base templates and home page rendering.
-   [in_progress] **Admin System**: Auth middleware and login page.
-   [ ] **Blog System**: CRUD handlers and Markdown rendering.
