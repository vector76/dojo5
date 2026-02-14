# Simple CRM for Dojo/Yoga Studio

## Overview
A lightweight CRM system for managing a small martial arts dojo or yoga studio, with a single-binary deployment model.

## Tech Stack
- **Frontend**: Vite + React (TypeScript)
- **Backend**: Go with embedded static files (using `embed` package)
- **Database**: SQLite (separate file from executable)
- **Deployment**: Single executable with embedded frontend assets
- **Testing**: TDD approach with comprehensive automated tests

## Development Methodology

**Test-Driven Development (TDD)**
- Write tests before or alongside implementation
- Backend: Go testing package with table-driven tests
- Frontend: Vitest + React Testing Library
- Aim for high test coverage on business logic and API endpoints

## Core Features (MVP)

### 1. Authentication & User Management
- **Roles**: admin, instructor, user (student/prospective)
- **Login**: Email serves as username (required)
- **Required fields**: Email (unique, used for login), phone number, name
- **First-run setup**: Virgin app presents UI to create initial admin account
- **User management**: Admin can create users and adjust roles
- **Password reset**: Admin manually resets passwords (no email flow)
- **Session management**: JWT tokens (stateless, stored in localStorage)

#### Role Permissions
| Action | Admin | Instructor | User |
|--------|-------|------------|------|
| Manage users/roles | Yes | No | No |
| Manage class types | Yes | No | No |
| Schedule classes | Yes | No | No |
| Record attendance | Yes | Yes | No |
| View own attendance | Yes | Yes | Yes |
| Record payments | Yes | No | No |
| View own payments | Yes | No | Yes |
| View all members | Yes | Yes | No |
| Edit own profile | Yes | Yes | Yes |

### 2. Member Management
- Add/edit/delete members
- **Required fields**: name, email (unique), phone
- **Optional fields**: emergency contact, join date, membership status
- Track membership type (monthly, annual, drop-in)

### 3. Class/Session Management
- Define class types (e.g., "Beginner Yoga", "Advanced Karate") - admin only
- Schedule classes with instructor, time, capacity - admin only
- Track attendance (which members attended which class) - admin or instructor

### 4. Basic Payments/Dues Tracking
- Record payments received - admin only
- View payment history per member
- Simple outstanding balance indicator

## Confirmed Assumptions

1. **Multi-user with roles**: Admin, Instructor, User roles with permissions as defined above
2. **Email as username**: Required, unique, used for login
3. **Phone required**: All users must have a phone number
4. **JWT authentication**: Stateless tokens for simplicity
5. **Admin password reset**: No email flow, admin resets manually
6. **No online payments**: Record-keeping only (cash/check/external processor)
7. **No email/SMS notifications**: Database and UI only
8. **Attendance**: Simple check-in, not advance booking/reservations
9. **Reports**: Basic views only, no PDF exports or charts
10. **Mobile**: Responsive web design, no native app
11. **Data backup**: Manual SQLite file backup, no cloud sync

## Project Structure

```
/
├── frontend/               # Vite React app
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── __tests__/
│   ├── package.json
│   └── vite.config.ts
├── backend/                # Go application
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── database/
│   │   ├── auth/
│   │   └── middleware/
│   ├── embed.go            # Embeds built frontend
│   └── go.mod
├── Makefile                # Build commands
└── README.md
```

## Build & Deploy Process

1. `npm run build` in frontend/ → produces `dist/`
2. Go build embeds `frontend/dist/` using `//go:embed`
3. Result: single binary that serves React app and API

## Status

**APPROVED** - Ready for implementation.
