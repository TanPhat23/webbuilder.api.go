# WebSocket API Reference

This document provides a detailed reference for the WebSocket API used in the Web Builder real-time synchronization system.

## 1. Connection

The WebSocket endpoint is dynamic based on the project ID.

- **URL:** `ws://<host>:<port>/ws/:project`
- **Protocol:** `json`

Example: `ws://localhost:8080/ws/my-project-id`

Authentication note:

- The server expects an Authorization header on the WebSocket upgrade request (for example: `Authorization: Bearer <token>`). Node clients (e.g., `ws`) can send headers on the handshake.
- Note: Browsers' native WebSocket API cannot set arbitrary request headers on the initial handshake. For browser-based clients, pass the token in a query parameter (for example: `wss://host/ws/my-project-id?token=<jwt>`) or use cookies and ensure the server reads the token from the query string or cookie during the upgrade.
- Example (Node `ws` client):

```js
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:8080/ws/my-project-id', {
  headers: { Authorization: 'Bearer <token>' },
});
```

---

## 2. Message Envelope

All communication (incoming and outgoing) uses a standard envelope:

The envelope includes the following fields:

| Field       | Type     | Description                                                                              |
| :---------- | :------- | :--------------------------------------------------------------------------------------- |
| `type`      | `string` | The action type (e.g., `element:update`)                                                 |
| `projectId` | `string` | ID of the project                                                                        |
| `pageId`    | `string` | ID of the page                                                                           |
| `userId`    | `string` | ID of the acting user (optional but recommended)                                         |
| `requestId` | `string` | Optional client-generated ID to correlate request/response (useful for acknowledgements) |
| `timestamp` | `int64`  | Unix milliseconds                                                                        |
| `payload`   | `object` | Action-specific data                                                                     |
| `clientOpId` | `string` | Optional client-generated operation id on `element:*` requests. Echoed on the resulting broadcast and on `error` replies (see [Ordering and acknowledgements](#ordering-and-acknowledgements)). |
| `requestId` | `string` | Optional client-generated request id. Echoed on the reply to that request: the `element:*` broadcast, the `sync:page` / `sync:project` response, or an `error`. |
| `seq`       | `int64`  | Server-assigned, per-page sequence number. Present only on `element:create/update/move/delete` broadcasts. |

Example envelope (client -> server):

```json
{
  "type": "element:create",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "requestId": "req-1234",
  "timestamp": 1680000000000,
  "payload": {
    "element": {
      "type": "text",
      "settings": { "content": "Hello World" },
      "styles": { "color": "black" },
      "order": 0
    }
  }
}
```

On success, the server will broadcast an `element:create` message to all clients in the room (including the sender) including server-assigned fields (e.g., `id`, `createdAt`, `createdBy`) and may include the original `requestId` for correlation.

---

## 3. Client to Server Messages

### Join Page

Must be sent before performing any element operations.

- **Type:** `join`
- **Payload:**

```json
{
  "pageId": "string"
}
```

_Note: Upon joining, the server will automatically respond with a `sync` message containing the current page state._

### Presence (Cursor/Selection)

Broadcasts your current activity to other users.

- **Type:** `presence`
- **Payload:**

```json
{
  "userId": "string",
  "userName": "string",
  "cursorX": 0.5,
  "cursorY": 0.8,
  "elementId": "uuid-string (optional)",
  "meta": {}
}
```

### Create Element

- **Type:** `element:create`

- **Client Payload Example:**

```json
{
  "type": "element:create",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "requestId": "req-1",
  "timestamp": 1680000000000,
  "payload": {
    "element": {
      "type": "text",
      "settings": { "content": "Hello World" },
      "styles": { "fontSize": 14, "color": "black" },
      "order": 0,
      "parentId": null
    }
  }
}
```

- **Server Broadcast Example (after persistence):**

```json
{
  "type": "element:create",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "system",
  "timestamp": 1680000001000,
  "payload": {
    "element": {
      "id": "uuid-1111",
      "type": "text",
      "settings": { "content": "Hello World" },
      "styles": { "fontSize": 14, "color": "black" },
      "order": 0,
      "parentId": null,
      "createdAt": 1680000001000,
      "createdBy": "user-abc"
    }
  }
}
```

Notes:

- The server is authoritative — clients should reconcile optimistic client updates with server broadcasts using `requestId` and the assigned `id`.
- If creation fails validation, the server will return an `error` message referencing the original `requestId` when available.

### Update Element (Partial)

Updates only the provided fields (patch semantics).

- **Type:** `element:update`

- **Client Payload Example:**

```json
{
  "type": "element:update",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "requestId": "req-2",
  "timestamp": 1680000010000,
  "payload": {
    "id": "uuid-1111",
    "settings": { "content": "Hello Universe" },
    "styles": { "color": "blue" }
  }
}
```

- **Server Broadcast Example (merged patch):**

```json
{
  "type": "element:update",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "timestamp": 1680000010100,
  "payload": {
    "id": "uuid-1111",
    "settings": { "content": "Hello Universe" },
    "styles": { "color": "blue" }
  }
}
```

Notes:

- Updates are **merged**, never replaced:
  - `styles` and `settings` are deep-merged into the stored maps. A key whose value is JSON `null` is deleted. Nested objects are merged recursively; arrays and scalars replace the stored value.
  - `properties` set the element's own fields. Supported keys: `name`, `content`, `href`, `src`, `tailwindStyles` (string or `null`) and `isLocked` (boolean). Unknown keys or wrongly typed values are rejected with `VALIDATION_ERROR`. Properties are never written into `settings`.
  - For compatibility with the web client, `name`, `content`, `src`, `href` and `tailwindStyles` are also accepted as top-level payload fields; they are equivalent to the same keys in `properties` (an explicit `properties` entry wins).
- The server merges partial updates into the stored element; the broadcasted message is authoritative.
- Conflict resolution is handled server-side (e.g., timestamp-based LWW or optimistic locking); clients should apply server broadcasts as source of truth.

### Move Element

Reparents or reorders an element.

- **Type:** `element:move`
- **Payload (client -> server):**

```json
{
  "id": "uuid-string",
  "parentId": "uuid-string or null",
  "order": 5
}
```

`order` is the target index among the new parent's children. The server allocates the element's `orderKey` (see below) and ignores any `orderKey` sent by the client.

- **Payload (server broadcast):** the same fields, with `order`, `orderKey` and `parentId` set to the values the server actually applied:

```json
{
  "id": "uuid-string",
  "parentId": "uuid-string or null",
  "order": 3,
  "orderKey": "a1V"
}
```

#### Order keys

Siblings are ordered by `orderKey`, a string fractional index (`roci.dev/fracdex`). Compare keys as plain strings; a lexicographically smaller key sorts first. Inserting or moving an element only assigns a key between its two new neighbours, so other siblings are never renumbered. Every element in `element:create` broadcasts and `sync:page` snapshots carries `orderKey`. The integer `order` is kept for compatibility and always equals the element's index among its siblings. It is derived from `orderKey`, so clients should sort by `orderKey` and use `order` only as a fallback for elements that have no key. When a document has missing, duplicate or out-of-order keys, the server repairs them on the next edit while keeping the existing sibling order.

### Delete Element

- **Type:** `element:delete`
- **Payload:**

```json
{
  "id": "uuid-string"
}
```

### Sync Page Request

Force a refresh of the current page state.

- **Type:** `sync:page`
- **Payload:** Empty

Example:
```json
{
  "type": "sync:page",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "requestId": "req-sync-1",
  "timestamp": 1680000020000,
  "payload": {}
}
```

### Sync Project Request

Force a refresh of the entire project state (all pages, workflows, element workflows).

- **Type:** `sync:project`
- **Payload:** Empty

Example:
```json
{
  "type": "sync:project",
  "projectId": "project-123",
  "userId": "user-abc",
  "requestId": "req-sync-proj-1",
  "timestamp": 1680000020000,
  "payload": {}
}
```

Note: Project-wide sync does not require a `pageId` and returns state for all pages in the project.

---

## 4. Server to Client Messages

### Initial Page State Sync

Sent automatically after a successful `join` or requested via `sync:page`. This message contains the authoritative document snapshot for the current page and the presence information for active users.

- **Type:** `sync:page`
- **Payload:**

```json
{
  "elements": [
    {
      "id": "uuid-1111",
      "type": "text",
      "settings": { "content": "Hello World" },
      "styles": { "fontSize": 14, "color": "black" },
      "order": 0,
      "parentId": null,
      "createdAt": 1680000001000,
      "createdBy": "user-abc"
    }
  ],
  "users": [
    {
      "userId": "user-abc",
      "userName": "Alice",
      "presence": {
        "cursorX": 0.5,
        "cursorY": 0.2,
        "elementId": null
      }
    }
  ],
  "seq": 41
}
```

`seq` is the page's current sequence number (0 if no operation has been applied yet). The `elements` snapshot already contains every operation with `seq <= payload.seq`; the next operation to apply is `seq + 1`.

Server Response Example (with requestId correlation):
```json
{
  "type": "sync:page",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "system",
  "requestId": "req-sync-1",
  "timestamp": 1680000020100,
  "payload": {
    "elements": [...],
    "users": [...]
  }
}
```

### Project State Sync

Sent in response to a `sync:project` request. This message contains the authoritative state for the entire project, including all pages, workflows, and element workflows.

- **Type:** `sync:project`
- **Payload:**

```json
{
  "elements": [
    {
      "id": "uuid-1111",
      "type": "text",
      "settings": { "content": "Hello World" },
      "styles": { "fontSize": 14, "color": "black" },
      "order": 0,
      "parentId": null,
      "createdAt": 1680000001000,
      "createdBy": "user-abc"
    }
  ],
  "workflows": [
    {
      "id": "workflow-1",
      "name": "Main Workflow",
      "createdAt": 1680000001000,
      "createdBy": "user-abc"
    }
  ],
  "elementWorkflows": [
    {
      "id": "elem-workflow-1",
      "elementId": "uuid-1111",
      "workflowId": "workflow-1"
    }
  ],
  "users": [
    {
      "userId": "user-abc",
      "userName": "Alice",
      "presence": {
        "cursorX": 0.5,
        "cursorY": 0.2,
        "elementId": null
      }
    }
  ]
}
```

Server Response Example (with requestId correlation):
```json
{
  "type": "sync:project",
  "projectId": "project-123",
  "userId": "system",
  "requestId": "req-sync-proj-1",
  "timestamp": 1680000020100,
  "payload": {
    "elements": [...],
    "workflows": [...],
    "elementWorkflows": [...],
    "users": [...]
  }
}
```

Client Snippets (JavaScript):

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/my-project-id');

ws.addEventListener('open', () => {
  // join the page (must be done before element ops)
  ws.send(
    JSON.stringify({
      type: 'join',
      projectId: 'my-project-id',
      pageId: 'page-1',
      userId: 'user-1',
      timestamp: Date.now(),
      payload: { pageId: 'page-1' },
    })
  );
});

ws.addEventListener('message', (ev) => {
  const msg = JSON.parse(ev.data);
  switch (msg.type) {
    case 'sync:page':
      // replace local page document with msg.payload.elements
      break;
    case 'sync:project':
      // replace entire project state with elements, workflows, and elementWorkflows
      break;
    case 'element:create':
      // insert element
      break;
    case 'element:update':
      // apply patch
      break;
    case 'error':
      // handle errors
      break;
  }
});
```

Reconnection guidance:

- Use exponential backoff with jitter (e.g., base 500ms, multiplier 1.5, max 30s).
- On reconnect, re-send `join` (with same `userId`) and request a `sync:page` if needed.
- Use `requestId` and message `timestamp`/server-assigned `id` to deduplicate and reconcile missed messages.
- Apply server-sent `sync:page` as authoritative snapshot; reconcile client optimistic updates by matching `requestId` or element IDs.

### Broadcasted Operations

When a user performs an element action (create, update, delete, move), the server applies it and broadcasts the same message type and payload to **all** participants in the room, **including the sender**. Presence, comment and workflow messages are unchanged (presence is still not echoed to its sender and carries no `seq`).

#### Ordering and acknowledgements

Every successful `element:create/update/move/delete` gets the next per-page `seq` (1, 2, 3, ...), assigned atomically with the change in Redis, so it is consistent across server instances. The broadcast envelope carries it:

```json
{
  "type": "element:update",
  "projectId": "project-123",
  "pageId": "page-456",
  "userId": "user-abc",
  "timestamp": 1680000010100,
  "seq": 42,
  "clientOpId": "op-7f3a",
  "payload": { "id": "uuid-1111", "styles": { "color": "blue" } }
}
```

Client rules:

- Send an optional `clientOpId` (unique per client) on each `element:*` request. The sender receives its own broadcast with the same `clientOpId` and `seq`: that echo is the acknowledgement (drop the pending optimistic op, do not apply it twice).
- Apply operations in `seq` order. Ignore any op with `seq <=` the last applied `seq` (this includes ops already contained in a `sync:page` snapshot). Buffer an op that arrives with a gap and, if the missing `seq` does not arrive shortly, send `sync:page` to resynchronise. A `seq` that goes backwards or resets (for example after a long idle period) must also trigger a resync.
- On failure the server does not broadcast and does not consume a `seq`. The sender receives an `error` message carrying the same `clientOpId` (on the envelope and in the payload) so the optimistic op can be rolled back.
- If the server cannot guarantee delivery to a client (its send buffer or the room's queue overflows) it logs this and closes that connection instead of silently dropping a message. Reconnect, `join` and continue from the `seq` in `sync:page`.

### Error Report

Sent when an operation fails validation or encounters a system error.

- **Type:** `error`
- **Payload:**

```json
{
  "code": "VALIDATION_ERROR | NOT_FOUND | JOIN_REQUIRED",
  "message": "Detailed error description",
  "clientOpId": "op-7f3a (only if the failed request carried one)"
}
```

---

## 5. Standard Error Codes

| Code               | Description                                                 |
| :----------------- | :---------------------------------------------------------- |
| `VALIDATION_ERROR` | The payload data failed schema validation.                  |
| `NOT_FOUND`        | The requested element or project does not exist.            |
| `JOIN_REQUIRED`    | You attempted an operation before sending a `join` message. |
| `PROCESS_ERROR`    | An internal error occurred while processing the request.    |
