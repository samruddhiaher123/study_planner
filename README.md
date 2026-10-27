# Study Planner

Study Planner is a Go web application for managing study tasks. It uses Gorilla Mux for HTTP routing, PostgreSQL with GORM for persistence, bcrypt for password hashing, JWT bearer tokens for API authentication, and a static HTML/CSS/JavaScript frontend.

## Contents

- [Project structure](#project-structure)
- [Technology and prerequisites](#technology-and-prerequisites)
- [Configuration](#configuration)
- [Database setup and startup](#database-setup-and-startup)
- [Application architecture](#application-architecture)
- [Request lifecycle](#request-lifecycle)
- [API reference](#api-reference)
- [Task rules, filters, and sorting](#task-rules-filters-and-sorting)
- [Dashboard and completion history](#dashboard-and-completion-history)
- [Database models and migration behavior](#database-models-and-migration-behavior)
- [Frontend behavior](#frontend-behavior)
- [Swagger / OpenAPI](#swagger--openapi)
- [Development and verification](#development-and-verification)
- [Current scope and limitations](#current-scope-and-limitations)

## Project structure

```text
study_planner/
├── main.go                 # Startup and dependency wiring
├── config/                 # Environment configuration
├── database/               # PostgreSQL connection and GORM migration
├── controllers/            # HTTP input/output, validation, status mapping
├── repositories/           # GORM queries and persistence
├── services/               # Authentication, task, and dashboard logic
├── routes/                 # Gorilla Mux endpoints, Swagger, static files
├── middleware/             # JWT authentication and CORS
├── models/                 # Database entities and API data structures
├── utils/                  # JWT and JSON HTTP helpers
├── migrations/             # Standalone SQL files (not run at startup)
├── docs/                   # Generated Swagger/OpenAPI artifacts
├── frontend/               # Static web pages, JavaScript, and CSS
├── go.mod
└── go.sum
```

### Layer responsibilities

- **Controllers** decode HTTP requests, validate request/query data, call services, and produce JSON responses with the established HTTP status codes.
- **Services** implement application behavior such as account registration, password verification, task completion, and dashboard calculations.
- **Repositories** perform database reads and writes. Task queries and ownership restrictions are enforced here as well as by service use of the authenticated user ID.
- **Models** define GORM entities and API request/filter/response structures.
- **Routes** connect endpoints to controllers and install middleware.
- **Middleware** validates bearer JWTs, stores the authenticated user ID in request context, and handles CORS.
- **Config** reads environment variables. **Database** opens PostgreSQL and performs startup migration/backfill work.
- **Utils** contains shared JSON helpers and JWT functions.

The dependency flow is constructed in `main.go`:

```text
GORM database
  ├── User repository → Auth service → Auth controller
  └── Task repository → Task service → Task controller
                    └→ Dashboard service → Dashboard controller
Controllers → routes.New() → HTTP server
```

## Technology and prerequisites

- Go 1.22 or newer
- PostgreSQL
- Dependencies listed in `go.mod`

No separate frontend build step is required; Go serves the files from `frontend/`.

## Configuration

The application loads `.env` from the current working directory using `godotenv`. Environment variables already set in the process are also used. A `.env.example` is not currently included. Create a private `.env` or configure the same values in your shell/deployment environment:

```dotenv
# Set DATABASE_URL, or use the individual DB_* values below.
DB_HOST=localhost
DB_PORT=5432
DB_USER=study_planner_app
DB_PASSWORD=replace_with_a_local_password
DB_NAME=study_planner
DB_SSLMODE=disable

JWT_SECRET=replace_with_a_long_random_secret

# Optional overrides:
HTTP_ADDRESS=:8080
STATIC_DIR=./frontend
```

Configuration variables:

| Variable | Required | Default / behavior |
|---|---:|---|
| `DATABASE_URL` | One database configuration is required | If non-empty, it is used instead of the individual `DB_*` values. |
| `DB_HOST` | No | `localhost` |
| `DB_PORT` | No | `5432` |
| `DB_USER` | Yes, unless `DATABASE_URL` is set | PostgreSQL application role |
| `DB_PASSWORD` | Yes, unless `DATABASE_URL` is set | PostgreSQL application password |
| `DB_NAME` | Yes, unless `DATABASE_URL` is set | PostgreSQL database name |
| `DB_SSLMODE` | No | `disable` |
| `JWT_SECRET` | Yes | Used to sign and verify JWTs. The server exits if it is empty. |
| `HTTP_ADDRESS` | No | `:8080` |
| `STATIC_DIR` | No | `./frontend`, relative to the process working directory |

Keep `.env` private and never commit real credentials or signing secrets. Use a long, random JWT secret and a least-privilege PostgreSQL application account.

## Database setup and startup

Create the database and an application role using your PostgreSQL administrator account. For example:

```sql
CREATE DATABASE study_planner;
CREATE USER study_planner_app WITH PASSWORD 'replace_with_a_local_password';
GRANT CONNECT ON DATABASE study_planner TO study_planner_app;
```

Connect to `study_planner` as its owner/administrator and grant schema permissions needed for GORM's initial table creation and subsequent migrations:

```sql
GRANT USAGE, CREATE ON SCHEMA public TO study_planner_app;
```

The role used for startup must have sufficient ownership/alteration rights for the tables GORM migrates. After the tables exist, runtime data access also requires table and sequence privileges. Coordinate these grants with your database administrator and existing ownership model; do not grant broad superuser access to the application.

From the repository root:

```bash
go mod download
go run .
```

On successful startup the server listens on the configured `HTTP_ADDRESS`, by default `http://localhost:8080`. The root URL redirects to `/login.html`. Swagger UI is at `/swagger/index.html`.

Startup order:

1. Load `.env` and read application configuration.
2. Require `JWT_SECRET`.
3. Open and ping PostgreSQL.
4. Run `database.Migrate` to create/update the application models and apply legacy task data backfills.
5. Construct the GORM repositories, services, and controllers and inject them into the router.
6. Start the HTTP server.

The server exits with a logged error if configuration, database connection, or migration startup fails.

## Application architecture

The standard request path is:

```text
HTTP request
  → CORS middleware
  → Gorilla Mux route
  → JWT middleware (protected API routes only)
  → controller
  → service
  → repository
  → PostgreSQL through GORM
```

The response returns from the repository/service to the controller, which writes JSON and the appropriate HTTP status. `frontend/common.js` parses the response and reports non-success responses to the page.

### Authentication and authorization

- `POST /api/register` and `POST /api/login` are public.
- All task and dashboard endpoints are behind `middleware.Auth`.
- Protected requests must send the login JWT as a Bearer credential in the HTTP Authorization header.
- Tokens use HS256 and contain `user_id` with a 24-hour expiration.
- The middleware validates the token and places its user ID in request context.
- Task reads, updates, and deletes are scoped using both the task ID and authenticated user ID. Task listing and dashboard queries also use the authenticated user ID.
- Passwords are hashed and checked with bcrypt; user password hashes are excluded from JSON model output.

## Request lifecycle

For example, when the task page loads:

1. `frontend/tasks.js` calls `requestTasks("/tasks...")`.
2. `frontend/common.js` prefixes `/api`, attaches the saved bearer token from local storage, and issues `fetch()`.
3. Gorilla Mux matches `GET /api/tasks` and passes it through the protected API subrouter.
4. `middleware.Auth` validates the JWT and adds the authenticated user ID to the request context; invalid or missing tokens receive `401`.
5. `TaskHandler.GetTasks` parses/validates query filters and calls `TaskService.List`.
6. `TaskService` delegates to `TaskRepository.List`, which executes the parameterized GORM query scoped to that user.
7. The controller responds with a JSON task array and `X-Server-Date`.
8. The browser parses the response and renders the task list. If the response is not successful, it displays the returned error.

For JSON request bodies, `utils.DecodeJSON` limits the body to 1 MiB and requires exactly one valid JSON value. JSON errors use `{"error":"..."}`.

## API reference

All API paths are prefixed with `/api`. The frontend helper adds this prefix when it calls the API.

| Method | Path | Authentication | Purpose |
|---|---|---|---|
| `POST` | `/api/register` | Public | Create an account |
| `POST` | `/api/login` | Public | Verify credentials and issue a JWT |
| `GET` | `/api/tasks` | Bearer JWT | List the current user's tasks |
| `POST` | `/api/tasks` | Bearer JWT | Create a task |
| `GET` | `/api/tasks/{id}` | Bearer JWT | Get one of the current user's tasks |
| `PUT` | `/api/tasks/{id}` | Bearer JWT | Update a task's editable fields |
| `DELETE` | `/api/tasks/{id}` | Bearer JWT | Delete one of the current user's tasks |
| `PUT` | `/api/tasks/{id}/complete` | Bearer JWT | Mark a task completed |
| `GET` | `/api/dashboard` | Bearer JWT | Get dashboard progress and streak data |

Typical responses by endpoint:

| Endpoint | Success | Common errors |
|---|---:|---|
| `POST /api/register` | `201` | `400` invalid input or existing email; `413` oversized JSON; `500` server/database error |
| `POST /api/login` | `200` | `400` invalid JSON; `401` invalid credentials; `413` oversized JSON; `500` lookup/token error |
| `GET /api/tasks` | `200` | `400` invalid filter; `401` missing/invalid token; `500` database error |
| `POST /api/tasks` | `201` | `400` invalid JSON/task; `401` missing/invalid token; `413` oversized JSON; `500` database error |
| `GET /api/tasks/{id}` | `200` | `400` invalid ID; `401` missing/invalid token; `404` not found/not owned; `500` database error |
| `PUT /api/tasks/{id}` | `200` | `400` invalid ID/JSON/task; `401` missing/invalid token; `404` not found/not owned; `413` oversized JSON; `500` database error |
| `DELETE /api/tasks/{id}` | `200` | `400` invalid ID; `401` missing/invalid token; `404` not found/not owned; `500` database error |
| `PUT /api/tasks/{id}/complete` | `200` | `400` invalid ID; `401` missing/invalid token; `404` not found/not owned; `500` database error |
| `GET /api/dashboard` | `200` | `401` missing/invalid token; `500` database error |

### Register

`POST /api/register`

Request:

```json
{
  "name": "Alex",
  "email": "alex@example.com",
  "password": "study123"
}
```

The name and email are trimmed, and the email is converted to lowercase. Name and email are required; name is limited to 200 characters, email to 320 characters, and password length must be 6–72 bytes.

Success (`201`):

```json
{"message":"Registration successful"}
```

An existing email returns `400` with `{"error":"Email already exists"}`.

### Login

`POST /api/login`

Request:

```json
{
  "email": "alex@example.com",
  "password": "study123"
}
```

Success (`200`):

```json
{
  "message": "Login successful",
  "token": "<JWT>"
}
```

Invalid email/password returns `401` with a generic error. The frontend stores the token under `studyPlannerToken` in local storage and sends it on subsequent protected API requests.

### Create a task

`POST /api/tasks`

```json
{
  "title": "Review biology",
  "subject": "Biology",
  "description": "Revise cell structure",
  "date": "2026-10-27",
  "startTime": "09:00",
  "endTime": "10:00",
  "status": "Not Started",
  "priority": "Medium"
}
```

Success returns `201` and the created task as JSON. If the submitted status is `Completed`, the task receives a completion timestamp and a study-completion history record in the same database transaction.

### Get, update, delete, and complete tasks

- `GET /api/tasks/{id}` returns the matching user's task (`200`), or `404` if it does not exist or belongs to another user.
- `PUT /api/tasks/{id}` accepts the same editable fields as task creation and returns the updated task (`200`). The update request validates all editable task fields. An omitted or empty priority is treated as the default `Medium`.
- `DELETE /api/tasks/{id}` returns `200` with `{"message":"Task deleted"}`; a missing/non-owned task returns `404`.
- `PUT /api/tasks/{id}/complete` returns `200` with `{"message":"Task completed"}`. It records a completion timestamp/history entry when the task is newly completed; repeating the request for an already-completed task with a timestamp does not add another history entry.

Task IDs must be positive integers; an invalid ID returns `400`.

### Dashboard

`GET /api/dashboard` returns data for the authenticated user, including:

| JSON field | Meaning |
|---|---|
| `todayDate` | Server's current date (`YYYY-MM-DD`) |
| `todayProgress` | Today's completed tasks as a percentage of today's planned tasks |
| `todayPlanned`, `todayCompleted`, `todayRemaining` | Today's task counts |
| `totalTasks`, `completed`, `pending` | Overall task totals |
| `notStarted`, `inProgress`, `onHold` | Task counts by status |
| `week` | Seven entries from Monday through Sunday, each containing `day`, `date`, `completed`, `total`, and `progress` |
| `weeklyProgress` | Completed tasks as a percentage of tasks scheduled this week |
| `previousWeekProgress` | Previous week's completion percentage |
| `previousWeekAvailable` | Whether there is task data for the previous week |
| `weekChange` | This week's progress minus previous week's progress, in percentage points |
| `currentStreak`, `longestStreak` | Current and longest consecutive-day study streak |
| `lastStudyDate` | Latest recorded completion date, or an empty string when there is none |

Progress percentages are integer percentages. The current week begins on Monday. Streaks use persisted completion timestamps, not scheduled task dates.

## Task rules, filters, and sorting

### Task validation and defaults

- `title` and `subject` are required after trimming whitespace.
- Title is limited to 255 characters; subject is limited to 150 characters.
- `date` must use `YYYY-MM-DD`.
- `startTime` and `endTime` must use `HH:MM`; end time must be later than start time.
- Supported statuses are `Not Started`, `In Progress`, `Completed`, and `On Hold`.
- Empty status and legacy `Pending` input normalize to `Not Started`.
- Supported priorities are `High`, `Medium`, and `Low`. Missing/empty priority defaults to `Medium`.

### Task-list query parameters

`GET /api/tasks` accepts these optional parameters. All results remain scoped to the authenticated user.

| Parameter | Behavior |
|---|---|
| `q` | `LIKE` search against title or description. |
| `subject` | Exact subject match. |
| `status` | Filter by supported status; legacy `Pending` is treated as `Not Started`. |
| `priority` | Filter by `High`, `Medium`, or `Low`. |
| `date` | Exact scheduled date in `YYYY-MM-DD`. |
| `from` | Inclusive start of the date range in `YYYY-MM-DD`. |
| `to` | Inclusive end of the date range in `YYYY-MM-DD`. |
| `view` | `all` (default), `completed`, `pending`, `overdue`, or `upcoming`. |
| `sort` | `default`/`dueDate`, `priority`, `newest`, `oldest`, or `status`. |

`pending` includes all tasks whose status is not `Completed`. `overdue` means before today's server date and not completed; `upcoming` means today or later and not completed. If both `from` and `to` are supplied, `from` cannot be after `to`. Invalid values return `400`.

Default sorting puts incomplete tasks before completed tasks, overdue tasks before future tasks, then orders by priority, date, start time, and ID. The `priority` sort also places incomplete tasks first and then sorts by priority/date/time. `newest` and `oldest` sort by creation time; `status` sorts by status, date, and start time.

Successful task-list responses include `X-Server-Date` in `YYYY-MM-DD` format for the frontend's date display.

## Dashboard and completion history

`models.StudyCompletion` stores a user's task ID and completion timestamp. Creating a task already marked complete, transitioning a task to complete, or calling the completion endpoint records history alongside the task write in a transaction. The dashboard uses these records to compute the latest study date and consecutive-day streaks.

When an existing database has completed tasks with `completed_at` values but no matching history record, the startup migration inserts the missing history record. It does not invent completion timestamps.

## Database models and migration behavior

- `models.User` stores the user's name, unique email, bcrypt password hash, and creation timestamp. The password is not serialized to JSON.
- `models.Task` stores its owner ID, title, subject, description, scheduled date and times, status, priority, optional completion timestamp, and creation timestamp. The user relationship is not serialized.
- `models.StudyCompletion` stores owner ID, task ID, and completion time. Its user relationship is not serialized.
- The subject is currently a task string; there is no separate subject model in this application.

The schema is managed by the versioned SQL files in `migrations/`; the application does not apply schema changes at startup. `database.BackfillLegacyData` still runs a data-only transaction at startup:

1. Changes `NULL`, blank, or legacy `Pending` task statuses to `Not Started`.
2. Changes `NULL` or blank task priorities to `Medium`.
3. Copies completed tasks with a recorded completion timestamp into `study_completions` if the matching `(task_id, user_id, completed_at)` row is missing.

The tracked `000001_create_study_notes_table.*.sql` files are retained as the first migration. The `users`, `tasks`, and `study_completions` tables are defined by versions `000002` through `000004`. There is no separate `College` or `Subject` model/table in the current application; task subjects are stored as text in `tasks.subject`.

Install the `golang-migrate` CLI (the PostgreSQL driver tag is required):

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
```

Set `DATABASE_URL` to a PostgreSQL connection URL, for example:

```text
postgres://USER:PASSWORD@HOST:5432/DB_NAME?sslmode=disable
```

Alternatively, the application builds the URL from `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSLMODE`. Apply migrations to a new, empty database with:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Verify the schema using `\dt`, `\d users`, `\d tasks`, and `\d study_completions` in `psql`; confirm foreign keys, the unique email index, and the two completion-history indexes.

**Existing databases:** do not run `up` until you have backed up the database and checked the existing schema against versions `000002`–`000004`. If those model-backed tables already exist and match, establish the baseline without executing SQL using `migrate -path migrations -database "$DATABASE_URL" force 4`. This also records the existing, non-model `000001` migration as applied; only do so if that is acceptable for this database (for example, `study_notes` already exists or you have explicitly decided to treat that legacy migration as historical). `force` only records a version and bypasses migration execution; it does not validate or alter the schema. If the existing database has only some of the model-backed tables, do not force version `4`; reconcile its schema and migration history deliberately before applying later migrations.

To roll back exactly one migration, use `migrate -path migrations -database "$DATABASE_URL" down 1`. Rollback drops the table created by the latest migration and therefore deletes its data; back up first and use this only when that data may be removed. The application must not be started against a new database until the migrations have been applied.

Do not restore GORM `AutoMigrate` to normal startup after adopting these SQL migrations: that would create a second schema-management path. The startup data backfill normalizes legacy task values and inserts missing completion history; it does not create or alter tables.

## Frontend behavior

- `frontend/login.html` and `frontend/register.html` use `frontend/auth.js` for form submission and client-side email/password checks.
- `frontend/common.js` prefixes API paths with `/api`, attaches the saved token, parses JSON, exposes the server date, and provides logout/login-guard helpers.
- `frontend/tasks.html` and `frontend/tasks.js` display tasks, support search/filter/sort, create and edit tasks, update status, and delete tasks.
- `frontend/calendar.js` requests the visible six-week date range and lets the user open a task for editing.
- `frontend/dashboard.html` and `frontend/dashboard.js` render summary counts, weekly percentages, and streak information.
- The router serves static content from `STATIC_DIR` (`./frontend` by default). `/` redirects to `/login.html`.

The frontend is served by the Go application; its requests use relative `/api/...` URLs.

## Swagger / OpenAPI

Swagger UI is registered at `/swagger/`; with the default address, open:

```text
http://localhost:8080/swagger/index.html
```

The generated specification is available at `/swagger/doc.json`. Protected endpoints document Bearer authentication. After changing Swagger annotations in `main.go` or controller methods, regenerate the artifacts:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -o docs
```

## Development and verification

Format Go files and run the project checks:

```bash
gofmt -w $(find . -name '*.go' -not -path './.git/*')
go test ./...
go vet ./...
go build ./...
```

Focused tests cover task query validation/defaults, task ownership and completion behavior, authentication hashing/JWT issuance, dashboard calculations, and protected-route behavior. PostgreSQL-dependent endpoint workflows should be tested against a dedicated test database; avoid using a real user database for tests that create/update/delete records.

## Current scope and limitations

- Study tracking currently records task completion dates and streaks. It does not record actual study duration/minutes or provide a separate study-time endpoint.
- There is no separate subject CRUD API or subject table managed by this application.
- The legacy `study_notes` SQL migration runs when the full migration directory is applied with `golang-migrate`; it is not applied by application startup and has no corresponding model.
- CORS currently allows all origins (`Access-Control-Allow-Origin: *`), matching the existing middleware behavior. Review this policy if deploying the frontend/API across origins in a production environment.
