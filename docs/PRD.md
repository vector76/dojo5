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
| View own payments | Yes | No* | Yes |
| View all members | Yes | Yes | No |
| Edit own profile | Yes | Yes | Yes |

*Instructors are staff, not students — they don't have membership dues or payment records. Payment tracking applies to users (students/prospective members) only.

### 2. Member Management
- Add/edit/soft-delete members (deletion preserves attendance and payment history)
- **Required fields**: name, email (unique), phone
- **Optional fields**: emergency contact, join date, membership status
- Track membership type (monthly, annual, drop-in)

### 3. Class/Session Management
- Define class types (e.g., "Beginner Yoga", "Advanced Karate") - admin only
- Schedule classes with instructor, time, duration, capacity - admin only
- Track attendance (which members attended which class) - admin or instructor

### 4. Basic Payments/Dues Tracking
- Record payments received - admin only
- View payment history per member
- Simple outstanding balance indicator: admin manually sets each member's expected balance (no automated dues schedule); payments reduce it; balance endpoint shows the difference

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

## Frontend Design Guidelines

### CSS Styling Philosophy
The application should have a **clean, simple, modern look** with the following principles:

- **Visual Hierarchy**: Clear distinction between sections using cards, borders, and shadows
- **Typography**: Modern system fonts with appropriate sizing and spacing
- **Color Scheme**: Support both light and dark modes with proper contrast ratios
- **Spacing**: Generous padding and margins for readability
- **Interactive Elements**:
  - Buttons with hover effects, subtle shadows, and proper focus states
  - Inputs with focus rings and clear borders
  - Tables with hover states and clean borders
- **Accessibility**:
  - Proper focus-visible states for keyboard navigation
  - Sufficient color contrast (WCAG AA compliant)
  - Clear visual feedback for interactive elements
- **Responsive**: Mobile-first approach with collapsible sidebar and stacked layouts

**Implementation Notes**:
- Use `index.css` for global styles (typography, forms, tables, buttons)
- Use `App.css` for component-specific styles (sidebar, layout, sections)
- Import `App.css` in `App.tsx` to ensure styles are bundled
- Leverage CSS variables and media queries for theming
- Avoid over-engineering - keep it simple and maintainable

## Build & Deploy Process

1. `make build` builds frontend and backend:
   - `frontend-build`: Runs `npm run build` → produces `frontend/dist/`
   - Copies `frontend/dist/` to `backend/internal/frontend/dist/`
   - `backend-build`: Go build embeds `backend/internal/frontend/dist/` using `//go:embed`
2. Result: single binary that serves React app and API

**Important**:
- **Do NOT commit build artifacts** (`backend/internal/frontend/dist/`) to git
- Build artifacts are added to `.gitignore` to keep repo clean
- Always run `make build` to generate fresh assets before deployment
- This reduces git noise and keeps repository size minimal

## Status

**APPROVED** - Ready for implementation.
