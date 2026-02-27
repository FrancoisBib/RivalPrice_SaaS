# RivalPrice SaaS - Architecture

## Vue d'ensemble

RivalPrice est une application SaaS de veille concurrentielle qui surveille les prix et contenus des sites concurrents.

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND (React)                         │
│                   http://localhost:3000                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     API GOLANG (REST API)                       │
│                   http://localhost:8080                         │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────────┐  │
│  │Users    │ │Projects  │ │Competitors│ │Monitored Pages  │  │
│  └─────────┘ └──────────┘ └──────────┘ └─────────────────┘  │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────────┐  │
│  │Alert    │ │Scraper   │ │Scheduler │ │AI Client        │  │
│  │Service  │ │Service   │ │Service   │ │(OpenAI)         │  │
│  └─────────┘ └──────────┘ └──────────┘ └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SCRAPER GOLANG                               │
│                   (Collecte les données)                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  AI PYTHON (Analyse IA)                         │
│            (Détection changements + OpenAI)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     POSTGRESQL (BDD)                            │
│        Users │ Projects │ Competitors │ Pages │ Snapshots     │
│        DetectedChanges │ AlertLogs │ Notifications             │
└─────────────────────────────────────────────────────────────────┘
```

## Composants

| Composant | Technologie | Port | Rôle |
|-----------|-------------|------|------|
| Frontend | React + TypeScript | 3000 | Interface utilisateur |
| API | Go + Gin | 8080 | REST API, logique métier |
| Scraper | Go + Colly | 9090 | Collecte des pages web |
| AI Engine | Python + SQLAlchemy | - | Détection changements, OpenAI |
| Database | PostgreSQL | 5432 | Stockage données |
| Cache | Redis | 6379 | Sessions, rate limiting |

## Flux de données

1. **Ajout competitor** → Frontend → API → BDD
2. **Scrape planned** → Scheduler → Scraper → Snapshots BDD
3. **Détection changements** → AI Python compare snapshots → DetectedChanges
4. **Génération alertes** → API analise → AlertLogs + Emails

## Environment

Voir `.env.example` pour les variables nécessaires:
- `DATABASE_URL` - PostgreSQL
- `REDIS_URL` - Redis  
- `OPENAI_API_KEY` - Pour analyses IA
- `JWT_SECRET` - Authentification
