# Scraper Go - Documentation

## Vue d'ensemble

Le Scraper est un worker Go qui récupère les pages web et stocke les snapshots en base de données.

## Fonctionnement

```
Redis Queue (scrape_job)
        │
        ▼
┌─────────────────┐
│  Worker Go      │
│  - Fetch HTTP  │
│  - Extract     │
│  - Save DB     │
└─────────────────┘
        │
        ▼
   PostgreSQL
   (snapshots)
```

## Job Redis

Le scraper écoute la queue Redis `scrape_job`. Format du job:

```json
{
  "page_id": 1,
  "url": "https://competitor.com/pricing",
  "type": "pricing"
}
```

## Extraction de données

### Prix (`extractPrice`)

Regex utilisés pour détecter les prix:
- `$99.99`, `USD 99.99`
- `€99.99`, `EUR 99.99`
- `£99.99`, `GBP 99.99`
- `data-price="99.99"`
- `class="price">99.99`

### Disponibilité (`extractAvailability`)

Détection par mot-clés:
- `out of stock` / `outofstock` → `out_of_stock`
- `in stock` / `instock` / `available` → `in_stock`
- `pre-order` / `preorder` → `pre_order`

### Titre

Extraction via regex `<title>([^<]+)</title>`

## Configuration

Variables d'environnement:

| Variable | Défaut | Description |
|----------|--------|-------------|
| `DATABASE_URL` | localhost:5432 | PostgreSQL |
| `REDIS_ADDR` | localhost:6379 | Redis |
| `REDIS_PASSWORD` | - | Mot de passe Redis |
| `REDIS_DB` | 0 | Numéro de base Redis |

## Modèle Snapshot

```go
type Snapshot struct {
    ID              uint
    MonitoredPageID uint
    Price           string    // Prix extrait
    Availability    string    // in_stock, out_of_stock, pre_order
    RawData         JSON      // {title, url, html, price_found, availability, status_code}
    ScrapedAt       time.Time
}
```

## Dépendances

- `github.com/go-redis/redis/v8` - Client Redis
- `gorm.io/gorm` - ORM PostgreSQL
- `gorm.io/driver/postgres` - Driver PostgreSQL

## Commandes

```bash
# Build
cd scraper-go && go build -o scraper ./cmd/main.go

# Run
./scraper

# Docker
docker build -t rivalprice-scraper ./scraper-go
```
