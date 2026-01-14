# Hearts Dating App

A modern dating application built with microservices architecture, real-time features, and event-driven communication.

## Architecture

- **Backend**: Go 1.24
  - **API Service** (`apps/api`): Main REST API, WebSocket (Chat), Clean Architecture.
  - **Notification Service** (`apps/notification-service`): Microservice for handling system notifications via Kafka and WebSockets.
- **Frontend**: React (Vite, TypeScript, TanStack Query).
- **Messaging**: Kafka (Event bus for inter-service communication).
- **Database**: PostgreSQL (Store), Redis (Cache - Optional/Planned), MinIO (Object Storage).
- **Communication**:
  - HTTP/REST
  - WebSocket (Real-time Chat & Notifications)
  - Kafka (Async Events)

## Prerequisites

- API: Go 1.24+
- Frontend: Node.js 18+
- Infrastructure: Docker & Docker Compose

## Getting Started

### 1. Start Infrastructure & Backend

The project uses Docker Compose to orchestrate the backend services (API, Notifications, Postgres, Kafka, MinIO).

```bash
# In the root directory
docker-compose up --build
```

This will start:

- **API Service**: http://localhost:8080
- **Notification Service**: http://localhost:8081
- **MinIO Console**: http://localhost:9001
- **Postgres**: localhost:5433
- **Kafka**: localhost:9092

### 2. Start Frontend

```bash
cd apps/web
npm install
npm run dev
```

## Features

- **Real-time Chat**: WebSocket-based secure chat with Ticket System authentication.
- **Typing Indicators**: Real-time "User is typing..." updates.
- **Notifications**: Async notifications processed via Kafka.
- **Microservices**: Decoupled Notification Service demonstrating Event-Driven Architecture.

## Project Structure

```
├── apps
│   ├── api                 # Main Backend Service
│   ├── notification-service # Notification Microservice
│   └── web                 # React Frontend
├── docker-compose.yml      # Orchestration
└── go.work                 # Go Workspace
```
