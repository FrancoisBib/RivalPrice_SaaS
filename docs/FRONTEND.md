# Frontend - Documentation

## Vue d'ensemble

Application React/Next.js pour l'interface utilisateur de RivalPrice.

## Stack technique

| Technologie | Version | Usage |
|-------------|---------|-------|
| Next.js | 14.x | Framework React SSR |
| React | 18.x | UI Library |
| Tailwind CSS | 3.x | Styling |
| Zustand | 4.x | State management |
| React Query | 5.x | Data fetching |
| Axios | 1.x | HTTP client |
| Lucide React | - | Icônes |
| date-fns | - | Formatage dates |

## Structure

```
frontend/
├── app/                 # Next.js App Router
│   ├── page.tsx        # Home page
│   └── layout.tsx      # Root layout
├── components/         # Composants React
│   └── ...
├── pages/              # Pages (legacy)
│   └── ...
├── package.json
└── Dockerfile
```

## State Management (Zustand)

```typescript
// Exemple de store
import { create } from 'zustand'

interface UserStore {
  user: User | null
  setUser: (user: User) => void
}

export const useUserStore = create<UserStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
}))
```

## Data Fetching (React Query)

```typescript
// Exemple de hook
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'

const useCompetitors = () => {
  return useQuery({
    queryKey: ['competitors'],
    queryFn: () => axios.get('/api/v1/competitors'),
  })
}
```

## API Client (Axios)

```typescript
// Configuration axios
import axios from 'axios'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
})

// Interceptor pour JWT
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

## Composants UI

### Icônes (Lucide React)

```typescript
import { AlertTriangle, TrendingUp, DollarSign } from 'lucide-react'

<AlertTriangle className="w-5 h-5 text-red-500" />
```

### Dates (date-fns)

```typescript
import { format } from 'date-fns'
import { fr } from 'date-fns/locale'

format(new Date(), 'dd MMM yyyy', { locale: fr })
// → "27 févr. 2026"
```

## Commandes

```bash
# Install
cd frontend && npm install

# Dev
npm run dev

# Build
npm run build

# Docker
docker build -t rivalprice-frontend ./frontend
```

## Environment

```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Pages principales

- `/` - Dashboard
- `/login` - Connexion
- `/register` - Inscription
- `/projects` - Liste des projets
- `/competitors` - Liste des concurrents
- `/monitored-pages` - Pages surveillées
