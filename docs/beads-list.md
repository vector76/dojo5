# Beads List — Dojo/Yoga Studio CRM

Breakdown of `implementation-plan.md` into individual trackable work items.

**Conventions:**
- `[Blocked by: #N]` indicates a dependency on another bead
- Each bead should be completable and testable independently (given its dependencies are met)
- Beads within a phase are ordered by natural implementation sequence
- TDD: tests are written as part of each bead, not as separate beads
- Numbering is not contiguous (#5 was merged into #4); IDs are kept stable to avoid churn

---

## Phase 1: Project Scaffolding & Build Pipeline

### #1 — Initialize Go backend with health endpoint
Set up `backend/go.mod`, choose HTTP router (chi or stdlib), create `cmd/server/main.go` with a `GET /api/health` endpoint. Include a Go test that hits the health endpoint and asserts a 200 response.

**Estimate: 1 hour**

### #2 — Scaffold frontend (Vite + React + TypeScript)
Initialize `frontend/` with Vite + React + TypeScript. Configure Vitest and React Testing Library. Add a minimal smoke test that the app renders without errors.

**Estimate: 1 hour**

### #3 — Embed frontend in Go binary
Wire up `//go:embed` so the Go binary serves the built React app's static files. The Go server should serve the API under `/api/` and the frontend for all other routes (SPA fallback).

`[Blocked by: #1, #2]`

**Estimate: 1.5 hours**

### #4 — Makefile with build, test, and dev targets
Create Makefile with:
- `make build` — builds frontend, then builds Go binary with embedded assets
- `make test` — runs both Go and frontend test suites
- `make dev` — runs concurrent frontend dev server + Go server with proxy

Verify exit criteria: `make build` produces a single binary that serves the React app and responds to `/api/health`. `make test` runs both suites successfully. Fix any integration issues.

`[Blocked by: #3]`

**Estimate: 1.5 hours**

---

## Phase 2: Database & Data Layer

### #6 — SQLite integration and migration runner
Add Go SQLite driver (`modernc.org/sqlite`). Implement DB open/create on startup. Build a simple numbered-SQL-file migration runner (read `.sql` files in order, track applied migrations). Include tests for the migration runner itself (applies migrations, skips already-applied, runs in order).

`[Blocked by: #1]`

**Estimate: 2.5 hours**

### #7 — Schema migrations: users table
Write SQL migration for the `users` table: id, name, email (unique), phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance, deleted_at (soft delete), created_at, updated_at.

`[Blocked by: #6]`

**Estimate: 30 minutes**

### #8 — Schema migrations: class_types, classes, attendance, payments tables
Write SQL migrations for remaining tables:
- `class_types` — id, name, description, timestamps
- `classes` — id, class_type_id (FK), instructor_id (FK), start_time, duration_minutes, capacity, timestamps
- `attendance` — id, class_id (FK), user_id (FK), checked_in_at
- `payments` — id, user_id (FK), amount, date, note, recorded_by (FK), timestamps

`[Blocked by: #7]`

**Estimate: 45 minutes**

### #9 — Users repository (CRUD)
Implement Go structs and repository functions for `users`: Create, GetByID, GetByEmail, List (excluding soft-deleted by default), Update, SoftDelete. Table-driven tests for all operations against a temp SQLite DB.

`[Blocked by: #7]`

**Estimate: 3 hours**

### #10 — Class types and classes repositories (CRUD)
Implement repository functions for `class_types` (Create, GetByID, List, Update, Delete) and `classes` (Create, GetByID, List with filters, Update, Delete). Table-driven tests for all operations.

`[Blocked by: #8]`

**Estimate: 3 hours**

### #11 — Attendance repository
Implement repository functions for `attendance`: RecordAttendance, ListByClass, ListByUser. Tests for recording attendance, querying by class, querying by user.

`[Blocked by: #8]`

**Estimate: 1.5 hours**

### #12 — Payments repository
Implement repository functions for `payments`: RecordPayment, ListByUser. Also implement balance calculation logic (expected_balance from user minus sum of payments). Tests for all operations including balance computation.

`[Blocked by: #8, #9]`

**Estimate: 2 hours**

---

## Phase 3: Authentication & Authorization

### #13 — Password hashing utilities
Implement bcrypt-based password hashing and verification functions. Tests for hash generation and verification (correct password succeeds, wrong password fails).

`[Blocked by: #1]`

**Estimate: 30 minutes**

### #14 — JWT token issuance and validation
Implement JWT token creation (with user ID, role, expiration) and validation middleware that extracts user context from the token. Tests for token generation, validation, expiration, and tampered tokens.

`[Blocked by: #1]`

**Estimate: 2.5 hours**

### #15 — Login endpoint
Implement `POST /api/auth/login` — accepts email/password, verifies credentials, rejects soft-deleted users (`deleted_at IS NULL`), returns JWT. Tests for successful login, wrong password, nonexistent user, and soft-deleted user.

`[Blocked by: #9, #13, #14]`

**Estimate: 2 hours**

### #16 — Auth me and setup endpoints
Implement:
- `GET /api/auth/me` — returns current user profile from JWT context
- `GET /api/auth/setup-status` — returns whether any admin user exists
- `POST /api/auth/setup` — creates first admin account (only when no users exist)

Tests for each endpoint, including the setup guard (setup must fail when users already exist).

`[Blocked by: #9, #13, #14]`

**Estimate: 2.5 hours**

### #17 — Role enforcement middleware
Implement middleware that checks `user.Role` from JWT context against per-route permission requirements. Should support requiring specific roles (e.g., admin-only, admin+instructor, any authenticated). Tests verifying access denied/allowed for each role level.

`[Blocked by: #14]`

**Estimate: 1.5 hours**

---

## Phase 4: User & Member Management API

### #18 — List and create user endpoints
Implement:
- `GET /api/users` — list users (admin, instructor); excludes soft-deleted by default
- `POST /api/users` — create user (admin only)

Validation: unique email, required fields (name, email, phone), valid role enum. Tests for CRUD, permission checks, and validation errors.

`[Blocked by: #9, #13, #17]`

**Estimate: 2.5 hours**

### #19 — Get, update, and delete user endpoints
Implement:
- `GET /api/users/:id` — get user detail (admin, instructor, or self)
- `PUT /api/users/:id` — update user (admin or self, with field restrictions for self)
- `DELETE /api/users/:id` — soft-delete user (admin only)

Tests for each operation with role-based permission checks.

`[Blocked by: #18]`

**Estimate: 2.5 hours**

### #20 — Role change and password reset endpoints
Implement:
- `PUT /api/users/:id/role` — change user role (admin only)
- `PUT /api/users/:id/password` — admin reset password (admin only)

Tests for permission enforcement and edge cases (can't demote last admin, etc.).

`[Blocked by: #19]`

**Estimate: 1.5 hours**

---

## Phase 5: Class & Attendance API

### #21 — Class types CRUD endpoints
Implement full CRUD for `CRUD /api/class-types` (admin only for create/update/delete; all authenticated for list/get). Tests for all operations with role enforcement.

`[Blocked by: #10, #17]`

**Estimate: 2 hours**

### #22 — Classes CRUD endpoints
Implement:
- `POST /api/classes` — schedule a class (admin only)
- `GET /api/classes` — list classes (all authenticated), with date filtering
- `GET /api/classes/:id` — get class detail (all authenticated)
- `PUT /api/classes/:id` — update class (admin only)
- `DELETE /api/classes/:id` — delete class (admin only)

Tests for scheduling, listing, permission checks.

`[Blocked by: #21]`

**Estimate: 3 hours**

### #23 — Attendance endpoints
Implement:
- `POST /api/classes/:id/attendance` — record attendance (admin, instructor)
- `GET /api/classes/:id/attendance` — list attendance for a class (admin, instructor)
- `GET /api/users/:id/attendance` — attendance history for a user (self, admin, instructor)

Tests for recording, querying, permission enforcement.

`[Blocked by: #11, #22]`

**Estimate: 2.5 hours**

---

## Phase 6: Payments API

### #24 — Payment recording and history endpoints
Implement:
- `POST /api/payments` — record payment (admin only)
- `GET /api/users/:id/payments` — payment history (admin, or self for user role)

Tests for recording, history retrieval, permission checks.

`[Blocked by: #12, #17]`

**Estimate: 2 hours**

### #25 — Balance endpoints
Implement:
- `GET /api/users/:id/balance` — outstanding balance (expected_balance minus sum of payments; same permissions as payment history)
- `PUT /api/users/:id/balance` — manually set expected balance (admin only)

Tests for balance calculation, manual adjustment, permission checks.

`[Blocked by: #24]`

**Estimate: 1.5 hours**

---

## Phase 7: Frontend — Layout, Routing & Auth UI

### #26 — API client and auth state management
Create a thin fetch wrapper that attaches JWT from localStorage. Implement auth context/hook: login, logout, current user state, auto-redirect on 401. Tests for the API client (header attachment, 401 handling) and auth hook.

`[Blocked by: #2]`

**Estimate: 2.5 hours**

### #27 — App layout and routing shell
Set up React Router with protected route wrapper that checks JWT. Build responsive layout: nav sidebar/header (role-aware — hide links the user can't access), content area. Include a placeholder Dashboard page as the default authenticated landing route. Tests for route protection (redirect when unauthenticated).

`[Blocked by: #26]`

**Estimate: 3 hours**

### #28 — Login and first-run setup pages
Build Login page (email/password form, error display, redirect on success). Build First-run Setup page (create initial admin form, shown when setup-status indicates no admin). Tests for form submission, error handling, setup flow.

`[Blocked by: #27]`

**Estimate: 3 hours**

---

## Phase 8: Frontend — Member Management UI

### #29 — Member list page
Build member list page with search and filter (by role, membership status). Role-aware: only visible to admin and instructor. Tests for rendering list, search/filter behavior, role-based visibility.

`[Blocked by: #27]`

**Estimate: 3 hours**

### #30 — Member detail/edit and add member pages
Build:
- Member detail view with edit form (includes set expected balance and admin password reset)
- Add member form (admin only)
- Self-service profile edit (for any role editing their own profile)

Role-aware: hide admin-only actions for non-admins. Tests for form validation, create/edit flows, role-based field visibility.

`[Blocked by: #29]`

**Estimate: 5 hours**

---

## Phase 9: Frontend — Class & Attendance UI

### #31 — Class type management page
Build admin page for managing class types (list, create, edit, delete). Tests for CRUD operations in the UI.

`[Blocked by: #27]`

**Estimate: 2.5 hours**

### #32 — Class schedule page
Build class schedule page: list view sorted by date, with date range filtering. Admin can create/edit/delete classes. All authenticated users can view. Tests for schedule display and admin actions.

`[Blocked by: #31]`

**Estimate: 3.5 hours**

### #33 — Attendance check-in and history pages
Build:
- Attendance check-in screen (admin/instructor selects a class, checks in members)
- "My attendance" history page (any authenticated user views their own history)

Tests for check-in flow, self-view history.

`[Blocked by: #32]`

**Estimate: 3.5 hours**

---

## Phase 10: Frontend — Payments UI

### #34 — Record payment and payment history pages
Build:
- Record payment form (admin selects member, enters amount/date/note)
- Payment history per member (admin view)
- "My payments" history page (user self-view)

Tests for payment form, history display.

`[Blocked by: #27]`

**Estimate: 3 hours**

### #35 — Balance summary display
Build balance summary: show expected balance, total paid, and outstanding amount on member detail and self-view pages. Tests for balance indicator rendering and edge cases (zero balance, overpayment).

`[Blocked by: #30, #34]`

**Estimate: 1.5 hours**

---

## Phase 11: Polish & Release Prep

### #36 — Error handling and UI polish
Implement consistent API error responses (standardized error JSON format). Add frontend error boundaries and toast notifications. Review and fix edge cases across all flows.

`[Blocked by: #4, #15, #16, #20, #23, #25, #28, #33, #35]`

**Estimate: 5 hours**

### #37 — Responsive design verification
Verify and fix mobile layouts across all pages. Ensure navigation works on small screens (collapsible sidebar/hamburger menu).

`[Blocked by: #36]`

**Estimate: 3 hours**

### #38 — Final build verification and documentation
Clean `make build` produces a working single binary. Walk through every user flow for each role. Update README with build, run, and deploy instructions. Ensure all test suites pass.

`[Blocked by: #37]`

**Estimate: 2 hours**

---

## Summary

| Phase | Beads | IDs | Estimate |
|-------|-------|-----|----------|
| 1 — Scaffolding | 4 | #1–#4 | 5 hrs |
| 2 — Database | 7 | #6–#12 | 13.25 hrs |
| 3 — Auth | 5 | #13–#17 | 9 hrs |
| 4 — User API | 3 | #18–#20 | 6.5 hrs |
| 5 — Class API | 3 | #21–#23 | 7.5 hrs |
| 6 — Payments API | 2 | #24–#25 | 3.5 hrs |
| 7 — Frontend Auth | 3 | #26–#28 | 8.5 hrs |
| 8 — Frontend Members | 2 | #29–#30 | 8 hrs |
| 9 — Frontend Classes | 3 | #31–#33 | 9.5 hrs |
| 10 — Frontend Payments | 2 | #34–#35 | 4.5 hrs |
| 11 — Polish | 3 | #36–#38 | 10 hrs |
| **Total** | **37** | | **~85 hrs** |

## Dependency Graph (Critical Path)

```
#1 (Go backend) ──┬──> #6 (SQLite) ──> #7 (users schema) ──┬──> #8 (other schemas) ──┬──> #10, #11, #12
                   │                                         │                         │
                   │                                         └──> #9 (users repo) ──┬──┤
                   │                                                                │  │
                   ├──> #13 (bcrypt) ──┬──> #15 (login)                             │  │
                   │                   ├──> #16 (me/setup)  ←───────────────────────-┘  │
                   ├──> #14 (JWT) ──┬──┘                                                │
                   │                ├──> #17 (role middleware) ──┬──> #18..#20 (user API, needs #13)
                   │                │                            ├──> #21..#23 (class API)
                   │                │                            └──> #24..#25 (payments API)
                   │                │
#2 (frontend) ────┬──> #26 (API client) ──> #27 (layout+dashboard) ──┬──> #28 (login/setup pages)
                   │                                                  ├──> #29..#30 (member UI)
                   │                                                  ├──> #31..#33 (class UI)
                   │                                                  └──> #34..#35 (payments UI)
                   │
#1 + #2 ──> #3 (embed) ──> #4 (Makefile+verify)

All leaf beads (#4, #15, #16, #20, #23, #25, #28, #33, #35) ──> #36..#38 (polish)
```
