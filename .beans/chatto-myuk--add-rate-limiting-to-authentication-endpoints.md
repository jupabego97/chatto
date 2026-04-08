---
# chatto-myuk
title: Add rate limiting to authentication endpoints
status: scrapped
type: bug
priority: normal
created_at: 2026-01-23T21:01:56Z
updated_at: 2026-04-08T16:02:12Z
order: zy
---

## Problem

Authentication endpoints lack rate limiting, making them vulnerable to brute-force attacks and credential stuffing. This is a security vulnerability that should be addressed.

## Affected Endpoints

In `cli/internal/http_server/auth.go`:

1. **Login endpoint** (line ~159) - `POST /api/auth/login`
   - Most critical - direct target for credential brute-forcing

2. **Registration endpoint** (line ~223) - `POST /api/auth/register`
   - Risk of spam account creation

3. **Forgot password endpoint** (line ~456) - `POST /api/auth/forgot-password`
   - Risk of email enumeration and spam

4. **Reset password endpoint** (line ~498) - `POST /api/auth/reset-password`
   - Risk of token brute-forcing

## Impact

- **Brute-force attacks**: Attackers can attempt unlimited password guesses
- **Credential stuffing**: Automated testing of leaked credentials
- **Resource exhaustion**: Spam registration can fill storage
- **Email spam**: Forgot-password abuse can spam users

## Suggested Solution

Implement rate limiting using a sliding window or token bucket algorithm:

\`\`\`go
// Per-IP rate limiting for auth endpoints
type RateLimiter struct {
    requests map[string]*rateBucket
    mu       sync.RWMutex
}

// Apply to auth handlers
func (h *AuthHandler) Login(c *gin.Context) {
    if !h.rateLimiter.Allow(c.ClientIP(), "login", 5, time.Minute) {
        c.JSON(429, gin.H{"error": "too many requests"})
        return
    }
    // ... existing logic
}
\`\`\`

Consider also:
- Exponential backoff after failed attempts
- Account lockout after N failures (with notification)
- CAPTCHA integration for repeated failures

## Files to Modify

- [ ] `cli/internal/http_server/rate_limiter.go` (new file)
- [ ] `cli/internal/http_server/auth.go` - apply rate limiting to endpoints
- [ ] `cli/internal/http_server/http_server.go` - initialize rate limiter

## Acceptance Criteria

- [ ] Rate limiter implementation (sliding window or token bucket)
- [ ] Login endpoint: max 5 attempts per minute per IP
- [ ] Registration endpoint: max 3 registrations per hour per IP
- [ ] Forgot-password endpoint: max 3 requests per hour per email
- [ ] Reset-password endpoint: max 5 attempts per token
- [ ] Tests for rate limiting behavior
- [ ] Appropriate 429 responses with Retry-After header

## Reasons for Scrapping

Duplicate of rate limiting beans. Consolidated.
