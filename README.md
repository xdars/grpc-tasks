# gRPC Tasks - React Frontend Demo

This is a React frontend for the gRPC Tasks application that demonstrates:

- **Authentication** with JWT tokens
- **Task CRUD operations** via REST API (temporary, will be replaced with grpc-gateway)
- **Real-time streaming** of task events using Server-Sent Events
- **gRPC interceptors** working behind the scenes (logging, auth, recovery)

## Getting Started

### Prerequisites

- Go 1.19+ for the gRPC server
- Node.js 14+ for the React frontend
- Protocol Buffers compiler (protoc)

### Running the Application

1. **Start the Go gRPC server**:

   ```bash
   go run cmd/server/main.go
   ```

   The server will start on ports 50052 (gRPC) and 8080 (HTTP).

2. **Start the React frontend** (in a new terminal):

   ```bash
   cd frontend
   npm install
   npm start
   ```

   The React app will start on port 3000.

3. **Open your browser** to http://localhost:3000

## Features Demonstrated

### 🔐 Authentication

- User registration and login
- JWT token management
- Automatic token storage in localStorage

### 📋 Task Management

- Create new tasks with title and description
- View all tasks with status indicators
- Real-time updates when tasks are created

### 📡 Real-time Streaming

- Server-Sent Events for live task updates
- Event log showing all task creation events
- Automatic task list refresh on new events

### 🛡️ gRPC Interceptors

- **Authentication interceptor**: Validates JWT tokens
- **Logging interceptor**: Logs all gRPC calls
- **Recovery interceptor**: Handles panics gracefully

## How to Test the Features

1. **Register** a new user or login with existing credentials
2. **Create tasks** and watch them appear in the list
3. **Open another browser tab** with http://localhost:3000
4. **Login with the same user** in the second tab
5. **Create tasks in one tab** and see them appear instantly in both tabs!

## Architecture

```
Browser (React, port 3000) → HTTP Server (Go, port 8080) → gRPC Service (Go, port 50052)
     ↓                              ↓                              ↓
  REST API (proxied)    →   REST Handlers   →   gRPC Methods
  SSE Stream  →   Event Stream    →   Pub/Sub System
```

The React frontend runs on port 3000 with a proxy configuration that routes `/api/*` requests to the Go server on port 8080. The current setup uses temporary REST endpoints that internally call the gRPC service. In the next step, we'll replace this with [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) for proper REST API generation.

## Next Steps

1. ✅ **Frontend**: React app with authentication and streaming
2. 🔄 **REST Gateway**: Replace temporary REST with grpc-gateway
3. 📊 **Production Features**: Metrics, health checks, proper logging
4. 🐳 **Containerization**: Docker setup
5. 🔒 **Security**: TLS, proper CORS, rate limiting</content>
   <parameter name="filePath">/Users/alien/learn/grpc-tasks/README.md
