# API Go - Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication

L'API utilise JWT. Include le token dans le header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Auth

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| POST | `/auth/register` | Inscription utilisateur | Non |
| POST | `/auth/login` | Connexion | Non |
| GET | `/auth/me` | Profil utilisateur | Oui |

### Users

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/users` | Liste utilisateurs | Oui |
| POST | `/users` | Créer utilisateur | Oui |
| GET | `/users/:id` | Détails utilisateur | Oui |

### Projects

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/projects` | Liste projets | Oui |
| POST | `/projects` | Créer projet | Oui |
| GET | `/projects/:id` | Détails projet | Oui |

### Competitors

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/competitors` | Liste concurrents | Oui |
| POST | `/competitors` | Ajouter concurrent | Oui |
| GET | `/competitors/:id` | Détails concurrent | Oui |

### Monitored Pages

| Méthode | Endpoint | Description | Auth |
|---------|----------|-------------|------|
| GET | `/monitored_pages` | Liste pages surveillées | Oui |
| POST | `/monitored_pages` | Ajouter page | Oui |
| GET | `/monitored_pages/:id` | Détails page | Oui |

## Modèles

### User
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### Project
```json
{
  "id": 1,
  "user_id": 1,
  "name": "My Project",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### Competitor
```json
{
  "id": 1,
  "project_id": 1,
  "name": "Competitor Corp",
  "website": "https://competitor.com",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### MonitoredPage
```json
{
  "id": 1,
  "competitor_id": 1,
  "url": "https://competitor.com/pricing",
  "page_type": "pricing",
  "scrape_interval": 3600,
  "created_at": "2026-01-01T00:00:00Z"
}
```

## Services

### AlertService (`services/alert_service.go`)
- `CreateAlert()` - Crée une alerte
- `GetUnnotifiedAlerts()` - Récupère alertes non envoyées
- `MarkAsNotified()` - Marque alerte comme envoyée

### AIClient (`services/ai_client.go`)
- `Analyze()` - Génère résumé et recommandation via OpenAI

### EmailService (`services/email_service.go`)
- `SendAlertEmail()` - Envoie email d'alerte

### SchedulerService (`services/scheduler_service.go`)
- `StartScheduler()` - Démarre le planificateur de scrapes

## Middleware

### JWTAuthMiddleware (`middleware/auth.go`)
Vérifie la validité du token JWT.

### RateLimitMiddleware (`middleware/ratelimit.go`)
Limite les requêtes (100/min par défaut).
