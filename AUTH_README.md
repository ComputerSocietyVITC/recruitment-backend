# JWT Authentication and RBAC Documentation

## Overview

This backend implements JWT (JSON Web Token) authentication with Role-Based Access Control (RBAC). The system supports four roles with different permission levels:

1. **Applicant** - Basic user role for job applicants
2. **Evaluator** - Can view and evaluate applications
3. **Admin** - Can manage users and roles (except super admin)
4. **Super Admin** - Full system access, can manage all roles

## Environment Variables

Ensure these variables are set in your `.env` file:

```bash
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY_DURATION=24h  # Optional: JWT token expiry duration (default: 24h)
                         # Examples: 1h, 30m, 2h30m, 7d, 168h
```

## API Endpoints

### Authentication Endpoints (Public)

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "phone_number": "+1234567890", // optional
  "role": "applicant" // optional, defaults to "applicant"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer <your-jwt-token>
```

#### Get Profile
```http
GET /api/v1/auth/profile
Authorization: Bearer <your-jwt-token>
```

### User Management Endpoints (Protected)

#### Create User (Admin+ only)
```http
POST /api/v1/users
Authorization: Bearer <your-jwt-token>
Content-Type: application/json

{
  "full_name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123",
  "role": "evaluator"
}
```

#### Get All Users (Evaluator+ only)
```http
GET /api/v1/users
Authorization: Bearer <your-jwt-token>
```

#### Get User by ID
```http
GET /api/v1/users/{user-id}
Authorization: Bearer <your-jwt-token>
```

### Super Admin Endpoints

#### Update User Role (Super Admin only)
```http
PUT /api/v1/super-admin/users/{user-id}/role
Authorization: Bearer <your-jwt-token>
Content-Type: application/json

{
  "role": "evaluator"
}
```

## Role Permissions

| Endpoint | Applicant | Evaluator | Admin | Super Admin |
|----------|-----------|-----------|-------|-------------|
| POST /auth/register | ✅ | ✅ | ✅ | ✅ |
| POST /auth/login | ✅ | ✅ | ✅ | ✅ |
| GET /auth/profile | ✅ | ✅ | ✅ | ✅ |
| POST /users | ❌ | ❌ | ✅ | ✅ |
| GET /users | ❌ | ✅ | ✅ | ✅ |
| GET /users/:id | ✅ | ✅ | ✅ | ✅ |
| PUT /super-admin/users/:id/role | ❌ | ❌ | ❌ | ✅ |

## Role Assignment Rules

- **Applicant role**: Can be assigned by anyone during registration
- **Evaluator role**: Can only be assigned by Super Admin
- **Admin role**: Can only be assigned by Super Admin
- **Super Admin role**: Can only be assigned by Super Admin

## JWT Token Structure

The JWT tokens contain the following claims:

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "applicant|evaluator|admin|super_admin",
  "iss": "recruitment-backend",
  "sub": "user-uuid",
  "iat": 1234567890,
  "exp": 1234567890,
  "nbf": 1234567890
}
```

## Middleware Usage

### Authentication Middleware
```go
// Requires valid JWT token
router.Use(middleware.JWTAuthMiddleware())
```

### Role-Based Middleware
```go
// Require specific roles
router.Use(middleware.RoleBasedAuthMiddleware(models.RoleAdmin, models.RoleSuperAdmin))

// Convenience middlewares
router.Use(middleware.SuperAdminOnlyMiddleware())
router.Use(middleware.AdminOrAboveMiddleware())
router.Use(middleware.EvaluatorOrAboveMiddleware())
```

## Error Responses

All authentication errors return appropriate HTTP status codes:

- `401 Unauthorized`: Invalid or missing token
- `403 Forbidden`: Valid token but insufficient permissions
- `400 Bad Request`: Invalid request format
- `500 Internal Server Error`: Server-side errors

Example error response:
```json
{
  "error": "Insufficient permissions"
}
```
