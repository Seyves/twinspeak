# API & HTTP Client Refactoring

## Overview
Refactored the API layer to use the `ky` HTTP client library with automatic token refresh and environment-based configuration.

## Changes Made

### 1. HTTP Client (`src/lib/httpClient.ts`)
- **Migrated from `fetch` to `ky`** - Uses ky's powerful hook system for request/response handling
- **Environment Variables** - Backend URL now configurable via `PUBLIC_BACKEND_URL` env variable
- **Automatic Token Management**:
  - `beforeRequest` hook: Automatically adds Bearer token to all requests
  - `beforeRetry` hook: Automatically refreshes expired tokens on 401 responses
  - `afterResponse` hook: Clears tokens and redirects to login on final 401 failure
- **Token Functions Exported**:
  - `getToken()` / `setToken()` - Access token management
  - `getRefreshToken()` / `setRefreshToken()` - Refresh token management
  - `clearTokens()` - Clear all tokens

### 2. Auth API (`src/api/auth.ts`)
- **Uses ky httpClient** - All requests now use the configured httpClient
- **New Functions Added**:
  - `signIn(credentials)` - Email/password login
  - `signUp(credentials)` - Email/password registration
- **Token Auto-Saving** - `googleProcessCallback`, `signIn`, and `signUp` automatically save tokens
- **Backward Compatibility** - Kept legacy exports: `setAuthToken()`, `getAuthToken()`, `clearAuthToken()`

### 3. Speech API (`src/api/speech.ts`)
- **Uses ky httpClient** - All requests now use the configured httpClient
- **FormData Support** - Properly handles file uploads with ky
- **Relative URLs** - Uses relative paths with httpClient's `prefixUrl`

### 4. Google Callback Component (`src/components/GoogleCallback.tsx`)
- **Simplified** - Removed redundant `setAuthToken()` call (now handled by `googleProcessCallback`)

### 5. Environment Variables
- **`.env` file created** with `PUBLIC_BACKEND_URL` configuration
- **`.env.example`** - Template for required environment variables

## Token Refresh Flow

```
1. User makes request
   ↓
2. `beforeRequest` hook adds Authorization header
   ↓
3. If 401 response and retryCount < 1:
   - `beforeRetry` hook calls `refreshToken()`
   - Refresh token endpoint is called
   - New access token is saved
   - Request is retried with new token
   ↓
4. If still 401 or no refresh token:
   - `afterResponse` hook clears tokens
   - User redirected to /auth
```

## Configuration

### Environment Variables
```env
# .env
PUBLIC_BACKEND_URL=http://localhost:3000/api/v1
```

For production/staging, update `PUBLIC_BACKEND_URL` to point to your backend server.

## Breaking Changes
None - All existing APIs remain compatible with backward-compatible exports.

## Dependencies Added
- `ky@^2.0.2` - Modern HTTP client with excellent TypeScript support and hook system
