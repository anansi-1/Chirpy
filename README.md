# Chirpy API

Chirpy is a lightweight Go based API service for managing users, authentication, and "chirps" (short messages), with built-in metrics tracking and webhook support.  
It’s designed with simplicity in mind and uses PostgreSQL as the primary data store.

---

## Features

- **User Management** – Create, update, and authenticate users.
- **JWT Authentication** – Secure token-based login and refresh flow.
- **Chirp Management** – Create, retrieve, validate, and delete chirps.

- **Metrics & Admin Tools** – Track file server hits and reset state.
- **PostgreSQL Backend** – Managed via `internal/database` queries.

---

## Requirements

- **Go** 1.21 or later
- **PostgreSQL** 13+

---

## Environment Variables

The following must be set before running the service:

| Variable       | Description                                                |
|----------------|------------------------------------------------------------|
| `DB_URL`       | PostgreSQL connection string                               |
| `PLATFORM`     | Application platform identifier (string)                   | 
| `JWT_SECRET`   | Secret key for signing JWT tokens                          |

## Installation & Running
### 1. Clone the repository

``` 
https://github.com/anansi-1/Chirpy
cd Chirpy
```
### 2. Install dependencies

```
go mod tidy
```

### 3. Set up the database

- Make sure PostgreSQL is running.
- Create a database named `chirpy` (or update `DB_URL` in `.env` to match your setup).
- Run migrations with Goose:
     ```bash
     goose -dir sql/schema postgres "$DB_URL" up
     ```

 The server will start on: http://localhost:8080

