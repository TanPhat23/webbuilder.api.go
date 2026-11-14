# Real-time Web Builder Architecture

This document describes the architectural design and implementation of the real-time synchronization system for the Web Builder.

## 1. System Overview

The system is built using Go, following Clean Architecture principles, and leverages WebSockets for low-latency bidirectional communication. It is designed to be horizontally scalable, allowing multiple server instances to synchronize state via a distributed messaging backplane.

## 2. Core Components

### 2.1 WebSocket Handler (`internal/delivery/ws`)

The entry point for real-time connections. It manages the lifecycle of a WebSocket connection and delegates incoming messages to specific command handlers.

- **Command Pattern:** Uses a registry of handlers (`map[MessageType]MessageHandler`) to decouple message parsing from business logic.
- **Client Management:** Each connection is wrapped in a `Client` entity that handles buffered writes and identifies the user's current context (Project and Page).

### 2.2 Hub & Room System (`internal/delivery/ws/entity`)

The Hub manages the logical grouping of clients into "Rooms" (usually based on `ProjectID:PageID`).

- **Rooms:** Encapsulate a set of clients collaborating on the same document.
- **Concurrency:** Uses Go channels for thread-safe registration, unregistration, and broadcasting.

### 2.3 Distributed Pub/Sub (Redis)

To support horizontal scaling, the `Hub` integrates with Redis Pub/Sub.

1. **Local Broadcast:** When a message is received, it is immediately sent to all clients on the _same_ server.
2. **Global Publish:** The message is published to a Redis channel (`ws:room:<id>`).
3. **Cluster Sync:** Other server instances subscribed to that channel receive the message and relay it to their local connected clients.
4. **Loop Prevention:** Each server has a unique ID to prevent re-broadcasting messages it originally published.

## 3. Communication Protocol

### 3.1 Message Envelope

All messages sent over WebSocket follow this standard JSON structure:

| Field       | Type     | Description                              |
| :---------- | :------- | :--------------------------------------- |
| `type`      | `string` | The action type (e.g., `element:update`) |
| `projectId` | `string` | ID of the project                        |
| `pageId`    | `string` | ID of the page                           |
| `timestamp` | `int64`  | Unix milliseconds                        |
| `payload`   | `object` | Action-specific data                     |

### 3.2 Message Types

#### Presence & Connection

- `join`: Sent by the client to enter a page. Triggers an initial state sync.
- `leave`: Sent by the client to exit a page.
- `presence`: Broadcasts cursor positions or user selections.

#### Document Operations

- `element:create`: Broadcasts a new element added to the canvas.
- `element:update`: Performs granular updates (styles, settings).
- `element:delete`: Removes an element.
- `element:move`: Reorders or changes the parent of an element.

#### Synchronization

- `sync`: Requests the current full state of the document.
- `error`: Sent by the server when an operation fails.

## 4. Scalability Considerations

### 4.1 Horizontal Scaling

By using Redis Pub/Sub, the application layer is stateless regarding collaborative state. You can spin up multiple instances behind a Round Robin or Least Connections load balancer.

### 4.2 Performance Optimization

- **Buffered Channels:** Each client has a configurable write buffer to prevent slow network consumers from blocking the room's event loop.
- **Granular Updates (Patching):** Instead of sending the entire document or entire element on every change, the system sends only the modified fields (e.g., just `styles`).
- **Write-Through Cache:** Document state is cached in Redis for fast retrieval during `join` events, reducing database load.

## 5. Error Handling

The system uses a structured error protocol. When an error occurs during message processing, the server responds with an `error` message type containing:

- `code`: A machine-readable string (e.g., `VALIDATION_ERROR`, `NOT_FOUND`).
- `message`: A human-readable description.

This allows the frontend to provide specific feedback to users when collaborative actions fail.
