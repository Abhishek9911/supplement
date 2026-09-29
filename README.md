# Suppliment Commerce Platform

A production-oriented e-commerce platform for industrial, pharmaceutical, and nutraceutical products. The solution combines a Next.js storefront with a Go + Gin backend and PostgreSQL persistence.

## Architecture overview

- Frontend: Next.js + TypeScript + Tailwind CSS
- Backend API: Go + Gin
- Database: PostgreSQL 16
- Auth: JWT access tokens and refresh-token flow
- RBAC: ADMIN and CUSTOMER roles enforced on the server
- Storage: S3-compatible endpoints can be configured via environment variables
- Deployment: Docker Compose

## Database relationship summary

- `users` stores credentials and role metadata.
- `categories` group products for browsing.
- `products` store catalog items and stock thresholds.
- `product_images` link many images to one product.
- `carts` and `cart_items` represent the authenticated user's cart.
- `orders` and `order_items` record customer purchase history with a snapshot of product name and price.
- `inventory_movements` tracks stock changes and administrative adjustments.
- `audit_logs` records high-value admin actions.
- `refresh_tokens` supports secure refresh-token handling.

## Project structure

- `frontend/` — storefront and admin UI
- `backend/` — Go REST API and PostgreSQL schema
- `backend/migrations/` — initial database migration and seed SQL
- `docker-compose.yml` — local orchestration
- `README.md` — setup and deployment guide
- `.env.example` — environment template

## Prerequisites

- Docker and Docker Compose
- Node.js 20+
- Go 1.23+
- PostgreSQL client (optional for manual inspection)

## Quick start

1. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
2. Start the stack:
   ```bash
   docker compose up --build
   ```
3. The API is available at `http://localhost:8080/api/v1`.
4. The frontend is available at `http://localhost:3000`.
5. Open the docs file at `backend/docs/openapi.yaml` or serve it via your API gateway.

## Admin bootstrap

Create the initial admin user with environment variables:

```bash
cd backend
ADMIN_NAME="System Administrator" ADMIN_EMAIL="admin@suppliment.local" ADMIN_PASSWORD="Admin@12345" go run ./cmd/admin
```

## Environment configuration

- `DATABASE_URL` for PostgreSQL connectivity
- `JWT_SECRET` for signing tokens
- `JWT_ACCESS_TTL` and `JWT_REFRESH_TTL` for token lifetimes
- `CORS_ALLOWED_ORIGINS` for allowed frontend origins
- `AWS_*` and `S3_ENDPOINT` values when you enable image uploads to object storage

## Test strategy

- Go unit tests for hashing, auth, order validation, and status rules
- API tests for customer and admin access rules
- PostgreSQL-backed integration checks for stock deduction and restoration
- E2E checks for cart and checkout paths when running the local stack

## Production notes

- TLS termination should be handled by a reverse proxy or cloud load balancer.
- Never commit real secrets.
- Keep database credentials and JWT secrets in environment management tooling or a secret store.
- For object storage, ensure bucket-level policies and IAM roles restrict access to the application only.
