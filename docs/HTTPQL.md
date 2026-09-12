# HTTPQL

HTTPQL is Hettix's typed query language for filtering captured traffic. Type a
query into the search box on Proxy logs (and other traffic views) to narrow the
list. Queries are matched in two phases: exact comparisons are pushed down to
SQLite as a fast pre-filter, and the full expression is then matched in memory.

## Quick examples

```text
req.method = "POST"
resp.code >= 400
req.host cont "example.com" and resp.code = 200
req.path =~ "/api/v[0-9]+/" and not resp.code = 404
req.header["Content-Type"] cont "json"
admin                       // free-text: matches anywhere in the request/response
```

## Structure

A query is a boolean expression of terms:

- `and` — both sides must match. Adjacent terms are **implicitly AND-ed**, so
  `req.method = "GET" resp.code = 200` means the same as joining them with `and`.
- `or` — either side matches (lower precedence than `and`).
- `not` — negates the following term.
- `( … )` — parentheses group expressions.
- `//` line comments and `/* … */` block comments are ignored.

## Fields

Reference a field as `namespace.field`. Namespaces: `req`, `resp` (alias `res`),
and `ws` (WebSocket).

### Request (`req`)

| Field | Type | Notes |
|-------|------|-------|
| `req.id` | string | |
| `req.method` | string | GET, POST, … |
| `req.host` | string | |
| `req.path` | string | path without query |
| `req.query` | string | raw query string |
| `req.ext` | string | path file extension (e.g. `js`, `png`) |
| `req.url` | string | full URL |
| `req.proto` | string | e.g. `HTTP/1.1` |
| `req.body` | string | |
| `req.port` | int | |
| `req.len` | int | body length |
| `req.tls` | bool | `true` / `false` |
| `req.created_at` | time | RFC 3339, e.g. `"2025-01-02T15:04:05Z"` |

### Response (`resp` / `res`)

| Field | Type | Notes |
|-------|------|-------|
| `resp.proto` | string | |
| `resp.code` | int | status code (legacy alias: `statusCode`) |
| `resp.reason` | string | status reason (legacy alias: `statusReason`) |
| `resp.body` | string | |
| `resp.len` | int | body length |
| `resp.roundtrip` | int | round-trip time |

### WebSocket (`ws`)

| Field | Type | Notes |
|-------|------|-------|
| `ws.host` `ws.path` `ws.url` | string | |
| `ws.direction` | string | |
| `ws.payload` | string | |
| `ws.opcode` | int | |

### Headers

Address a specific header by name (case-insensitive key):

```text
req.header["Authorization"] cont "Bearer"
resp.header["Content-Type"] = "application/json"
```

Bare `req.header` / `resp.header` matches against every `Key: Value` line.

## Operators

| Word | Symbol | Meaning |
|------|--------|---------|
| `eq` | `=`, `==` | equals |
| `ne` | `!=` | not equal |
| `cont` | | contains (case-insensitive substring) |
| `ncont` | | does not contain |
| `like` | | SQL `LIKE` (`%`, `_` wildcards, case-insensitive) |
| `nlike` | | negated `LIKE` |
| `regex` | `=~` | matches Go/RE2 regular expression |
| `nregex` | `!~` | does not match regex |
| `gt` `gte` | `>` `>=` | greater than / or equal (int, time) |
| `lt` `lte` | `<` `<=` | less than / or equal (int, time) |

## Values

- **Strings** are double-quoted with `\n \r \t \\ \"` escapes: `"a b"`.
- **Integers** are bare numbers: `req.port = 443`.
- **Booleans** are `true` / `false`: `req.tls = true`.
- **Times** are RFC 3339 strings: `req.created_at gt "2025-01-01T00:00:00Z"`.
- A **bareword** that isn't `namespace.field` is a free-text term, searching the
  method, URL, proto, body and all headers, case-insensitively.
