# SaltyBytes Dashboard

Operational dashboard for the SaltyBytes recipe platform -- surfaces metrics, cost tracking, and data health checks across AI providers, caching layers, recipe pipelines, and user subscriptions via a read-only connection to the production database.

---

## Screenshots

| Overview | Cost Center |
|----------|-------------|
| ![Overview](docs/screenshots/overview.png) | ![Cost Center](docs/screenshots/cost-center.png) |

| Recipes | Search & Cache |
|---------|----------------|
| ![Recipes](docs/screenshots/recipes.png) | ![Search & Cache](docs/screenshots/search-cache.png) |

| Users & Subscriptions | Allergens & Families |
|-----------------------|----------------------|
| ![Users & Subscriptions](docs/screenshots/users.png) | ![Allergens & Families](docs/screenshots/allergens.png) |

| Infrastructure | Data Quality |
|----------------|--------------|
| ![Infrastructure](docs/screenshots/infrastructure.png) | ![Data Quality](docs/screenshots/data-quality.png) |

---

## Features

- **8 dashboard pages** -- Overview, Cost Center, Recipes, Search & Cache, Users & Subscriptions, Allergens & Families, Infrastructure, Data Quality
- **Real-time metrics** with an in-memory cache that refreshes automatically and on demand
- **Cost tracking** -- estimated daily/monthly spend by AI provider and feature, cost-per-user, cost-per-recipe, and cache savings calculations
- **Rate card editor** -- adjust per-token, per-image, and fixed-cost rates directly in the UI; changes persist to disk and trigger an immediate cost recalculation
- **Health checks** -- monitors database connectivity and service-level indicators
- **Search and canonical cache hit-rate monitoring** with savings estimates
- **Firecrawl credit usage tracking** against monthly budget
- **Password-protected** -- session-based authentication with a configurable dashboard password
- **Single binary deployment** -- the React frontend is embedded into the Go binary at build time

---

## Quick Start

1. Copy the example environment file and fill in your values:

   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your PostgreSQL connection string and a dashboard password:

   ```
   DATABASE_URL=postgres://dashboard_ro:password@your-rds-endpoint:5432/saltybytes?sslmode=require
   DASHBOARD_PASSWORD=a-strong-password
   ```

3. Start the container:

   ```bash
   docker compose up -d
   ```

4. Open `http://localhost` in your browser and log in with the password you set.

---

## Unraid

An Unraid Community Applications XML template is included at `unraid/saltybytes-dashboard.xml`.

To install manually:

1. In the Unraid web UI, go to **Docker > Add Container > Template Repositories**.
2. Add the template URL:
   ```
   https://raw.githubusercontent.com/saltybytes/saltybytes-dashboard/main/unraid/saltybytes-dashboard.xml
   ```
3. Click **Add**, then find **saltybytes-dashboard** in the template list.
4. Fill in the **Database URL** and **Dashboard Password** fields, then click **Apply**.

The container image is pulled from `ghcr.io/saltybytes/saltybytes-dashboard:latest`. Persistent data (rate card configuration) is stored at `/mnt/user/appdata/saltybytes-dashboard` by default.

---

## Configuration

| Variable             | Required | Default | Description                                                        |
|----------------------|----------|---------|--------------------------------------------------------------------|
| `DATABASE_URL`       | Yes      | --      | PostgreSQL connection string for the read-only dashboard user      |
| `DASHBOARD_PASSWORD` | Yes      | --      | Password to access the dashboard web UI                            |
| `PORT`               | No       | `80`    | Port the HTTP server listens on                                    |
| `DATA_DIR`           | No       | `/data` | Directory for persistent files (rate card JSON, etc.)              |

---

## Database Setup

The dashboard requires only **read** access. Create a dedicated read-only PostgreSQL user before connecting:

```sql
-- Create the read-only role
CREATE ROLE dashboard_ro WITH LOGIN PASSWORD 'a-strong-password';

-- Grant connect and usage
GRANT CONNECT ON DATABASE saltybytes TO dashboard_ro;
GRANT USAGE ON SCHEMA public TO dashboard_ro;

-- Grant SELECT on all existing tables
GRANT SELECT ON ALL TABLES IN SCHEMA public TO dashboard_ro;

-- Ensure future tables are also readable
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO dashboard_ro;
```

Use the resulting connection string as `DATABASE_URL`:

```
postgres://dashboard_ro:a-strong-password@your-host:5432/saltybytes?sslmode=require
```

---

## Development

### Prerequisites

- Go 1.23+
- Node.js 20+
- A PostgreSQL database with the SaltyBytes schema

### Run locally

1. Start the Vite dev server (with hot reload):

   ```bash
   cd frontend
   npm ci --legacy-peer-deps
   npm run dev
   ```

   The dev server runs on `http://localhost:5173` by default.

2. In a separate terminal, start the Go backend:

   ```bash
   export DATABASE_URL="postgres://dashboard_ro:password@localhost:5432/saltybytes?sslmode=disable"
   export DASHBOARD_PASSWORD="dev"
   export PORT="3000"
   go run ./cmd/dashboard
   ```

   The Vite dev server proxies `/api` requests to this address (configured in `vite.config.ts`).

---

## Architecture

```
Browser  -->  Go HTTP server (:80)
                |
                |-- /api/*        -->  JSON handlers (auth, metrics, rate card)
                |-- /*            -->  Embedded React SPA (static files via go:embed)
                |
                +-- MetricCache   -->  In-memory cache, refreshed on a timer + on-demand
                |     |
                |     +-- GORM    -->  Read-only PostgreSQL queries
                |
                +-- RateCard      -->  Editable cost rates, persisted to DATA_DIR/ratecard.json
```

The Go binary embeds the compiled React frontend at build time using `go:embed`. At runtime it opens a single read-only database connection, populates an in-memory metric cache, and serves both the SPA and JSON API from one process. There is no write path to the database -- all mutations (rate card edits) are stored on the local filesystem.

---

## Tech Stack

| Layer    | Technology                        |
|----------|-----------------------------------|
| Backend  | Go, GORM, PostgreSQL (read-only)  |
| Frontend | React, TypeScript, Tailwind CSS, Recharts, Tremor |
| Tooling  | Vite, Docker, GitHub Actions      |
| Registry | GitHub Container Registry (ghcr.io) |

---

## License

See [LICENSE](LICENSE) for details.
