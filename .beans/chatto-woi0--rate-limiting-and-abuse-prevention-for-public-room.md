---
# chatto-woi0
title: Rate limiting and abuse prevention for public room endpoints
status: todo
type: task
priority: normal
created_at: 2026-02-10T08:06:03Z
updated_at: 2026-03-18T05:34:10Z
order: zzzzzzzzzz
parent: chatto-2quz
blocked_by:
    - chatto-bwc7
---

Add rate limiting to public room queries to prevent abuse from unauthenticated traffic.

## Implementation

### Rate limiting
- Add IP-based rate limiting to GraphQL queries that serve public rooms
- Suggested limits: ~60 requests/minute per IP for room events queries
- Use a simple in-memory rate limiter (e.g., \`golang.org/x/time/rate\` or similar)
- Consider whether this should be middleware-level or resolver-level

### Approach options

**Option A: Middleware-level (recommended)**
- Add rate limiting middleware in \`cli/internal/http_server/\` that applies to unauthenticated requests
- Authenticated users are not rate-limited (they're already trusted via session)
- Simple token bucket per IP

**Option B: Resolver-level**
- Check rate limit inside the public room auth bypass
- More granular but more complex

### Considerations
- Don't rate-limit authenticated users (they already passed auth)
- Log excessive request patterns for monitoring
- Return standard 429 Too Many Requests with Retry-After header
- Consider caching public room responses at the HTTP level (Cache-Control headers) to reduce load

### HTTP caching
- For public room event queries, add \`Cache-Control: public, max-age=30\` headers
- This allows CDN/proxy caching and reduces origin load
- The 30s cache aligns with the frontend polling interval

## Key files
- \`cli/internal/http_server/\` (middleware)
- \`cli/internal/graph/authz.go\` (if resolver-level)

## Tests
- Unit test: unauthenticated requests are rate-limited
- Unit test: authenticated requests bypass rate limiting
- Unit test: 429 response includes Retry-After header
