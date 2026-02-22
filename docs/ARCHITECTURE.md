# Vargo Microservices Architecture Plan

This document outlines the architectural plan for the `vargo` e-commerce application, focusing on a polyglot microservices approach to demonstrate fundamentals of distributed systems.

## Core Principles

1.  **Polyglot Persistence & Compute**: Using the "right tool for the job" based on service characteristics (I/O bound, CPU bound, AI, Enterprise/Complex Domain).
2.  **Event-Driven Architecture**: Services communicate asynchronously for non-critical path operations (e.g., "Order Placed" -> "Send Email") to ensure loose coupling and high availability.
3.  **Synchronous (gRPC/HTTP)**: Used for critical, real-time data requirements where consistency is paramount (e.g., API Gateway checking Auth, or Order Service checking Inventory availability).

## Service Breakdown

### 1. Auth Service (`/services/auth`)

- **Role**: Handles user registration, authentication, token issuance (JWT), and profile management.
- **Technology**: **NestJS (Node.js)**
- **Reasoning**:
  - **High I/O**: Auth services handle a massive number of lightweight requests. Node.js event loop excels here.
  - **Ecosystem**: Excellent libraries (Passport.js) for OAuth/OIDC integration.
- **Database**: PostgreSQL (Relational data for users/roles).

### 2. Catalog Service (`/services/catalog`)

- **Role**: Manages products, categories, attributes, and pricing.
- **Technology**: **.NET 8/9 (C#)**
- **Reasoning**:
  - **Robustness**: C# provides strong typing and enterprise patterns (DDD) perfect for complex domain logic (product variations, hierarchical categories).
  - **Performance**: Extremely fast for read-heavy workloads.
- **Database**: MongoDB (Flexible schema for varying product attributes).

### 3. Inventory Service (`/services/inventory`)

- **Role**: Tracks stock levels, reserves items during checkout, handles "out of stock" logic.
- **Technology**: **Go (Golang)**
- **Reasoning**:
  - **High Concurrency**: The "Race Condition" problem (two users buying the last item) requires precise, high-performance locking or atomic updates. Go's goroutines and raw performance are ideal for this CPU/Concurrency-intensive task.
- **Database**: Redis (for fast, atomic decrement operations) + PostgreSQL (for durable ledger).

### 4. Order Service (`/services/orders`)

- **Role**: Orchestrates the checkout process, manages shopping carts, and order history.
- **Technology**: **NestJS (Node.js)**
- **Reasoning**:
  - **Orchestration**: Order processing involves talking to many other services (Payment, Inventory, Notification). Node.js is great at asynchronous orchestration (waiting for multiple promises).
- **Database**: PostgreSQL (Complex relational data for orders/line-items).

### 5. Payment Service (`/services/payment`)

- **Role**: Wraps payment gateway interactions (Stripe/PayPal), handles refunds, and ledger security.
- **Technology**: **.NET (C#)**
- **Reasoning**:
  - **Precision**: decimal types in C# are first-class for financial calculations.
  - **Security/Compliance**: Strong enterprise background.
- **Database**: PostgreSQL (ACID compliance is mandatory).

### 6. Recommendation API (`/services/recommendation`)

- **Role**: Provides "Products you might like" and chatbot support.
- **Technology**: **FastAPI (Python)**
- **Reasoning**:
  - **AI/ML**: Python is the lingua franca of AI. Access to PyTorch/TensorFlow/LangChain is native.
- **Database**: Vector Database (e.g., pgvector in Postgres or a dedicated one) for embeddings.

### 7. Notification Service (`/services/notification`)

- **Role**: Sends Emails, SMS, Push Notifications.
- **Technology**: **Go (Golang)**
- **Reasoning**:
  - **Lightweight Consumer**: It mainly listens to RabbitMQ queues and fires off requests. Go binaries are tiny and efficient for this "fire and forget" workload.

---

## Communication Patterns

### Asynchronous (Event-Driven) - using RabbitMQ

- **Events**:
  - `UserRegistered` -> Notification Service (Send Welcome Email)
  - `OrderPlaced` -> Email Service (Confirmation), Recommendation Service (Update User Profile), Inventory (Commit Reservation).
  - `PaymentFailed` -> Order Service (Cancel Order), Email Service (Alert User).

### Synchronous (gRPC / HTTP)

- **gRPC**:
  - Internal service-to-service communication where strict contracts and speed are needed.
  - _Example_: `OrderService` -> `InventoryService.ReserveStock()` (Need to know immediately if successful).
- **HTTP/REST**:
  - Public facing API Gateway (or BFF) to Frontend.

## Infrastructure Map

| Service                         | Language       | Port (Internal) | Database                  | Bus      |
| :------------------------------ | :------------- | :-------------- | :------------------------ | :------- |
| **Api Gateway** (Nginx/Traefik) | -              | 80/443          | -                         | -        |
| **Auth**                        | TS/NestJS      | 8001            | Postgres (DB: `auth`)     | RabbitMQ |
| **Catalog**                     | C#/.NET        | 8002            | Mongo (DB: `catalog`)     | RabbitMQ |
| **Orders**                      | TS/NestJS      | 8003            | Postgres (DB: `orders`)   | RabbitMQ |
| **Inventory**                   | Go             | 8004            | Redis + Postgres          | RabbitMQ |
| **Payment**                     | C#/.NET        | 8005            | Postgres (DB: `payments`) | RabbitMQ |
| **Notification**                | Go             | 8006            | -                         | RabbitMQ |
| **AI/ML**                       | Python/FastAPI | 8007            | Postgres (`pgvector`)     | RabbitMQ |

## Getting Started

1.  Each service has its own `Dockerfile`.
2.  Shared Protocol Buffers (`.proto`) are stored in `shared/protos` and generated for each language.
