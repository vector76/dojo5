# Implementation Plan

High-level plan for building the Dojo/Yoga Studio CRM described in [PRD.md](PRD.md).

**Approach**: All backend API phases (2–6) are completed before frontend phases (7–10). This means nothing is visually demonstrable until Phase 7, but it ensures the full API surface is stable before building the UI against it.

## Phase 1: Project Scaffolding & Build Pipeline

Set up the monorepo structure and build tooling so that every subsequent phase starts from a working, testable baseline.

- **Backend**: Initialize Go module (`backend/go.mod`), choose an HTTP router (e.g., `chi` or standard library `net/http`), create `cmd/server/main.go` with a health-check endpoint
- **Frontend**: Scaffold Vite + React + TypeScript app in `frontend/`, configure Vitest + React Testing Library
- **Embedding**: Wire up `//go:embed` so the Go binary serves the built React app
- **Makefile**: Targets for `build`, `test`, `dev` (concurrent frontend dev server + Go server with proxy)
- **Smoke tests**: Go test hitting health endpoint; frontend renders without errors

**Exit criteria**: `make build` produces a single binary that serves the React app and responds to `/api/health`. `make test` runs both Go and frontend test suites.

## Phase 2: Database & Data Layer

Stand up SQLite with a migration strategy and implement the core data models.

- **SQLite integration**: Choose a Go SQLite driver (e.g., `modernc.org/sqlite`), open/create DB file on startup
- **Migrations**: Simple numbered-SQL-file migration runner (no heavy ORM)
- **Models & repositories**: Go structs + CRUD functions for `users`, `class_types`, `classes`, `attendance`, `payments`
- **Schema design**:
  - `users` — id, name, email (unique), phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance, deleted_at (soft delete), timestamps
  - `class_types` — id, name, description, timestamps
  - `classes` — id, class_type_id, instructor_id, start_time, duration_minutes, capacity, timestamps
  - `attendance` — id, class_id, user_id, checked_in_at (append-only, no updated_at needed)
  - `payments` — id, user_id, amount, date, note, recorded_by, timestamps
- **Tests**: Table-driven Go tests for every repository function; test against an in-memory or temp-file SQLite DB

**Exit criteria**: All CRUD operations pass tests. Migration runner creates schema from scratch on empty DB.

## Phase 3: Authentication & Authorization

Implement login, JWT sessions, role-based access, and the first-run admin setup.

- **Password hashing**: bcrypt for storing/verifying passwords
- **JWT middleware**: Issue token on login; middleware validates token, injects user context
- **Soft-delete guard**: Login must reject soft-deleted users (`deleted_at IS NULL` check)
- **Role enforcement middleware**: Check `user.Role` against per-route permission requirements
- **API endpoints**:
  - `POST /api/auth/login` — returns JWT
  - `GET  /api/auth/me` — returns current user profile
  - `GET  /api/auth/setup-status` — reports whether initial admin exists
  - `POST /api/auth/setup` — create first admin account (only works when no users exist)
- **Tests**: Verify login success/failure, token validation, role-based access denied/allowed, first-run setup guard

**Exit criteria**: Auth round-trip works end-to-end in tests. Unauthorized and forbidden requests are rejected correctly per the role-permission matrix.

## Phase 4: User & Member Management API

Build out the admin-facing user CRUD and self-service profile endpoints.

- **API endpoints**:
  - `GET    /api/users` — list users (admin, instructor); excludes soft-deleted by default
  - `POST   /api/users` — create user (admin)
  - `GET    /api/users/:id` — get user detail (admin, instructor, or self)
  - `PUT    /api/users/:id` — update user (admin or self)
  - `DELETE /api/users/:id` — soft-delete user (admin); sets deleted_at, preserves attendance and payment history
  - `PUT    /api/users/:id/role` — change role (admin)
  - `PUT    /api/users/:id/password` — admin reset password
- **Validation**: Unique email, required fields, role enum
- **Tests**: CRUD operations, permission checks, validation errors

**Exit criteria**: Full user lifecycle works through API tests with proper role enforcement.

## Phase 5: Class & Attendance API

Endpoints for class type definitions, scheduling, and attendance tracking.

- **API endpoints**:
  - `CRUD  /api/class-types` — manage class types (admin)
  - `CRUD  /api/classes` — schedule classes (admin); list classes (all authenticated)
  - `POST  /api/classes/:id/attendance` — record attendance (admin, instructor)
  - `GET   /api/classes/:id/attendance` — list attendance for a class (admin, instructor)
  - `GET   /api/users/:id/attendance` — attendance history for a user (self, admin, instructor)
- **Tests**: Scheduling, capacity, attendance recording, permission checks

**Exit criteria**: Classes can be created, scheduled, and attendance recorded — all verified by tests.

## Phase 6: Payments API

Simple payment recording and history.

- **API endpoints**:
  - `POST /api/payments` — record payment (admin)
  - `GET  /api/users/:id/payments` — payment history (admin or self with user role; instructors are staff and don't have payment records)
  - `GET  /api/users/:id/balance` — outstanding balance indicator (same permission as payment history)
  - `PUT  /api/users/:id/balance` — manually set expected balance (admin)
- **Balance model**: There is no automated dues schedule. Admin manually sets each member's expected balance; payments reduce it. The balance endpoint returns the difference.
- **Tests**: Record payment, view history, balance calculation, permission checks

**Exit criteria**: Payments can be recorded and queried with correct role enforcement.

## Phase 7: Frontend — Layout, Routing & Auth UI

Build the application shell and authentication screens.

- **Router**: React Router with protected route wrapper that checks JWT
- **Layout**: Responsive shell — nav sidebar/header (role-aware: hide links the current user cannot access), content area
- **Pages**: Login, First-run Setup, Dashboard (placeholder)
- **Auth state**: Context/hook managing JWT in localStorage, login/logout, redirect on 401
- **API client**: Thin fetch wrapper that attaches JWT header
- **Tests**: Login form submission, redirect on unauthenticated, setup flow

**Exit criteria**: User can log in, see a dashboard shell, and be redirected when unauthenticated — all covered by component tests.

## Phase 8: Frontend — Member Management UI

Admin screens for managing users/members.

- **Pages**: Member list (with search/filter), Member detail/edit form (includes set expected balance and password reset), Add member form
- **Role-aware UI**: Hide admin-only actions for non-admin users; self-service profile edit
- **Tests**: Render list, form validation, role-based visibility

**Exit criteria**: Admin can create, view, edit, and delete members through the UI.

## Phase 9: Frontend — Class & Attendance UI

Screens for class management and attendance tracking.

- **Pages**: Class type management, Class schedule (list view sorted by date), Attendance check-in screen
- **User view**: "My attendance" history page
- **Tests**: Schedule display, attendance check-in flow, user self-view

**Exit criteria**: Admin can manage classes, instructors can record attendance, users can view their own history.

## Phase 10: Frontend — Payments UI

Payment recording and history screens.

- **Pages**: Record payment form, Payment history per member, Balance summary
- **User view**: "My payments" history page
- **Tests**: Payment form, history display, balance indicator

**Exit criteria**: Admin can record payments; users can view their own payment history.

## Phase 11: Polish & Release Prep

Final integration, edge cases, and deployment readiness.

- **End-to-end review**: Walk through every user flow for each role
- **Error handling**: Consistent API error responses, frontend error boundaries and toast notifications
- **Responsive check**: Verify mobile layouts
- **Build verification**: Clean `make build` produces working single binary
- **Documentation**: Update README with build/run/deploy instructions

**Exit criteria**: Single binary runs, serves the app, and all test suites pass. README documents how to build and run.
