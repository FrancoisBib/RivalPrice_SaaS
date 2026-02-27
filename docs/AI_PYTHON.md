# AI Python - Documentation

## Vue d'ensemble

Le module AI Python analyse les données scrapées et détecte les changements via OpenAI.

## Architecture

```
PostgreSQL (snapshots)
        │
        ▼
┌─────────────────────────┐
│  Change Detection       │
│  Service                │
│  - Compare snapshots   │
│  - Détecte changements │
└─────────────────────────┘
        │
        ▼
PostgreSQL (detected_changes)
        │
        ▼
┌─────────────────────────┐
│  AI Analysis Service   │
│  - OpenAI analysis     │
│  - Summary + Impact    │
└─────────────────────────┘
        │
        ▼
PostgreSQL (ai_analysis)
```

## Services

### ChangeDetectionService (`services/change_detection_service.py`)

Détecte les changements entre 2 snapshots:

```python
service = IntelligentChangeDetectionService()
change = service.detect_changes(page_id=4)
```

**Types de changements détectés:**

| Type | Description |
|------|-------------|
| `price_increase` | Prix augmenté |
| `price_decrease` | Prix diminué |
| `feature_added` | Nouvelle fonctionnalité |
| `feature_removed` | Fonctionnalité supprimée |
| `messaging_change` | Changement de texte |
| `content_change` | Autre changement de contenu |

**Détection par hash:**
- Calcule un hash global du contenu
- Si hash différent → changement détecté

### AIAnalysisService (`services/ai_analysis_service.py`)

Génère une analyse via OpenAI:

```python
service = AIAnalysisService()
analysis = service.analyze_change(change_id=1)
```

**Retourne:**
- `summary`: Résumé en 1-2 phrases
- `recommendation`: Recommandation actionnable
- `impact_level`: Score 1-10
- `model`: Modèle OpenAI utilisé

## Modèles

### DetectedChange

```python
class DetectedChange:
    page_id: int
    page_type: str           # pricing, features
    old_price: str
    new_price: str
    change_percent: float
    old_availability: str
    new_availability: str
    old_features: JSON
    new_features: JSON
    features_added: JSON
    features_removed: JSON
    old_text: str
    new_text: str
    change_type: str         # price_increase, etc.
    detected_at: datetime
```

### AIAnalysis

```python
class AIAnalysis:
    change_id: int
    summary: str
    recommendation: str
    change_type: str
    page_type: str
    old_price: str
    new_price: str
    change_percent: str
    model: str
    created_at: datetime
```

## Dépendances

```
# requirements.txt
sqlalchemy>=2.0
psycopg2-binary
openai
python-dotenv
```

## Configuration

```bash
# .env
DATABASE_URL=postgresql://user:pass@host/db
OPENAI_API_KEY=sk-...
```

## Commandes

```bash
# Install dependencies
pip install -r ai-python/requirements.txt

# Run change detection
cd ai-python
python -c "from services.change_detection_service import IntelligentChangeDetectionService; s = IntelligentChangeDetectionService(); s.run_detection_for_all_pages()"

# Run AI analysis
python -c "from services.ai_analysis_service import AIAnalysisService; s = AIAnalysisService(); s.analyze_all_pending_changes()"
```
