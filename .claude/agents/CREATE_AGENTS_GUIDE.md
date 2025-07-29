# Guide de création des agents MMORPG

J'ai créé les fichiers de configuration pour 9 agents spécialisés. Voici comment les créer dans Claude Code :

## 📋 Instructions

Pour chaque agent, exécutez la commande `/agents` dans Claude Code et suivez ces étapes :

1. Choisissez **"project"** level (pour qu'ils soient spécifiques à ce projet)
2. Copiez les informations depuis le fichier correspondant dans `.claude/agents/`
3. Ajoutez les outils listés dans chaque fichier

## 🤖 Liste des agents à créer

### Agents Techniques

#### 1. go-backend-specialist
- **Fichier**: `.claude/agents/go-backend-specialist.md`
- **Rôle**: Développement microservices Go, architecture hexagonale
- **Outils**: Bash, Read, Edit, MultiEdit, Grep, Task, WebSearch

#### 2. ue5-cpp-specialist  
- **Fichier**: `.claude/agents/ue5-cpp-specialist.md`
- **Rôle**: Développement C++ Unreal Engine 5.6, intégration Blueprint
- **Outils**: Bash, Read, Edit, MultiEdit, Grep, Task

#### 3. api-proto-specialist
- **Fichier**: `.claude/agents/api-proto-specialist.md`
- **Rôle**: Conception API REST/WebSocket, Protocol Buffers
- **Outils**: Read, Edit, MultiEdit, Task, Grep

#### 4. devops-pipeline-specialist
- **Fichier**: `.claude/agents/devops-pipeline-specialist.md`
- **Rôle**: CI/CD, Docker, Kubernetes, monitoring
- **Outils**: Bash, Read, Edit, Task, WebSearch

#### 5. testing-qa-specialist
- **Fichier**: `.claude/agents/testing-qa-specialist.md`
- **Rôle**: Tests automatisés, QA, performance testing
- **Outils**: Bash, Read, Edit, Task, Grep

#### 6. database-architecture-specialist
- **Fichier**: `.claude/agents/database-architecture-specialist.md`
- **Rôle**: Architecture PostgreSQL/Redis, optimisation, sharding
- **Outils**: Bash, Read, Edit, Task, Grep

### Agents de Gestion

#### 7. project-manager
- **Fichier**: `.claude/agents/project-manager.md`
- **Rôle**: Gestion de projet, tracking d'avancement, coordination
- **Outils**: TodoWrite, Read, Edit, MultiEdit, Task, Grep

#### 8. technical-writer
- **Fichier**: `.claude/agents/technical-writer.md`
- **Rôle**: Rédaction documentation phases (requirements, design, tasks)
- **Outils**: Read, Write, Edit, MultiEdit, Task, WebSearch

#### 9. system-architect
- **Fichier**: `.claude/agents/system-architect.md`
- **Rôle**: Architecture globale, décisions techniques cross-système
- **Outils**: Read, Edit, Task, WebSearch, Grep

## 💡 Utilisation des agents

Une fois créés, vous pouvez invoquer un agent spécifique avec :

```
@go-backend-specialist Aide-moi à implémenter le service de personnages
```

ou utiliser la commande `/chat` pour démarrer une conversation avec un agent :

```
/chat go-backend-specialist
```

## 🎯 Cas d'usage recommandés

- **Planification de phase**: `project-manager` + `technical-writer` + `system-architect`
- **Phase 2 (Networking)**: `go-backend-specialist` + `ue5-cpp-specialist` + `api-proto-specialist`
- **Documentation**: `technical-writer` pour requirements/design/tasks de chaque phase
- **Architecture**: `system-architect` pour les décisions techniques majeures
- **Performance**: `database-architecture-specialist` + `testing-qa-specialist`
- **Déploiement**: `devops-pipeline-specialist` pour l'infrastructure
- **Tracking**: `project-manager` pour le suivi quotidien et les rapports

## 🔄 Workflow recommandé pour une nouvelle phase

1. **Planification** (avec `project-manager` + `technical-writer` + `system-architect`)
   - Définir les requirements avec `technical-writer`
   - Valider l'architecture avec `system-architect`
   - Créer le plan de tracking avec `project-manager`

2. **Implémentation** (agents techniques spécialisés)
   - Backend: `go-backend-specialist`
   - Frontend: `ue5-cpp-specialist`
   - API: `api-proto-specialist`
   - Database: `database-architecture-specialist`

3. **Validation** (avec `testing-qa-specialist` + `devops-pipeline-specialist`)
   - Tests et QA
   - Pipeline CI/CD

4. **Documentation finale** (avec `technical-writer` + `project-manager`)
   - Completion report
   - Mise à jour du tracking

## 📝 Notes

- Chaque agent a un prompt système détaillé avec son expertise
- Les agents connaissent l'architecture du projet et les standards de code
- Ils sont configurés pour travailler ensemble de manière cohérente
- Les prompts incluent les leçons apprises des phases précédentes
- Les 3 agents de gestion (7-9) sont essentiels pour la coordination globale