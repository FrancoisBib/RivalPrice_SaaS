# RivalPrice 🦁

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat&logo=python)
![Next.js](https://img.shields.io/badge/Next.js-14+-000000?style=flat&logo=next.js)

> **Français** | English

**RivalPrice** est une plateforme de veille stratégique dédiée aux fondateurs de SaaS. L'outil automatise la surveillance de vos concurrents : de l'évolution des tarifs au déploiement de nouvelles fonctionnalités. Avec RivalPrice, vous ne vous laissez plus surprendre par le marché ; vous gardez toujours une longueur d'avance.

---

## English Description

**RivalPrice** is a strategic intelligence platform designed for SaaS founders. The tool automates competitor monitoring—from pricing evolution to new feature deployments. With RivalPrice, you'll never be surprised by the market again; you'll always stay one step ahead.

Thanks to AI, we don't just monitor your rivals: we analyze their pricing changes, decode their new features, and summarize their strategy for you. Stop receiving alerts; start receiving action plans.

---

## 🚀 Features

- **🔍 Automated Scraping** - Schedule and monitor competitor websites
- **💰 Price Tracking** - Monitor pricing changes in real-time
- **✨ Feature Monitoring** - Detect new features and changes
- **🤖 AI-Powered Analysis** - LLM-powered strategic insights
- **📊 Action Plans** - Get actionable recommendations
- **🔔 Smart Alerts** - Customizable notifications
- **📈 Dashboard** - Beautiful Next.js interface

---

## 🏗️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│  API (Go)   │────▶│  PostgreSQL │
│  (Next.js)  │     │    Gin     │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                   │
                           ▼                   ▼
                    ┌─────────────┐     ┌─────────────┐
                    │    Redis    │◀────│   Scrapers  │
                    │   (Queue)   │     │   (Workers) │
                    └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ AI Engine   │
                    │ (Python)    │
                    └─────────────┘
```

### Tech Stack

| Component | Technology |
|-----------|------------|
| **Backend API** | Go + Gin |
| **Database** | PostgreSQL |
| **Cache/Queue** | Redis |
| **Scrapers** | Go Workers |
| **AI Engine** | Python + LLM |
| **Frontend** | Next.js 14+ |
| **Container** | Docker |

---

## 📦 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Python 3.11+ (for AI engine)
- PostgreSQL 15+
- Redis 7+

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/FrancoisBib/RivalPrice_SaaS.git
cd RivalPrice_SaaS
```

2. **Start with Docker**
```bash
docker-compose up -d
```

3. **Environment Variables**
Create a `.env` file:
```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/rivalprice?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# AI Engine
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...

# JWT
JWT_SECRET=your-secret-key
```

---

## 📁 Project Structure

```
RivalPrice_SaaS/
│
├── api-go/                  # Backend API (Go + Gin)
│   ├── cmd/
│   │   └── main.go         # Entry point
│   ├── config/             # Configuration
│   ├── models/             # GORM models
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── competitor.go
│   │   ├── monitored_page.go
│   │   └── snapshot.go
│   ├── routes/             # HTTP routes
│   ├── controllers/        # HTTP handlers
│   ├── services/           # Business logic
│   └── utils/              # Helpers
│
├── scraper-go/             # Scraping workers (Go)
│   ├── cmd/
│   ├── workers/
│   └── utils/
│
├── ai-python/              # AI Strategic Engine (Python)
│   ├── services/
│   │   ├── change_detector.py
│   │   └── ai_analyzer.py
│   └── requirements.txt
│
├── frontend/               # Dashboard (Next.js)
│   ├── app/
│   ├── components/
│   └── pages/
│
└── docker-compose.yml      # Full stack orchestration
```

---

## 🔌 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Get current user

### Projects
- `GET /api/projects` - List projects
- `POST /api/projects` - Create project
- `GET /api/projects/:id` - Get project
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project

### Competitors
- `GET /api/projects/:id/competitors` - List competitors
- `POST /api/projects/:id/competitors` - Add competitor
- `GET /api/competitors/:id` - Get competitor
- `DELETE /api/competitors/:id` - Delete competitor

### Monitored Pages
- `GET /api/competitors/:id/pages` - List monitored pages
- `POST /api/competitors/:id/pages` - Add page to monitor
- `DELETE /api/pages/:id` - Stop monitoring

### Snapshots & Analysis
- `GET /api/pages/:id/snapshots` - List snapshots
- `GET /api/pages/:id/analysis` - Get AI analysis

---

## 🗄️ Database Schema

### Tables

- **users** - User accounts
- **projects** - Customer projects
- **competitors** - Competitor entries
- **monitored_pages** - Pages to track
- **snapshots** - Page snapshots (JSONB)
- **detected_changes** - Change detection
- **ai_analysis** - AI-powered insights
- **alerts** - User notifications

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🦁 Author

**François Bib** 🦁
- GitHub: [@FrancoisBib](https://github.com/FrancoisBib)

---

*Ne recevez plus des alertes, recevez des plans d'action.*
