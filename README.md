# appwrite-go-kit-template

Runnable Gin service showing `appwrite-go-kit` end to end: env config, optional
capabilities, guarded routes, database CRUD and storage uploads.

Targets self-hosted Appwrite 2.7.x. Requires Go 1.26.11 or newer.

## Clone both repositories side by side

```
somewhere/
├── appwrite-go-kit/            # the library
└── appwrite-go-kit-template/   # this repo
```

`go.mod` uses `replace github.com/e3tu/appwrite-base => ../appwrite-go-kit`.
Once the library is tagged and pushed, swap that line for a versioned require.

## Appwrite setup

1. Create a project in your self-hosted instance.
2. Create an API key with `databases`, `tables`, `rows` and `files` scopes.
3. Create a database, then a `todos` table inside it with columns:
   - `title` - `varchar`, size `255`, required
   - `desc` - `text`, not required
   - `completed` - `boolean`, not required
4. Optionally create a storage bucket, then set `APPWRITE_STORAGE_BUCKET`.

## Run

```
cp .env.example .env
# fill in project id, API key, endpoints, database id
go run .
```

## Routes

| Method | Path                       | Needs             |
|--------|----------------------------|-------------------|
| GET    | `/healthz`                 | nothing           |
| GET    | `/v1/capabilities`         | nothing           |
| GET    | `/v1/todos`                | db                |
| POST   | `/v1/todos`                | db + session      |
| PUT    | `/v1/todos/:id`            | db + session      |
| DELETE | `/v1/todos/:id`            | db + session      |
| GET    | `/v1/me`                   | auth + session    |
| POST   | `/v1/files`                | storage + session |
| GET    | `/v1/files/:id/url`        | storage           |

"session" means the `a_session_<project>` cookie from a signed-in Appwrite user.

## What to look at

- `internal/config/config.go` - the only file that reads env.
- `internal/router/router.go` - routes exist only for enabled capabilities.
- `internal/handlers/todos.go` - row CRUD with `query.*` helpers.
- `internal/handlers/files.go` - upload plus file URL built from `Endpoint`.
