---
name: auth-refactor
goal: Replace legacy session-based auth with JWT tokens
status: active
session_count: 1
created: "2026-03-16"
updated: "2026-03-16"
current_task: 0
tasks:
    - text: Define JWT token schema and signing config
      status: todo
      spec: Choose RS256, set expiry to 1h for access, 7d for refresh
      verify:
        - tokens validate with jose library
      files:
        - internal/auth/jwt.go
    - text: Build token issuing endpoint
      status: todo
      spec: POST /auth/token accepts credentials, returns access+refresh pair
      verify:
        - returns 200 with valid JWT
        - refresh token rotates on use
      files:
        - internal/auth/handler.go
        - internal/auth/handler_test.go
    - text: Add middleware to validate tokens on protected routes
      status: todo
      files:
        - internal/middleware/auth.go
    - text: Remove legacy session code
      status: todo
context:
    current_file: null
    last_error: null
    test_state: null
    open_questions: null
    pending_refactor: null
decisions: []
---

<!-- generated below — do not edit, use trail commands -->

## goal

Replace legacy session-based auth with JWT tokens

## tasks

- [ ] 00 · Define JWT token schema and signing config
- [ ] 01 · Build token issuing endpoint
- [ ] 02 · Add middleware to validate tokens on protected routes
- [ ] 03 · Remove legacy session code

## context

| field | value |
|---|---|
| current_file | ~ |
| last_error | ~ |
| test_state | ~ |
| open_questions | ~ |
| pending_refactor | ~ |

## decisions

No decisions yet.

## notes
