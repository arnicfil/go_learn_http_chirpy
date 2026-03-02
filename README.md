# Chirpy API

A robust, lightweight backend HTTP server built in Go for "Chirpy", a Twitter/X-like clone. This project focuses on demonstrating fundamental concepts of building RESTful APIs in Go, including routing, middleware, database integration, and authentication.

## What it does
Chirpy API provides the complete backend service to power a microblogging application. It handles user management, authentication, and the core functionality of creating, reading, and deleting "chirps" (messages). It also features built-in administrative tools and metrics endpoints.

## How it works
The server is built using Go's standard `net/http` package and uses PostgreSQL for data persistence. 

### Prerequisites
1. Ensure you have [Go installed](https://go.dev/doc/install) (version 1.22+ recommended).
2. Install and configure a [PostgreSQL](https://www.postgresql.org/) database.

### Setup
1. Clone the repository and navigate into it:
   ```bash
   git clone https://github.com/arnicfil/go_learn_http_chirpy.git
   cd go_learn_http_chirpy
   ```
2. Set up your `.env` file with the following variables:
   ```env
   DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
   PLATFORM=dev
   SECRET=your_super_secret_jwt_string
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Execution
Run the server! By default, it runs on port 8080.
```bash
go run .
```
You can now access the endpoints at `http://localhost:8080/api/` (for example, a health check at `GET /api/healthz`).
