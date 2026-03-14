# Catalog Service

ASP.NET Core Web API (.NET 10) that manages the product catalog for the RTS platform. It owns the `products`, `categories`, and `brands` collections in MongoDB and exposes a REST API consumed by the frontend and other microservices.

**Key capabilities:**

- Full CRUD for products, categories, and brands
- Full-text product search with filtering, sorting, and pagination
- Hierarchical category tree (materialized-path model)
- Product image upload and management via MinIO (S3-compatible)
- Redis caching for hot read paths
- Domain event publishing to RabbitMQ on state changes
- JWT bearer authentication shared with the auth service
- Health checks for MongoDB, Redis, and RabbitMQ

---

## Table of Contents

1. [Tech Stack](#tech-stack)
2. [Architecture](#architecture)
3. [Domain Entities](#domain-entities)
4. [API Endpoints](#api-endpoints)
   - [Products](#products-apiproducts)
   - [Categories](#categories-apicategories)
   - [Brands](#brands-apibrands)
   - [Health](#health-health)
5. [Configuration](#configuration)
6. [Development Setup](#development-setup)
7. [Caching Strategy](#caching-strategy)
8. [Event Publishing](#event-publishing)
9. [Image Upload](#image-upload)
10. [MongoDB Indexes](#mongodb-indexes)

---

## Tech Stack

| Concern | Technology |
|---|---|
| Runtime | .NET 10 (ASP.NET Core) |
| Database | MongoDB 8.x (driver v3) |
| Cache | Redis (StackExchange.Redis + `IDistributedCache`) |
| Object storage | MinIO via AWS SDK for .NET (S3-compatible) |
| Messaging | RabbitMQ (RabbitMQ.Client v7) |
| Auth | JWT Bearer (`Microsoft.AspNetCore.Authentication.JwtBearer`) |
| API docs | Swagger / Swashbuckle |
| Health checks | `AspNetCore.HealthChecks.*` (Mongo, Redis, RabbitMQ) |

---

## Architecture

The service follows Clean Architecture with four logical layers inside `Catalog.Api/`:

```
Catalog.Api/
├── Controllers/          # HTTP layer — routes, authorization, request binding
├── Application/
│   ├── DTOs/             # Request and response shapes
│   └── Handlers/         # One handler class per use-case (no MediatR)
├── Domain/
│   ├── Entities/         # Product, Category, Brand + value objects
│   ├── Events/           # Domain events (IDomainEvent)
│   ├── Interfaces/       # Repository contracts
│   └── Models/           # Search criteria models
└── Infrastructure/
    ├── Repositories/     # MongoDB repository implementations
    ├── Messaging/        # RabbitMQ event publisher
    ├── Services/         # Redis cache service, MinIO storage service
    ├── Settings/         # Strongly-typed configuration POCOs
    ├── Extensions/       # ServiceExtensions (DI wiring), SwaggerExtension
    ├── CatalogUnitOfWork.cs
    └── MongoIndexesInitializer.cs
```

**Request flow:**

```
HTTP Request
  └─> Controller
        └─> Handler (use-case)
              ├─> CatalogUnitOfWork (repository access)
              ├─> ICacheService (Redis read/write/invalidate)
              ├─> IImageStorageService (MinIO uploads)
              └─> ICatalogEventPublisher (RabbitMQ)
```

**`CatalogUnitOfWork`** aggregates the three repositories (`Products`, `Categories`, `Brands`) and is injected as a single dependency into handlers that need cross-collection access.

**`MongoIndexesInitializer`** is a hosted service that runs on startup and ensures all required MongoDB indexes exist.

---

## Domain Entities

### Product

Stored in the `products` collection.

| Field | Type | Notes |
|---|---|---|
| `Id` | `string` (UUID v7) | Primary key |
| `Sku` | `string` | Unique, max 100 chars |
| `Slug` | `string` | Unique URL slug, max 120 chars |
| `Title` | `string` | Max 200 chars |
| `Description` | `string?` | Max 2000 chars |
| `BrandId` | `string?` | Reference to a Brand |
| `CategoryIds` | `string[]` | References to Categories |
| `Status` | `ProductStatus` | `Draft` \| `Active` \| `Archived` |
| `Price` | `Money` | `Amount` (decimal) + `Currency` (3-char ISO) |
| `Variants` | `ProductVariant[]` | Each has its own SKU, attributes, and optional price |
| `Images` | `ProductImage[]` | Ordered by `SortOrder`; first is `IsPrimary` |
| `AverageRating` | `decimal?` | Maintained externally |
| `ReviewCount` | `int` | Maintained externally |
| `Tags` | `string[]` | Free-form tags |
| `CreatedAt` | `DateTime` | UTC |
| `UpdatedAt` | `DateTime` | UTC, updated on every `Touch()` |
| `Version` | `int` | Incremented on every write; used for optimistic concurrency |

Product lifecycle: `Draft` → `Active` → `Archived`. The `DELETE /api/products/{id}` endpoint performs a soft delete by transitioning the product to `Archived`.

### Category

Stored in the `categories` collection. Supports unlimited nesting via the **materialized-path** pattern.

| Field | Type | Notes |
|---|---|---|
| `Id` | `string` (UUID v7) | Primary key |
| `Name` | `string` | Max 100 chars |
| `Slug` | `string` | Unique |
| `Description` | `string?` | Max 500 chars |
| `ParentId` | `string?` | `null` for root categories |
| `Path` | `string[]` | Ancestor IDs from root to parent (materialized path) |
| `Depth` | `int` | 0 for root |
| `SortOrder` | `int` | Sibling ordering |
| `IsActive` | `bool` | Default `true` |
| `ImageUrl` | `string?` | Optional category image |
| `CreatedAt` / `UpdatedAt` | `DateTime` | UTC |

### Brand

Stored in the `brands` collection.

| Field | Type | Notes |
|---|---|---|
| `Id` | `string` (UUID v7) | Primary key |
| `Name` | `string` | Max 100 chars |
| `Slug` | `string` | Unique |
| `Description` | `string?` | Max 500 chars |
| `LogoUrl` | `string?` | URL to brand logo |
| `Website` | `string?` | Brand website URL |
| `IsActive` | `bool` | Default `true` |
| `CreatedAt` / `UpdatedAt` | `DateTime` | UTC |

---

## API Endpoints

All mutating endpoints require an `Authorization: Bearer <token>` header (or an `access_token` cookie) carrying a JWT issued by the auth service with `role: admin`.

Error responses follow the shape `{ "code": "ERROR_CODE", "message": "Human-readable description." }`.

---

### Products `/api/products`

#### `POST /api/products`

Create a new product. Requires `Admin` policy.

**Request body:**

```json
{
  "sku": "SKU-001",
  "slug": "my-product",
  "title": "My Product",
  "description": "Optional description",
  "amount": 29.99,
  "currency": "USD",
  "brandId": "<brand-id>",
  "categoryIds": ["<category-id>"]
}
```

| Field | Required | Constraints |
|---|---|---|
| `sku` | Yes | max 100 chars, must be unique |
| `slug` | Yes | max 120 chars, must be unique, stored lowercase |
| `title` | Yes | max 200 chars |
| `description` | No | max 2000 chars |
| `amount` | Yes | >= 0.01 |
| `currency` | Yes | exactly 3 chars (ISO 4217), stored uppercase |
| `brandId` | No | must reference an existing brand |
| `categoryIds` | No | each must reference an existing category |

**Response `201 Created`:** `ProductResponseDto` (see below).

**Error codes:** `DUPLICATE_SKU`, `DUPLICATE_SLUG`, `BRAND_NOT_FOUND`, `CATEGORY_NOT_FOUND`.

---

#### `GET /api/products/{id}`

Get a product by its ID. Public.

**Response `200 OK`:** `ProductResponseDto`.

**Response `404 Not Found`.**

---

#### `GET /api/products/slug/{slug}`

Get a product by its URL slug. Public.

**Response `200 OK`:** `ProductResponseDto`.

**Response `404 Not Found`.**

---

#### `GET /api/products`

Search and list products with filtering and pagination. Public.

**Query parameters:**

| Parameter | Type | Default | Description |
|---|---|---|---|
| `query` | `string` | — | Full-text search across title (weight 10) and description (weight 1) |
| `categoryId` | `string` | — | Filter by category; automatically includes all descendant categories |
| `brandId` | `string` | — | Filter by brand |
| `minPrice` | `decimal` | — | Minimum price (inclusive) |
| `maxPrice` | `decimal` | — | Maximum price (inclusive) |
| `status` | `string` | — | `Draft`, `Active`, or `Archived` |
| `sortBy` | `string` | — | Field to sort by |
| `page` | `int` | `1` | Page number |
| `pageSize` | `int` | `20` | Items per page, capped at 100 |

**Response `200 OK`:**

```json
{
  "items": [ /* ProductResponseDto[] */ ],
  "totalCount": 120,
  "page": 1,
  "pageSize": 20,
  "totalPages": 6
}
```

---

#### `PUT /api/products/{id}`

Update a product. Requires `Admin` policy.

Uses optimistic concurrency: the `version` field in the request body must match the current document version.

**Request body:**

```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "brandId": "<brand-id>",
  "categoryIds": ["<category-id>"],
  "amount": 39.99,
  "currency": "USD",
  "tags": ["sale", "new"],
  "version": 3
}
```

**Response `200 OK`:** `ProductResponseDto`.

**Error codes:** `PRODUCT_NOT_FOUND`, `BRAND_NOT_FOUND`, `CATEGORY_NOT_FOUND`.

---

#### `DELETE /api/products/{id}`

Soft-delete (archive) a product. Requires `Admin` policy.

Sets `status` to `Archived`. The product record is retained in MongoDB.

**Response `204 No Content`.**

---

#### `PATCH /api/products/{id}/status`

Change the lifecycle status of a product. Requires `Admin` policy.

**Request body:**

```json
{ "status": "Active" }
```

Valid values: `Draft`, `Active`, `Archived`.

**Response `200 OK`:** `ProductResponseDto`.

---

#### `POST /api/products/{id}/images`

Upload an image to a product. Requires `Admin` policy.

**Request:** `multipart/form-data`

| Field | Type | Notes |
|---|---|---|
| `file` | `IFormFile` | See [Image Upload](#image-upload) for constraints |
| `altText` | `string` (form field) | Optional alt text |

**Response `200 OK`:** `ProductResponseDto` with the updated `images` array.

---

#### `DELETE /api/products/{id}/images/{imageId}`

Remove an image from a product. Requires `Admin` policy.

**Response `200 OK`:** `ProductResponseDto` with the updated `images` array.

---

#### `PUT /api/products/{id}/images/order`

Reorder product images. Requires `Admin` policy.

**Request body:**

```json
{
  "order": [
    { "imageId": "<id>", "sortOrder": 0 },
    { "imageId": "<id>", "sortOrder": 1 }
  ]
}
```

**Response `200 OK`:** `ProductResponseDto`.

---

#### `ProductResponseDto` shape

```json
{
  "id": "...",
  "sku": "SKU-001",
  "slug": "my-product",
  "title": "My Product",
  "description": "...",
  "amount": 29.99,
  "currency": "USD",
  "status": "Draft",
  "brandId": null,
  "categoryIds": [],
  "images": [
    {
      "imageId": "...",
      "url": "http://localhost:9000/catalog-images/...",
      "altText": null,
      "sortOrder": 0,
      "isPrimary": true
    }
  ],
  "averageRating": null,
  "reviewCount": 0,
  "tags": [],
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-01T00:00:00Z",
  "version": 1
}
```

---

### Categories `/api/categories`

#### `POST /api/categories`

Create a category. Requires `Admin` policy.

**Request body:**

```json
{
  "name": "Electronics",
  "slug": "electronics",
  "description": "Optional",
  "parentId": null,
  "sortOrder": 0,
  "imageUrl": null
}
```

**Response `201 Created`:** `CategoryResponseDto`.

**Error codes:** `DUPLICATE_SLUG`.

---

#### `GET /api/categories`

List all active categories as a flat list. Public.

**Response `200 OK`:** `CategoryResponseDto[]`.

---

#### `GET /api/categories/tree`

Return active categories as a nested tree structure. Public. Cached for 15 minutes.

**Response `200 OK`:** `CategoryTreeDto[]`

```json
[
  {
    "id": "...",
    "name": "Electronics",
    "slug": "electronics",
    "description": null,
    "parentId": null,
    "sortOrder": 0,
    "isActive": true,
    "imageUrl": null,
    "children": [
      {
        "id": "...",
        "name": "Phones",
        "slug": "phones",
        "children": []
      }
    ]
  }
]
```

---

#### `GET /api/categories/{id}`

Get a single category by ID. Public.

**Response `200 OK`:** `CategoryResponseDto`.

**Response `404 Not Found`.**

---

#### `GET /api/categories/{id}/children`

Get immediate children of a category. Public.

**Response `200 OK`:** `CategoryResponseDto[]`.

---

#### `PUT /api/categories/{id}`

Update a category. Requires `Admin` policy.

**Request body:**

```json
{
  "name": "Updated Name",
  "description": null,
  "parentId": null,
  "sortOrder": 1,
  "isActive": true,
  "imageUrl": null
}
```

**Response `200 OK`:** `CategoryResponseDto`.

---

#### `DELETE /api/categories/{id}`

Delete a category. Requires `Admin` policy. Returns `409 Conflict` if the category has child categories or associated products.

**Response `204 No Content`.**

---

### Brands `/api/brands`

#### `POST /api/brands`

Create a brand. Requires `Admin` policy.

**Request body:**

```json
{
  "name": "Acme",
  "slug": "acme",
  "description": null,
  "logoUrl": null,
  "website": "https://acme.example.com"
}
```

**Response `201 Created`:** `BrandResponseDto`.

**Error codes:** `DUPLICATE_SLUG`.

---

#### `GET /api/brands`

List all active brands. Public.

**Response `200 OK`:** `BrandResponseDto[]`.

---

#### `GET /api/brands/{id}`

Get a brand by ID. Public.

**Response `200 OK`:** `BrandResponseDto`.

**Response `404 Not Found`.**

---

#### `PUT /api/brands/{id}`

Update a brand. Requires `Admin` policy.

**Request body:**

```json
{
  "name": "Updated Name",
  "description": null,
  "logoUrl": null,
  "website": null,
  "isActive": true
}
```

**Response `200 OK`:** `BrandResponseDto`.

---

#### `DELETE /api/brands/{id}`

Delete a brand. Requires `Admin` policy. Returns `409 Conflict` if products reference this brand.

**Response `204 No Content`.**

---

### Health `/health`

**`GET /health`** — Aggregates health checks for MongoDB, Redis, and RabbitMQ.

Standard ASP.NET Core health check response. Returns `200 OK` when all dependencies are healthy, `503 Service Unavailable` otherwise.

---

## Configuration

All settings are read from `appsettings.json` (or environment variable overrides). Sensitive values should be supplied via environment variables in production.

### `appsettings.json` reference

```json
{
  "Catalog": {
    "Mongo": {
      "ConnectionString": "mongodb://admin:admin@localhost:27020",
      "DatabaseName": "catalog"
    }
  },
  "Jwt": {
    "AccessSecret": ""
  },
  "Redis": {
    "ConnectionString": "localhost:6379"
  },
  "Minio": {
    "Endpoint": "localhost:9000",
    "AccessKey": "minioadmin",
    "SecretKey": "minioadmin",
    "BucketName": "catalog-images",
    "UseSSL": false,
    "PublicBaseUrl": "http://localhost:9000/catalog-images"
  },
  "RabbitMq": {
    "ConnectionString": "amqp://guest:guest@localhost:5672",
    "ExchangeName": "catalog.exchange",
    "QueueName": "catalog.events"
  }
}
```

### Environment variable overrides

The following environment variables are also checked by the application and take precedence over `appsettings.json` where noted:

| Variable | Overrides |
|---|---|
| `JWT_ACCESS_SECRET` | `Jwt:AccessSecret` |

All `appsettings.json` keys can also be supplied as environment variables using the `__` separator convention (e.g. `Catalog__Mongo__ConnectionString`).

### Key settings explained

| Setting | Purpose |
|---|---|
| `Catalog:Mongo:ConnectionString` | MongoDB connection string including credentials |
| `Catalog:Mongo:DatabaseName` | MongoDB database name (default: `catalog`) |
| `Jwt:AccessSecret` | Shared HS256 secret; must match the auth service's `JWT_ACCESS_SECRET` |
| `Redis:ConnectionString` | Redis host:port string |
| `Minio:Endpoint` | MinIO server hostname and port |
| `Minio:BucketName` | S3 bucket for product images |
| `Minio:PublicBaseUrl` | Base URL used to construct public image URLs |
| `Minio:UseSSL` | Whether to connect to MinIO over HTTPS |
| `RabbitMq:ConnectionString` | AMQP connection string |
| `RabbitMq:ExchangeName` | Durable direct exchange for catalog events |
| `RabbitMq:QueueName` | Durable queue bound to the exchange |

---

## Development Setup

### Prerequisites

- .NET 10 SDK
- Docker and Docker Compose

### Start MongoDB

The service ships with a `docker-compose.yml` that starts a local MongoDB instance on port **27020**.

```bash
cd services/catalog
docker compose up -d
```

MongoDB credentials: `admin` / `admin`. MongoDB indexes are created automatically on first startup by `MongoIndexesInitializer`.

### Run the API

```bash
cd services/catalog/Catalog.Api
dotnet run
```

Swagger UI is available at `http://localhost:<port>/swagger` in the `Development` environment.

### Run Tests

```bash
cd services/catalog
dotnet test
```

### Format Code

```bash
cd services/catalog
dotnet format Catalog.slnx
```

---

## Caching Strategy

Redis is used to cache hot read paths. All cache operations are fire-and-forget on failure — a Redis outage degrades to full database reads without breaking the API.

| Cache key pattern | Content | TTL | Invalidated on |
|---|---|---|---|
| `catalog:product:{id}` | `ProductResponseDto` by ID | 5 minutes | Product update, delete, status change, image upload/delete/reorder |
| `catalog:product:slug:{slug}` | `ProductResponseDto` by slug | 5 minutes | Product update, delete, status change, image upload/delete/reorder |
| `catalog:search:{hash}` | `SearchProductsResponse` for a given query fingerprint | 2 minutes | TTL expiry only (no active invalidation) |
| `catalog:categories:tree` | Full nested `CategoryTreeDto[]` | 15 minutes | TTL expiry only |

The search key is a 16-character hex prefix of the SHA-256 hash of the serialized query parameters (including resolved category descendants).

---

## Event Publishing

Events are published to RabbitMQ after successful write operations. The publisher is non-blocking: if RabbitMQ is unavailable, the failure is logged as a warning and the HTTP response is still returned. Events are not persisted for retry.

**Exchange:** `catalog.exchange` (durable, direct)
**Queue:** `catalog.events` (durable)
**Routing key:** `catalog.events`

All events are published as JSON with the envelope:

```json
{
  "eventType": "catalog.product.created",
  "occurredOn": "2026-01-01T00:00:00Z",
  "payload": { /* event-specific fields */ }
}
```

### Published events

| Event type | Trigger | Payload fields |
|---|---|---|
| `catalog.product.created` | `POST /api/products` | `productId`, `sku`, `title`, `brandId`, `categoryIds`, `price`, `currency` |
| `catalog.product.updated` | `PUT /api/products/{id}` (when fields actually change) | `productId`, `changedFields` (string list) |
| `catalog.product.deleted` | `DELETE /api/products/{id}` | `productId`, `sku` |
| `catalog.product.status_changed` | `PATCH /api/products/{id}/status` | `productId`, `oldStatus`, `newStatus` |
| `catalog.category.created` | `POST /api/categories` | `categoryId`, `name`, `parentId` |
| `catalog.brand.created` | `POST /api/brands` | `brandId`, `name` |

---

## Image Upload

Product images are stored in MinIO (an S3-compatible object store). The `POST /api/products/{id}/images` endpoint enforces the following validation before uploading:

| Rule | Detail |
|---|---|
| Max file size | 5 MB |
| Allowed MIME types | `image/jpeg`, `image/jpg`, `image/png`, `image/webp` |
| Magic byte validation | File header bytes are checked against the declared `Content-Type` to prevent content spoofing |
| Archived products | Uploading to an archived product is rejected |

After a successful upload, the image URL is constructed from `Minio:PublicBaseUrl` and stored on the product. The first uploaded image is automatically set as the primary image (`isPrimary: true`, `sortOrder: 0`). Subsequent images receive incrementing `sortOrder` values.

Deleting an image removes it from the product's `images` array. The MinIO object itself is also deleted via the stored `key`.

---

## MongoDB Indexes

Created automatically at startup by `MongoIndexesInitializer`:

**`products` collection:**

| Fields | Type | Notes |
|---|---|---|
| `sku` | Unique ascending | Enforces SKU uniqueness |
| `slug` | Unique ascending | Enforces slug uniqueness |
| `status`, `brandId` | Compound ascending | Supports status + brand filtering |
| `title`, `description` | Text (weighted) | Full-text search; title weight 10, description weight 1 |
| `categoryIds` | Ascending | Supports category-based filtering |
| `price.amount` | Ascending | Supports price range queries |

**`categories` collection:**

| Fields | Type | Notes |
|---|---|---|
| `slug` | Unique ascending | Enforces slug uniqueness |
| `parentId` | Ascending | Supports children lookup |
| `path` | Ascending | Supports ancestor/descendant traversal |

**`brands` collection:**

| Fields | Type | Notes |
|---|---|---|
| `slug` | Unique ascending | Enforces slug uniqueness |
| `name` | Ascending | Supports name-based queries |
