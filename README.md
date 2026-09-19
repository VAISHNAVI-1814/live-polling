# Live Polling Application (Real-Time)

A high-performance, production-grade real-time live polling web application built for the **GUVI Developer Internship**. 

Users create polls, generate shareable links or codes, and audience members cast votes in real-time. Results update across **all connected browser windows instantly without requiring any page refresh**, powered by an asynchronous **Redis Pub/Sub** and **WebSocket** streaming pipeline backed by **MongoDB** for durable persistence and **Go (Gin)** for high-throughput concurrency.

---

## 🏗️ System Architecture & Data Flow

```
                      [ AUDIENCE VOTES ]
                              │
                              ▼
                     React 18 Frontend
                              │  POST /api/polls/:id/vote
                              ▼
                      Go (Gin) Backend
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
           [ Server Validation ]  [ Voter Token Check ]
                    │
                    ├───────────────────────────────────────┐
                    ▼                                       ▼
             MongoDB Database                         Redis Engine
        (Durable Storage & Audit)             (Fast In-Memory State)
        • Collections: users, polls, votes    • SADD poll:{id}:voters (O(1) duplicate check)
        • Persistent backup & indexing        • HINCRBY poll:{id}:votes {opt} 1 (Atomic tally)
                                              • PUBLISH poll:{id}:events (Pub/Sub Event)
                                                            │
                                                            ▼
                                                   Go WebSocket Hub
                                                            │
                                        ┌───────────────────┼───────────────────┐
                                        ▼                   ▼                   ▼
                                    Browser A           Browser B           Browser C
                                (Audience Voter)    (Audience Voter)    (Poll Creator)
                                        │                   │                   │
                                        └───────────────────┴───────────────────┘
                                         All screens update live without refresh!
```

---

## ⚡ Tech Stack & Roles

| Layer | Technology | Role & Why It Was Chosen |
|---|---|---|
| **Frontend** | **React 18 + Vite** | High-performance reactive UI with Tailwind CSS, animated progress bars, live indicators, and dynamic state management. |
| **Backend** | **Go (Gin Framework)** | Ultra-fast compiled API server with lightweight goroutines handling hundreds of concurrent HTTP and WebSocket connections efficiently. |
| **Database** | **MongoDB** | Schemaless document store handling user profiles, poll metadata, option structures, and durable voting audit records with unique indexes. |
| **Realtime** | **Redis (v9)** | **The core real-time engine**: Fast $O(1)$ atomic tallying (`HINCRBY`), instant duplicate voter prevention (`SADD`/`SISMEMBER`), and high-speed Pub/Sub messaging (`PUBLISH`/`SUBSCRIBE`). |
| **Sockets** | **Gorilla WebSocket** | Full-duplex persistent connections bridging Redis Pub/Sub events directly into the browser in milliseconds. |

---

## 📂 Project Structure

```
live-polling/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Application entrypoint & HTTP server bootstrap
│   ├── config/
│   │   └── config.go                # Environment variable configuration loader
│   ├── controllers/
│   │   ├── auth_controller.go       # Signup, Login, and Me handlers
│   │   └── poll_controller.go       # Create, Vote, Results, Share, and Status handlers
│   ├── middleware/
│   │   ├── auth_middleware.go       # JWT Bearer authentication verification
│   │   └── cors.go                  # Cross-Origin Resource Sharing handling
│   ├── models/
│   │   ├── user.go                  # User model & auth DTOs
│   │   ├── poll.go                  # Poll, PollOption, and Result models
│   │   └── vote.go                  # Vote document model
│   ├── repository/
│   │   ├── user_repo.go             # MongoDB queries & unique email index
│   │   └── poll_repo.go             # MongoDB queries for polls & vote tracking
│   ├── services/
│   │   ├── auth_service.go          # Bcrypt hashing & JWT token generation
│   │   ├── poll_service.go          # Business logic, input validation, expiration
│   │   ├── poll_service_test.go     # Automated unit tests for Redis & Auth
│   │   └── redis_service.go         # Redis atomic counters, Set voters, & Pub/Sub
│   ├── websocket/
│   │   ├── client.go                # WebSocket client read/write pumps & heartbeat
│   │   └── hub.go                   # Room manager dispatching Redis events to sockets
│   ├── go.mod                       # Go modules definition
│   ├── go.sum                       # Go checksums
│   ├── Dockerfile                   # Multi-stage production container build
│   └── .env.example                 # Sample environment variables
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Navbar.jsx           # Top navigation with live pulsing status & auth state
│   │   │   ├── ProtectedRoute.jsx   # Auth guard redirecting to /login
│   │   │   ├── LiveResultBar.jsx    # Real-time animated percentage bars & leader badge
│   │   │   └── ShareModal.jsx       # Share code & link copy modal with feedback
│   │   ├── context/
│   │   │   └── AuthContext.jsx      # Global auth state & persistent token management
│   │   ├── hooks/
│   │   │   └── usePollWebSocket.js  # Auto-reconnecting WebSocket hook for live updates
│   │   ├── pages/
│   │   │   ├── Login.jsx            # Sign In form
│   │   │   ├── Signup.jsx           # Account registration form
│   │   │   ├── Dashboard.jsx        # Creator dashboard (metrics, close/delete, share)
│   │   │   ├── CreatePoll.jsx       # Dynamic option adder & expiration selector
│   │   │   ├── VotePage.jsx         # Clean voter UI with duplicate prevention & confetti
│   │   │   └── ResultsPage.jsx      # Live real-time results projector view
│   │   ├── services/
│   │   │   └── api.js               # API service & persistent voter token generator
│   │   ├── App.jsx                  # Main router setup
│   │   ├── main.jsx                 # React root renderer
│   │   └── index.css                # Tailwind base styles & animations
│   ├── Dockerfile                   # Multi-stage frontend container build
│   ├── nginx.conf                   # Nginx reverse proxy for SPA, API, and WebSockets
│   ├── package.json                 # Frontend dependencies & scripts
│   ├── tailwind.config.js           # Tailwind design tokens
│   └── vite.config.js               # Vite build configuration
│
├── docker-compose.yml               # Complete orchestration (Mongo, Redis, API, UI)
├── .gitignore                       # Repository ignore rules
├── README.md                        # Documentation and architecture guide
└── VIDEO_SCRIPT.md                  # 3-5 minute presentation script for submission
```

---

## 🚀 Getting Started

You can run the entire stack with Docker Compose or natively on your local machine.

### Method 1: Instant Start with Docker Compose (Recommended)

Make sure Docker and Docker Compose are installed:

```bash
# Clone the repository
git clone <your-repo-url>
cd live-polling

# Start all 4 services (MongoDB, Redis, Go Backend, React Frontend)
docker-compose up --build
```

Once started:
- **Frontend App**: `http://localhost:3000`
- **Backend API**: `http://localhost:8080`
- **MongoDB**: `localhost:27017`
- **Redis**: `localhost:6379`

---

### Method 2: Native Local Setup

#### Prerequisites
- **Node.js**: v18 or higher (`node -v`)
- **Go**: v1.22 or higher (`go version`)
- **MongoDB**: Running on `mongodb://localhost:27017` (or MongoDB Atlas URI)
- **Redis**: Running on `localhost:6379` (or Upstash Redis URI)

#### 1. Backend Setup
```bash
cd backend

# Copy environment variables
cp .env.example .env

# Download dependencies
go mod download

# Run unit tests
go test -v ./services/...

# Start backend server
go run ./cmd/server
```
The backend starts on `http://localhost:8080`.

#### 2. Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Start Vite dev server
npm run dev
```
The frontend opens on `http://localhost:5173`.

---

## 🔒 Environment Variables

### Backend (`backend/.env`)
| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port for the Go Gin HTTP & WebSocket server |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection URI (or MongoDB Atlas) |
| `MONGO_DB_NAME` | `livepolling` | Name of the MongoDB database |
| `REDIS_URL` | `localhost:6379` | Redis host:port or `redis://...` connection URL |
| `REDIS_PASSWORD` | `""` | Optional password for Redis authentication |
| `JWT_SECRET` | `super-secure-key` | Secret key used for signing JWT tokens |
| `CLIENT_ORIGIN` | `http://localhost:5173` | Allowed frontend origin for CORS |

### Frontend (`frontend/.env`)
| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080/api` | REST API base endpoint |
| `VITE_WS_URL` | `ws://localhost:8080/ws` | WebSocket base endpoint |

---

## 📡 API Documentation

### Authentication Endpoints
- **POST `/api/auth/signup`**
  - Body: `{"name": "Alice", "email": "alice@example.com", "password": "password123"}`
  - Response: `201 Created` with `{ "token": "...", "user": { ... } }`
- **POST `/api/auth/login`**
  - Body: `{"email": "alice@example.com", "password": "password123"}`
  - Response: `200 OK` with `{ "token": "...", "user": { ... } }`
- **GET `/api/auth/me`** (Protected - Bearer Token)
  - Headers: `Authorization: Bearer <token>`
  - Response: `200 OK` with `{ "user": { ... } }`

### Poll Endpoints
- **POST `/api/polls`** (Protected - Bearer Token)
  - Body:
    ```json
    {
      "question": "What is your favorite programming language?",
      "options": ["Go", "TypeScript", "Python", "Rust"],
      "expires_in_minutes": 1440
    }
    ```
  - Response: `201 Created` with poll object and unique 6-character `share_code`.
- **GET `/api/polls/my-polls`** (Protected - Bearer Token)
  - Lists all polls created by the authenticated user.
- **GET `/api/polls/:id`** (Public)
  - Retrieves poll metadata.
- **GET `/api/polls/share/:shareCode`** (Public)
  - Retrieves poll details by its 6-character public share code (e.g. `ABC123`).
- **PATCH `/api/polls/:id/status`** (Protected)
  - Body: `{"status": "closed"}` (or `"active"`). Closes or reopens the poll.
- **DELETE `/api/polls/:id`** (Protected)
  - Deletes the poll and cleans up associated vote records.

### Voting & Real-Time Endpoints
- **POST `/api/polls/:id/vote`** (Public)
  - Body: `{"option_id": "opt_xxx", "voter_token": "unique-client-uuid"}`
  - Validates:
    1. Poll exists and is active.
    2. Poll has not expired.
    3. Option belongs to the poll.
    4. Client has not already voted (`409 Conflict` if duplicate).
  - Performs atomic `HINCRBY` in Redis and persists to MongoDB.
  - Publishes `VOTE_UPDATED` event to Redis Pub/Sub.
- **GET `/api/polls/:id/results`** (Public)
  - Returns current vote breakdown and percentages computed from Redis cache.
- **GET `/ws/polls/:id`** (WebSocket)
  - Upgrades HTTP connection to WebSocket.
  - Joins the poll's live broadcast room.
  - Automatically receives live JSON payloads whenever any user votes.

---

## 🏎️ How Redis Drives the Live Updates

Redis is not an afterthought in this application; it is the **backbone** of the real-time voting experience:

1. **Atomic In-Memory Counters (`HINCRBY`)**:
   - Each poll has a Redis Hash: `poll:<poll_id>:votes`.
   - Fields correspond to option IDs, and values are current counts.
   - When a vote is cast, `HINCRBY` increments the option count in $<1\text{ms}$, avoiding database lock contention.
2. **Instant Duplicate Prevention (`SADD` / `SISMEMBER`)**:
   - Each poll maintains a Redis Set: `poll:<poll_id>:voters`.
   - Adding a voter token via `SADD` is atomic ($O(1)$). If `SADD` returns `0`, the voter already cast a vote and the request is immediately rejected before touching disk.
3. **Event Multicasting (`PUBLISH` / `SUBSCRIBE`)**:
   - Once the counter is updated, the server publishes the updated poll snapshot to Redis channel `poll:<poll_id>:events`.
   - The Go WebSocket Hub subscribes to this channel and pushes the updated data frame across all open browser WebSocket connections simultaneously.

---

## 🧪 Multi-Browser Testing (Verification)

To verify the core requirement (**No page refresh required**):

1. **Step 1**: Open Browser 1 (e.g., Chrome). Log in, click **Create Poll**, enter:
   - Question: *"Which cloud provider do you prefer?"*
   - Options: `AWS`, `Google Cloud`, `Azure`
   - Click **Publish Live Poll**.
2. **Step 2**: You are redirected to the **Live Results Page** (`/poll/:id/results`). Notice the pulsing green dot: `Live Stream Active`.
3. **Step 3**: Click **Share Link** and copy the audience voting URL (e.g., `http://localhost:5173/poll/XYZ789`).
4. **Step 4**: Open **Browser 2 (Incognito or Firefox)** and paste the voting link.
5. **Step 5**: Open **Browser 3 (Mobile or another window)** and paste the same voting link.
6. **Step 6**: In Browser 2, select `Google Cloud` and click **Cast Live Vote**.
7. **Observe**: **Browser 1 (Results Page) and Browser 3 immediately update their vote counts and progress bars without any manual refresh!**

---

## ☁️ Production Deployment Guide

This project is built to deploy on free cloud tiers:

### 1. Database & Cache Services
- **MongoDB**: Create a free M0 cluster on [MongoDB Atlas](https://www.mongodb.com/atlas). Copy the connection URI.
- **Redis**: Create a free serverless Redis database on [Upstash](https://upstash.com). Copy the `redis://...` endpoint.

### 2. Backend Deployment (Render / Railway / Fly.io)
- Deploy `backend/` as a Go Web Service.
- Set environment variables:
  - `PORT=8080`
  - `MONGO_URI=<your-atlas-uri>`
  - `REDIS_URL=<your-upstash-redis-uri>`
  - `JWT_SECRET=<random-secret>`
  - `CLIENT_ORIGIN=https://<your-frontend>.vercel.app`

### 3. Frontend Deployment (Vercel / Netlify)
- Deploy `frontend/` as a Single Page Application.
- Set build command: `npm run build`
- Output directory: `dist`
- Set environment variables:
  - `VITE_API_URL=https://<your-backend-service>.onrender.com/api`
  - `VITE_WS_URL=wss://<your-backend-service>.onrender.com/ws`

---

## 💡 Key Engineering Challenges & Solutions

### 1. Distributed Real-Time Consistency Under Concurrent Votes
- **Challenge**: If multiple voters click simultaneously, standard database read-modify-write queries create race conditions and inaccurate vote counts.
- **Solution**: We offload atomic increments to Redis using `HINCRBY`. Redis runs commands single-threaded in memory, guaranteeing serializable, race-condition-free counting at sub-millisecond latency. MongoDB is updated asynchronously for durable persistence.

### 2. Dual-Layer Duplicate Vote Prevention
- **Challenge**: Preventing double-voting without imposing latency or database connection spikes.
- **Solution**: We created a dual-layer strategy:
  1. An in-memory Redis Set (`SADD poll:<id>:voters`) checks membership in $O(1)$ time.
  2. A MongoDB unique compound index `(poll_id, voter_token)` acts as the persistent authority.
  3. The frontend stores a persistent client-fingerprint token in `localStorage`.

### 3. Graceful WebSocket Fallback & Reconnection
- **Challenge**: Network hiccups or server restarts can sever WebSocket connections.
- **Solution**: The `usePollWebSocket` React hook implements an exponential backoff auto-reconnect strategy. If a disconnection occurs, the client automatically attempts reconnection every 3 seconds while showing an indicator in the UI.

---

## 🤖 AI Usage Transparency

In accordance with the GUVI internship evaluation criteria:
- **AI Tool Used**: Antigravity AI Coding Assistant.
- **How AI Helped**: AI was instrumental in accelerating boilerplate generation (such as initial model struct tags, Tailwind layout styling, and Docker multi-stage configuration), allowing focus on the core real-time architecture, Redis atomic counter concurrency, and WebSocket hub synchronization.
- **Key Takeaways & Mastery**: The foundational distributed system concepts—Redis atomic operations, Pub/Sub message dispatch, gorilla/websocket pump goroutines, and JWT security flows—were thoroughly reviewed and tested to ensure deep technical understanding for the technical interview rounds.
