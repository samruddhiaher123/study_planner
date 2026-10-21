# Study Planner

Study Planner is a full-stack study scheduling app built with Go, Gorilla Mux, PostgreSQL, GORM, JWT, bcrypt, and a static HTML/CSS/JavaScript frontend.

## Project structure

- `main.go` - startup, migrations, routes, and static frontend serving
- `config/` - PostgreSQL connection and safe GORM migrations
- `handlers/` - authentication, task, dashboard, and JSON HTTP handlers
- `middleware/` - JWT authentication
- `models/` - GORM database models and API structures
- `utils/` - JWT helpers
- `frontend/` - static pages, styles, and browser-side API calls
- `docs/` - generated Swagger/OpenAPI artifacts

## Configure PostgreSQL

PostgreSQL should be running, and the `study_planner` database should exist. Create a regular application account rather than using the PostgreSQL superuser:

```sql
CREATE USER study_planner_app WITH PASSWORD 'choose_a_local_password';
GRANT CONNECT ON DATABASE study_planner TO study_planner_app;
```

Connect to the `study_planner` database as its owner or administrator and grant schema access:

```sql
GRANT USAGE, CREATE ON SCHEMA public TO study_planner_app;
```

The application needs to create and update its own `users`, `tasks`, and `study_completions` tables. GORM `AutoMigrate` only adds missing columns/tables and does not drop existing columns or tables. Other existing relations (including `college` and `subjects`) are not migrated by this app. The role used at startup must own the existing `users` table (or an administrator must perform the initial migration), because PostgreSQL does not grant `ALTER TABLE` through ordinary table grants.

After creating the app tables, an administrator can grant normal runtime access:

```sql
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO study_planner_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO study_planner_app;
```

1. Copy `.env.example` to `.env`.
2. Set `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, and `JWT_SECRET`. Keep `.env` private and do not commit it.
3. Start the server:

   ```bash
   go mod tidy
   go run .
   ```

For a local PostgreSQL service on Ubuntu, check it with:

```bash
sudo systemctl status postgresql
pg_isready
```

Start it if needed with `sudo systemctl start postgresql`. Use the local PostgreSQL username/password you configured for `study_planner_app`; the application does not use or need the `postgres` superuser account.

Open `http://localhost:8080/`. Swagger UI is at `http://localhost:8080/swagger/index.html`.

## Existing schema and data

The app leaves unrelated tables and columns intact. Existing `users` columns such as `id`, `name`, `email`, and `age` are preserved; GORM adds the application fields it needs (including password and timestamps) without deleting `age`. Existing records are not overwritten.

Rows in `users` that predate the password column will not have a usable password and cannot log in until a password is set through a deliberate account recovery process. The app does not invent passwords or assign one to existing users.

Startup also converts legacy task status `Pending` to `Not Started`, sets blank priorities to `Medium`, and copies completed tasks with recorded completion timestamps into history if the matching history row does not exist. It does not invent completion timestamps.

Before first startup against a database containing a `users` table, review that table's columns, primary key, email uniqueness, and existing email duplicates. If emails are duplicated, resolve them deliberately before the unique email index can be added; this application will not delete or merge user rows. Ensure the database account owns or has `USAGE` and `CREATE` privileges on the schema.

## Features

- Registration and login with bcrypt password hashing
- JWT-protected API routes and per-user task ownership
- Task create, list, view, update, delete, and completion
- Task status, priority, search, filters, sorting, and date ranges
- Monthly calendar and task editing
- Daily dashboard and weekly completion/study-time summaries
- Current and longest study streaks based on persisted completion history
- Swagger UI and generated OpenAPI documentation

Actual study time is opt-in. Planned task duration is never counted as actual study time. Historical streak dates come only from persisted completion timestamps.

## API routes

Public:

- `POST /api/register`
- `POST /api/login`

JWT-protected:

- `GET|POST /api/tasks`
- `GET|PUT|DELETE /api/tasks/{id}`
- `PUT /api/tasks/{id}/complete`
- `GET /api/dashboard`

`GET /api/tasks` accepts `q`, `subject`, `status`, `priority`, `date`, `from`, `to`, `view`, and `sort`.

## Swagger

Regenerate Swagger artifacts after changing API annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -o docs
```
