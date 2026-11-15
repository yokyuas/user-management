# User Management API

REST API with Go, Gin, MongoDB and JWT HS256.

## Requirements
- Go 1.23+
- MongoDB 7+ (or Docker)
- JWT secret of at least 32 characters

## Run locally
```bash
cp .env.example .env
# Set a strong JWT_SECRET in .env
export $(grep -v '^#' .env | xargs)
go mod tidy
go run ./cmd/api
```

## Run with Docker
```bash
export JWT_SECRET='replace-with-a-long-random-secret-value'
docker compose up --build
```
API: `http://localhost:8080`; health: `GET /health`.

## Endpoints
- `POST /auth/register` (public)
- `POST /auth/login` (public)
- `POST /users` (JWT protected; creates user)
- `GET /users` (JWT protected)
- `GET /users/:id` (JWT protected)
- `PUT /users/:id` (JWT protected; name/email only)
- `DELETE /users/:id` (JWT protected)

Registration JSON:
```json
{"name":"Alice","email":"alice@example.com","password":"strong-password"}
```
Login JSON:
```json
{"email":"alice@example.com","password":"strong-password"}
```
Login response:
```json
{"token":"<JWT>","token_type":"Bearer"}
```
Use the token in protected requests:
```http
Authorization: Bearer <JWT>
```

## Tests
```bash
go test ./...
```
The service depends on a repository interface so it can be mocked in unit tests. Add tests with a fake repository or `testify/mock`.

## Decisions / assumptions
- MongoDB ObjectID is the auto-generated user ID.
- Email is normalized to lowercase and uniquely indexed.
- Passwords are bcrypt hashes; never included in JSON.
- Password length minimum is 8 characters.
- JWT uses HS256 and configurable expiration.
- List endpoint currently returns all users; production should add pagination.
- `/auth/register` and protected `POST /users` share user creation behavior.
- User count is logged every 10 seconds.

## Input Validation

- Register validates Name (Unicode letters/marks with spaces, apostrophes, or hyphens), Email format/length, and Password length (8–72 bytes).
- Login validates Email format and Password length (1–72 bytes); malformed credentials return the same generic invalid-credentials response.
- Update validates Name and Email with the same rules.
- Email is trimmed and normalized to lowercase. Password is not trimmed.
- Email uniqueness must also be enforced with a MongoDB unique index; validation alone is not concurrency-safe.
