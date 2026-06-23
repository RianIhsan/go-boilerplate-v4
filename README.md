# Go Todo App — Clean Architecture

Golang REST API dengan Clean Architecture + Domain-Driven Design.

## Tech Stack

| Layer | Library |
|---|---|
| Router | [Chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL + `database/sql` (raw query) |
| Cache | Redis (`go-redis/v9`) |
| Auth | JWT (`golang-jwt/jwt/v5`) |
| Mocking | GoMock (`golang/mock`) |
| Logging | Zap (`uber-go/zap`) |
| Validation | `go-playground/validator` |

## Project Structure

```
go-todo-app/
├── cmd/api/            # Entrypoint — main.go only
├── config/             # Config loader dari env
├── migrations/         # SQL migration files
├── scripts/            # Helper scripts (mock gen, test runner)
├── pkg/                # Reusable packages (jwt, logger, crypto, pagination)
└── internal/
    ├── domain/
    │   ├── auth/       # Bounded context: Auth
    │   │   ├── handler/
    │   │   ├── usecase/
    │   │   ├── repository/
    │   │   ├── entity/
    │   │   └── dto/
    │   └── todo/       # Bounded context: Todo
    │       ├── handler/
    │       ├── usecase/
    │       ├── repository/
    │       ├── entity/
    │       └── dto/
    ├── infrastructure/
    │   ├── persistence/  # DB implementation of repositories
    │   ├── database/     # Postgres connection
    │   └── cache/        # Redis connection
    ├── mock/             # GoMock generated files
    └── shared/           # Cross-cutting: middleware, response, errors, validator
```

## Getting Started

### 1. Clone & setup env

```bash
cp .env.example .env
# Edit .env sesuai konfigurasi lokal
```

### 2. Jalankan dependencies

```bash
docker-compose up postgres redis -d
```

### 3. Run migrations

```bash
make migrate-up
```

### 4. Jalankan server

```bash
make run
```

Server berjalan di `http://localhost:8080`

---

## API Endpoints

### Auth

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Register user baru |
| POST | `/api/v1/auth/login` | Login, mendapat JWT token |

### Todo (🔒 JWT required)

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/todos` | Buat todo baru |
| GET | `/api/v1/todos` | List semua todo milik user |
| GET | `/api/v1/todos/{id}` | Detail satu todo |
| PUT | `/api/v1/todos/{id}` | Update todo |
| DELETE | `/api/v1/todos/{id}` | Hapus todo |

### Query Params (GET /todos)

| Param | Default | Description |
|---|---|---|
| `page` | 1 | Halaman |
| `limit` | 10 | Jumlah item per halaman |

---

## Example Requests

### Register
```json
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

### Login
```json
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "password123"
}
```

### Create Todo
```
POST /api/v1/todos
Authorization: Bearer <access_token>

{
  "title": "Beli groceries",
  "description": "Susu, Telur, Roti"
}
```

### Update Todo
```json
PUT /api/v1/todos/{id}
{
  "title": "Updated title",
  "status": "in_progress"
}
```

---

## Testing

```bash
# Jalankan semua unit test
make test

# Dengan coverage HTML
make test-coverage

# Atau langsung via script
bash scripts/run_test.sh
```

## Regenerate Mocks

```bash
make mock
# atau
bash scripts/generate_mock.sh
```

---

## Design Principles

- **Clean Architecture** — dependency hanya ke dalam (handler → usecase → repository)
- **Domain-Driven Design** — tiap domain punya folder sendiri: handler, usecase, repository, entity, dto
- **Dependency Inversion** — repository dan usecase berbasis interface, bukan concrete struct
- **Table-Driven Testing** — semua unit test menggunakan pola `[]struct{ name, input, want }`
- **Raw SQL** — tidak ada ORM, query transparan dan bisa di-optimize langsung
