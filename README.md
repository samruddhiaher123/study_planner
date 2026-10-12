# Study Planner

Study Planner is a full-stack study scheduling application built with Go, Gorilla Mux, MySQL/GORM, JWT, and a static HTML/CSS/JavaScript frontend.

## Implemented features

### Accounts and API
- User registration and login with bcrypt password hashing
- JWT-protected API routes
- Per-user task ownership checks for list, read, update, delete, completion, and dashboard queries
- Swagger UI at `/swagger/index.html`

### Task planning
- Create, list, view, update, delete, and complete study tasks
- Task title, subject, description, scheduled date, start/end time, status, priority, and optional recorded study minutes
- Statuses: Not Started, In Progress, Completed, and On Hold
- Priorities: High, Medium, and Low
- Existing `Pending` status requests are accepted as `Not Started`; the existing completion route remains available
- Validated search, subject/status/priority filters, exact date and date-range filters, and active/completed/overdue/upcoming views
- Server-side sort choices with a stable default order: active overdue tasks first, then priority, scheduled date, start time, and ID; completed tasks appear after active tasks
- Optional subject grouping in the task list
- Monthly calendar with month navigation and task editing from calendar entries

### Dashboard and progress
- Today's planned, completed, and remaining task counts and completion percentage
- Total task and status counts
- Recorded actual study minutes (optional minutes entered on a task; included after the task is completed)
- Monday-to-Sunday task completion and study-time summary
- Weekly completion percentage and previous-week comparison when prior-week tasks exist
- Current and longest completion streak, plus the last study date

Today's completion percentage is the number of tasks scheduled for the server's current local date whose status is Completed, divided by the number planned for that date; it is zero when no tasks are planned. The weekly chart counts task completion by scheduled date. Recorded study minutes are grouped by the local date of the persisted completion timestamp. The previous-week comparison is shown only when at least one task was scheduled in the previous week.

Streaks are based on persisted completion history and the server's local calendar date (the same date returned by the API for overdue display). Completing at least one task on a local calendar day counts as a study day. The current streak continues through today or yesterday; a missed full day breaks it. New completions are stored independently of the current task status, so reopening or deleting a task does not erase its historical study day. Pre-existing completed rows without a timestamp are not assigned an invented date and therefore do not contribute to historical streaks.

Actual study time is opt-in: task durations are not treated as actual study time. Enter the number of minutes studied on the task and mark it completed for those minutes to contribute to daily and weekly study-time totals.

## Project structure

- `main.go` - environment setup, GORM migration, Gorilla Mux routes, and static frontend serving
- `config/` - MySQL connection setup
- `handlers/` - authentication, task, and dashboard HTTP handlers
- `middleware/` - JWT authentication middleware
- `models/` - GORM user, task, and per-user study-completion history models
- `utils/` - JWT helpers
- `frontend/` - login, registration, dashboard, task management, styles, and browser-side API calls
- `docs/` - generated Swagger/OpenAPI artifacts

## Setup

1. Create a MySQL database.
2. Set `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `JWT_SECRET` in your local `.env` or environment. Use your own credentials and a strong JWT secret; do not commit real secrets.
3. Start the application:

   ```bash
   go run .
   ```

4. Open the app at `http://localhost:8080/`. Swagger UI is available at `http://localhost:8080/swagger/index.html`.

The application uses GORM `AutoMigrate` to add the task priority, actual study minutes, completion timestamp, and study-completion history table without deleting task rows. Legacy `Pending` task statuses are migrated to `Not Started`; blank priorities are set to `Medium`. Existing completed records with a persisted completion timestamp are copied into completion history once on startup. For rows without a timestamp, the application preserves the task and does not fabricate a historical completion date.

## API overview

Public routes:

- `POST /api/register`
- `POST /api/login`

JWT-protected routes:

- `GET|POST /api/tasks`
- `GET|PUT|DELETE /api/tasks/{id}`
- `PUT /api/tasks/{id}/complete`
- `GET /api/dashboard`

`GET /api/tasks` supports `q`, `subject`, `status`, `priority`, `date`, `from`, `to`, `view`, and `sort` query parameters. Swagger documents the task fields, filters, dashboard response, and authentication requirements.

Regenerate Swagger artifacts after changing API annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -o docs
```

## Future work

Potential follow-up work includes password reset and email verification, reminders, recurring tasks, profile settings, production deployment/monitoring, rate limiting, and browser-to-database integration tests.
