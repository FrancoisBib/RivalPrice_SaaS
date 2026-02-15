RivalPrice, une plateforme de veille stratégique dédiée aux fondateurs de SaaS. L'outil automatise la surveillance de vos concurrents : de l'évolution des tarifs au déploiement de nouvelles fonctionnalités. Avec RivalPrice, vous ne vous laissez plus surprendre par le marché ; vous gardez toujours une longueur d'avance. Grâce à l'IA, nous ne nous contentons pas de surveiller vos rivaux : nous analysons leurs changements de prix, décryptons leurs nouvelles fonctionnalités et résumons leur stratégie pour vous. Ne recevez plus des alertes, recevez des plans d'action.

🎯 Architecture ajustée
🔹 API Backend → Go

✅ API REST en Go

Frameworks recommandés :

Gin (mature, simple)

Fiber (très rapide, DX moderne)

Echo (solide et structuré)

🔹 Python pour :

AI Strategic Engine

Analyse LLM

Génération des plans d’action

🏗 Nouvelle structure

Frontend (Next.js)
        ↓
API Backend (Go)
        ↓
PostgreSQL
        ↓
Redis (queue)
        ↓
Scraper Workers (Go)
        ↓
AI Engine (Python)

✅ Avantages de remplacer FastAPI

API plus performante

Consommation RAM plus faible

Une seule stack backend (Go) pour :

API

Scraper

Workers

Moins de context switching

⚠️ Points à anticiper

Validation des requêtes (struct tags + validator)

Gestion ORM (GORM ou sqlc recommandé)

Migration DB (golang-migrate)

Structuration propre dès le départ (hexagonal ou clean architecture)



Structure de dossier

/RivalPrice_SaaS
│
├── api-go/                  # Backend principal en Go (Gin)
│   ├── cmd/                 # Point d’entrée
│   │   └── main.go
│   ├── config/              # Config app (DB, Redis)
│   │   └── config.go
│   ├── models/              # Modèles GORM
│   │   └── user.go
│   │   └── project.go
│   │   └── competitor.go
│   │   └── monitored_page.go
│   │   └── snapshot.go
│   ├── routes/              # Routes HTTP
│   │   └── routes.go
│   ├── controllers/         # Handlers HTTP
│   │   └── user_controller.go
│   │   └── project_controller.go
│   ├── services/            # Logique métier
│   │   └── task_service.go
│   ├── utils/               # Helpers, validation
│   │   └── logger.go
│   └── go.mod
│
├── scraper-go/              # Workers Go pour scraping
│   ├── cmd/
│   │   └── main.go
│   ├── workers/
│   │   └── scraper_worker.go
│   ├── utils/
│   │   └── http_client.go
│   └── go.mod
│
├── ai-python/               # AI Engine en Python
│   ├── main.py
│   ├── services/
│   │   └── change_detector.py
│   │   └── ai_analyzer.py
│   ├── requirements.txt
│
├── frontend/                # Next.js Dashboard
│   ├── app/
│   ├── components/
│   ├── pages/
│   └── package.json
│
├── docker-compose.yml       # PostgreSQL + Redis + API Go + AI Python
└── README.md


Schéma base de données optimisé pour Go / RivalPrice

-- USERS
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- PROJECTS (projets surveillés par les clients)
CREATE TABLE projects (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- COMPETITORS
CREATE TABLE competitors (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT REFERENCES projects(id),
    name VARCHAR(255) NOT NULL,
    url VARCHAR(512),
    created_at TIMESTAMP DEFAULT NOW()
);

-- MONITORED PAGES
CREATE TABLE monitored_pages (
    id BIGSERIAL PRIMARY KEY,
    competitor_id BIGINT REFERENCES competitors(id),
    page_type VARCHAR(50), -- pricing / features
    url VARCHAR(512) NOT NULL,
    css_selector TEXT, -- optionnel pour ciblage spécifique
    created_at TIMESTAMP DEFAULT NOW()
);

-- SNAPSHOTS
CREATE TABLE snapshots (
    id BIGSERIAL PRIMARY KEY,
    monitored_page_id BIGINT REFERENCES monitored_pages(id),
    snapshot JSONB NOT NULL, -- version complète de la page
    hash CHAR(64) NOT NULL,  -- hash pour comparaison rapide
    created_at TIMESTAMP DEFAULT NOW()
);

-- DETECTED CHANGES
CREATE TABLE detected_changes (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT REFERENCES snapshots(id),
    change_type VARCHAR(50), -- price / feature / content
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- AI ANALYSIS
CREATE TABLE ai_analysis (
    id BIGSERIAL PRIMARY KEY,
    detected_change_id BIGINT REFERENCES detected_changes(id),
    summary TEXT,
    impact_score INT,
    plan_of_action JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ALERTS
CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT REFERENCES projects(id),
    message TEXT,
    alert_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
