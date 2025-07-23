# Phase 0 - Guide de Test Complet

## 🧪 Objectif
Tester tous les composants implémentés dans la Phase 0 pour vérifier que tout fonctionne correctement.

## 📋 Prérequis
- Docker Desktop lancé
- Unreal Engine 5.6 installé
- Visual Studio 2022 installé
- Terminal/PowerShell ouvert

## 🚀 Tests à Effectuer

### 1. Test de l'Infrastructure Backend

#### 1.1 Démarrer les services Docker
```bash
# Se positionner dans le dossier backend
cd mmorpg-backend

# Démarrer tous les services
docker-compose up -d

# Vérifier que tout est lancé
docker-compose ps
```

**✅ Résultat attendu :**
- PostgreSQL sur le port 5432
- Redis sur le port 6379
- NATS sur le port 4222
- Prometheus sur le port 9090
- Grafana sur le port 3000

#### 1.2 Compiler et lancer le Gateway
```bash
# Compiler
go build -o bin/gateway.exe cmd/gateway/main.go

# OU utiliser Make
make build

# Lancer le gateway
./bin/gateway.exe
```

**✅ Résultat attendu :**
- Message "Gateway server starting on :8090"
- Pas d'erreurs de connexion aux services

#### 1.3 Tester les endpoints HTTP
```bash
# Test de base
curl http://localhost:8090/

# Test de l'API
curl http://localhost:8090/api/v1/test

# Test echo
curl -X POST http://localhost:8090/api/v1/echo -H "Content-Type: application/json" -d '{"message":"Hello MMORPG"}'

# Health check
curl http://localhost:8090/health
```

**✅ Résultat attendu :**
- Réponses JSON valides
- Status 200 OK

### 2. Test du Plugin Unreal Engine

#### 2.1 Ouvrir le projet UE5
1. Lancer Unreal Engine 5.6
2. Créer un nouveau projet vide
3. Copier le dossier `UnrealEngine/Plugins/MMORPGTemplate` dans `Plugins/` du projet
4. Relancer l'éditeur pour charger le plugin

**✅ Résultat attendu :**
- Plugin visible dans Edit → Plugins → Project → MMORPG
- Pas d'erreurs de compilation

#### 2.2 Tester la connexion réseau
1. Créer un nouveau niveau
2. Ouvrir le Level Blueprint
3. Ajouter les nodes suivants :

```blueprint
BeginPlay → Get MMORPG Core Module → Get Network Manager → Connect to Server
├─ Host: localhost
└─ Port: 8090
```

4. Jouer en éditeur (PIE)
5. Vérifier la console de sortie (Window → Developer Tools → Output Log)

**✅ Résultat attendu :**
- Message "Connecting to localhost:8090..."
- Pas d'erreurs de connexion

### 3. Test de la Console Développeur

#### 3.1 Créer le Widget Console
1. Content Browser → Clic droit → User Interface → Widget Blueprint
2. Nommer : `WBP_DeveloperConsole`
3. Parent Class : `MMORPGConsoleWidget`
4. Ouvrir et designer :
   - Ajouter un `ScrollBox` pour l'output
   - Ajouter un `EditableTextBox` pour l'input
   - Configurer les événements

#### 3.2 Tester la console en jeu
1. Jouer en éditeur
2. Appuyer sur **F1** (ou taper `mmorpg.console` dans la console UE)
3. Tester les commandes :

```
help                    # Liste des commandes
status                  # État du système
connect localhost 8090  # Connexion au serveur
test                    # Test de connexion
clear                   # Effacer la console
fps                     # Afficher FPS
memstats                # Stats mémoire
```

**✅ Résultat attendu :**
- Console qui s'ouvre/ferme avec F1
- Commandes qui fonctionnent
- Historique avec flèches haut/bas

### 4. Test du Système d'Erreurs

#### 4.1 Tester les erreurs réseau
1. Arrêter le serveur backend
2. Dans UE5, essayer de se connecter
3. Observer la gestion d'erreur

**✅ Résultat attendu :**
- Message d'erreur clair
- Pas de crash
- Option de retry si applicable

#### 4.2 Tester depuis Blueprint
```blueprint
Report Error
├─ Error Code: 1001
├─ Message: "Test error from Blueprint"
└─ Severity: Error
```

**✅ Résultat attendu :**
- Erreur loggée dans Output Log
- Notification si configurée

### 5. Test des Protocol Buffers

#### 5.1 Compiler les protos
```bash
cd mmorpg-backend
make proto

# OU sur Windows
scripts/compile_proto_local.bat
```

**✅ Résultat attendu :**
- Fichiers .pb.go générés dans `pkg/proto/`
- Fichiers .pb.h/.pb.cc générés dans UE plugin

#### 5.2 Tester la sérialisation (Blueprint)
1. Créer un Blueprint Actor
2. Ajouter ce graph :

```blueprint
Event BeginPlay → Create Login Request
├─ Username: "testuser"
├─ Password: "testpass"
└─ → Serialize to JSON → Print String
```

**✅ Résultat attendu :**
- JSON valide affiché
- Pas d'erreurs de sérialisation

### 6. Test du CI/CD (GitHub Actions)

#### 6.1 Vérifier les workflows
1. Aller sur https://github.com/cafe1231/MMORPG_GameTemplate/actions
2. Vérifier que les workflows sont présents :
   - Go Backend CI
   - Unreal Plugin CI
   - Protocol Buffers CI
   - Main CI/CD

**✅ Résultat attendu :**
- Workflows visibles
- Prêts à s'exécuter sur push

### 7. Test de Monitoring

#### 7.1 Accéder à Grafana
1. Ouvrir http://localhost:3000
2. Login : admin/admin
3. Importer un dashboard

**✅ Résultat attendu :**
- Interface Grafana accessible
- Métriques Prometheus disponibles

### 8. Tests de Performance

#### 8.1 Test de charge basique
```bash
# Installer hey (outil de test HTTP)
go install github.com/rakyll/hey@latest

# Test avec 100 requêtes
hey -n 100 -c 10 http://localhost:8090/api/v1/test
```

**✅ Résultat attendu :**
- Temps de réponse < 50ms
- Pas d'erreurs

## 📝 Checklist de Validation

### Backend Go
- [ ] Docker Compose démarre tous les services
- [ ] Gateway compile et démarre
- [ ] Endpoints HTTP répondent
- [ ] Pas d'erreurs dans les logs

### Plugin UE5
- [ ] Plugin se charge sans erreur
- [ ] Network Manager se connecte
- [ ] Blueprint nodes disponibles
- [ ] Pas de warnings de compilation

### Console Développeur
- [ ] Console s'ouvre avec F1
- [ ] Commandes de base fonctionnent
- [ ] Historique fonctionne
- [ ] Auto-complétion fonctionne

### Gestion d'Erreurs
- [ ] Erreurs réseau gérées proprement
- [ ] Messages d'erreur clairs
- [ ] Pas de crash sur erreur
- [ ] Retry logic fonctionne

### Protocol Buffers
- [ ] Compilation réussie
- [ ] Sérialisation Go ↔ UE5 fonctionne
- [ ] Types convertis correctement

### Infrastructure
- [ ] Hot reload fonctionne (Go)
- [ ] Logs structurés
- [ ] Métriques collectées
- [ ] Configuration chargée

## 🔧 Troubleshooting

### Le serveur ne démarre pas
```bash
# Vérifier les ports
netstat -an | findstr "8090"

# Voir les logs Docker
docker-compose logs -f

# Reconstruire si nécessaire
docker-compose down -v
docker-compose up --build
```

### Plugin UE5 ne compile pas
1. Vérifier la version d'UE (5.6+)
2. Régénérer les fichiers projet
3. Nettoyer et recompiler

### Console ne s'ouvre pas
1. Vérifier que le widget est créé
2. Vérifier le mapping de touche F1
3. Utiliser la commande `mmorpg.console`

## 🎉 Tests Réussis ?

Si tous les tests passent, la Phase 0 est validée ! Tu peux :
1. Faire un commit de célébration
2. Passer à la Phase 1
3. Partager ton succès

## 📊 Rapport de Test

Après les tests, documente les résultats :
- Version testée : 
- Date : 
- Environnement : Windows/Mac/Linux
- Tests réussis : X/Y
- Issues trouvées : 
- Performance : 

Bonne chance avec les tests ! 🚀