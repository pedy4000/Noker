# Noker (Note + Taker)

**Backend-first system to automate meeting note processing and extract Opportunities (OST)**

**Automatically turn raw customer conversations into a clean, deduplicated, evidence-rich Opportunity Solution Tree (OST)**  
Built for product teams who refuse to lose signal in noise.

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql&logoColor=white)](https://postgresql.org)
![Status: Production Ready](https://img.shields.io/badge/status-production%20ready-brightgreen)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Core Features](#core-features)
3. [Roadmap](#roadmap)
4. [Architecture](#architecture)
5. [Technical Decisions & Trade-offs](#technical-decisions--trade-offs)
6. [Setup & Installation](#setup--installation)
7. [Running the System](#running-the-system)
8. [API Endpoints & Sample Requests](#api-endpoints--sample-requests)
9. [Sample Input & Output](#sample-input--output)
10. [Testing](#testing)
11. [Notes](#notes)

---

## Project Overview

`Noker` is a backend-first MVP that:

* Receives meeting notes (manual or simulated Notion input)
* Cleans and normalizes text
* Extracts Opportunities following OST structure
* Stores data in a database (PostgreSQL)
* Processes AI tasks asynchronously via a queue
* Provides API endpoints and Slack-like commands

**Goal:** Reduce manual effort in turning raw meeting notes into structured Opportunities and evidence.

---

## Core Features

| Feature                        | Status  | Description |
|-------------------------------|--------|-----------|
| Async meeting ingestion       | Done    | POST → queued → processed |
| Intelligent deduplication    | Done    | Same job-to-be-done → merged |
| Evidence preservation         | Done    | Every quote linked to source |
| Theme auto-categorization     | Done    | 6+ canonical themes |
| Real-time graph API           | Done    | `/opportunities/recent` |
| In-memory queue (swapable)    | Done    | Kafka-ready interface |
| Multi-mode deployment         | Done    | API-only, Worker-only, or combined |
| test coverage                 | Done    | Including in-memory SQLite integration tests |

---
## Roadmap

- Slack bot integration (`/ost add`)
- Notion → Noker sync
- Zoom transcript auto-ingestion
- Export to Linear / Jira / Notion
- Confidence scoring per opportunity
- Web dashboard with filtering & search
- Opportunity impact tagging (revenue / retention / acquisition)
- On-premise / private cloud deployment (enterprise)
- Sentiment & urgency detection
- Opportunity prioritization matrix (RICE / ICE)
- API webhooks & change notifications

More coming soon...

---

## Architecture


```text
                     ┌─────────────────┐
        Clients ───► │   API Server    │
                     │   (Gin/Fiber)   │
                     └───────┬─────────┘
                             │
                     ┌──────▼───────┐
                     │   Job Queue   │ ← In-memory (channel-based)
                     └──────┬───────┘
        ┌──────────────────┘   └──────────────────┐
        ▼                                         ▼
┌───────────────┐                          ┌───────────────┐
│  AI Worker    │  gRPC / HTTP → OpenAI    │  OST-GPT      │
│  (Extractor)  │ ◄──────────────────────►│  Reasoning    │
└───────┬───────┘                          └───────┬───────┘
        ▼                                          ▼
┌───────────────┐                          ┌───────────────┐
│ PostgreSQL    │ ◄──────────────────────► │ Opportunities │
│ meetings      │   sqlc + goose           │ Evidence      │
│ opportunities │                          │ Themes        │
└───────────────┘                          └───────────────┘
```

### Key Components:

* **API Server:** Accepts meetings, provides endpoints, simulates Slack commands.
* **Job Queue:** Handles asynchronous processing of meetings.
* **AI Worker:** Processes queue jobs to normalize text and extract opportunities.
* **Database (PostgreSQL):** Stores meetings, opportunities, evidence, and themes.

---

## Technical Decisions & Trade-offs

### 1. **Golang**

* Chosen for **performance, concurrency, and simplicity** in backend services.
* Native support for HTTP servers, goroutines, and channels fits the async processing requirement.

### 2. **PostgreSQL**

* Strong support for structured queries, joins, and indexing.
* Compatible with `sqlc` to generate **type-safe Go queries**.

### 3. **sqlc**

* Generates type-safe Go code from SQL queries.
* Reduces runtime errors and ensures strong typing.
* Easier maintenance and refactoring of SQL queries.

### 4. **Goose for Migrations**
* Used for database schema management.
* Tracks migration history and ensures database consistency across environments.
* Supports multiple environments (e.g., dev, test, prod) using config files.
* Makes rolling back and re-applying migrations simple.

### 5. **In-memory Queue**

* Chosen over Kafka or RabbitMQ because the workload is **small (1k–10k meetings/day)**.
* Avoids setup complexity of external brokers.
* In case of server crash or failed processing, **failed meetings can be re-queued from the DB**.
* Extensible: queue is an interface; switching to Kafka or any other queue only requires implementing **3 functions**.

### 6. **AI Provider**

* GPT chosen for familiarity, comfort, and extensive documentation.
* AI is abstracted via an interface; other providers can be plugged in easily.

### 7. **Deployment Flexibility**

* Can run as **API-only** (accept meetings, enqueue jobs).
* Can run as **AI-only** (process queue, extract opportunities).
* Or run **combined** on a single server.
* Enables scenarios like:

  * API on one server for light CPU usage
  * AI on another server for security like ip limiting and resource-intensive processing

### 8. **Other Design Choices**

* **Middleware-based authentication** (simple API Key).
* **Structured logging** with clear info/warn/error levels.
* **Docker-ready** for reproducible environments.
* **Extensible architecture:** new input sources, AI providers, or queue backends can be added without touching core logic.

---

## Setup & Installation

### Requirements

* Go 1.21+
* Docker & Docker Compose (optional, recommended)
* PostgreSQL (or use Docker service)
* [`Make`](https://makefiletutorial.com/) for running predefined commands
* [`goose`](https://github.com/pressly/goose) for migrations
* [`sqlc`](https://sqlc.dev/) for generating type-safe Go queries

### 1. Clone Repository

```bash
git clone https://github.com/pedy4000/noker.git
cd noker
```

### 2. Install Go Libraries

```bash
go mod tidy
go get ./...
```

### 3. Environment Variables

Create `.env` file:

```env
OPENAI_API_KEY=your-openai-api-key
```

### 4. Database Setup

#### Using Docker Compose

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: noker-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: noker
      POSTGRES_PASSWORD: noker
      POSTGRES_DB: noker
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

Run:

```bash
docker-compose up -d postgres
```
Or with make command

```bash
make postgres-up
```

Run migrations:

```bash
DATABASE_URL ?= postgres://noker:noker@postgres:5432/noker?sslmode=disable
goose -dir migrations postgres "$(DATABASE_URL)" up
```
Or with make command
```bash
make migrate-up
```

---

## Running the System

### 1. Run API + AI in one instance

```bash
go build -trimpath -ldflags="-s -w" -o bin/noker ./cmd/noker
./bin/noker
```
Or with make command
```bash
make run
```

### 2. Run API only

* Disable AI processing in `config.yaml`.
* Build the project

### 3. Run AI extractor only

* Disable Server in `config.yaml`.
* Build the project

### 4. Run with Docker Compose

```bash
docker-compose up --build
```
Or with make command

```bash
make run-prod
```
Or if you want to see demo

```bash
make demo-up
```

---

## API Endpoints & Sample Requests

### Create Meeting

```bash
curl -X POST http://localhost:8080/api/meetings \
  -H "Content-Type: application/json" \
  -H "X-API-Key: noker-dev-key-2025" \
  -d '{
        "title": "Acme Corp – Export Hell",
        "notes": "Every Monday we spend 3 hours cleaning exported CSVs...",
        "source": "manual",
        "metadata": {"customer": "Acme Corp", "plan": "enterprise"}
      }'
```

**Response:**

```json
{
  "id": "uuid-1234",
  "status": "queued"
}
```

### List Recent Opportunities

```bash
curl -X GET http://localhost:8080/api/opportunities/recent?include_evidence=true \
  -H "X-API-Key: noker-dev-key-2025"
```

**Response:**

```json
[
  {
    "id": "op-123",
    "user_segment": "enterprise",
    "struggle": "CSV exports get corrupted",
    "evidence": [
      {"quote": "Persian text becomes garbage", "meeting_id": "uuid-1234"},
      {"quote": "Fonts broken again", "meeting_id": "uuid-5678"}
    ],
    "theme": "export-issues",
    "created_at": "2025-12-03T14:00:00Z"
  }
]
```

---

## Sample Input & Output

**Input (Meeting Notes):**

```
Title: Acme Corp – Export Hell
Notes: Every Monday we spend 3 hours cleaning exported CSVs. Persian text becomes garbage, dates are wrong, columns missing.
```

**Output (Opportunity Extracted):**

```json
{
  "struggle": "CSV exports get corrupted",
  "evidence": [
    "Persian text becomes garbage"
  ],
  "theme": "export-issues"
}
```

---

## Testing

```bash
make test
```

Tests cover:

* Full end-to-end flow: meeting creation → queue → AI extraction → opportunity creation
* Deduplication of opportunities
* Evidence linkage
* Slack-like command responses

---

## Notes

* System designed for **high extensibility** (queue backends, AI providers, input sources).
* Async queue allows **scaling independently** of API or AI processing.
* Clear separation of concerns ensures **production readiness**, maintainability, and observability.
