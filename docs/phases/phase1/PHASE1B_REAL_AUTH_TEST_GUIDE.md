# Guide de Test de l'Authentification Réelle

## 0. Configuration Base de Données (AUTOMATIQUE)

**Bonne nouvelle !** Vous n'avez **RIEN à créer dans pgAdmin4**. Tout est automatique :

- La base de données `mmorpg` est créée automatiquement par Docker
- Les tables `users` et `sessions` sont créées par les scripts de migration
- Les migrations SQL dans `/mmorpg-backend/migrations/` s'exécutent automatiquement au premier démarrage

### Ce qui se passe automatiquement :
1. Docker crée la base de données `mmorpg` avec user/password `dev`
2. Les scripts SQL créent :
   - Table `users` : stockage des utilisateurs avec mot de passe hashé
   - Table `sessions` : gestion des sessions actives et refresh tokens
   - Indexes pour les performances
   - Triggers pour les timestamps automatiques

Vous pouvez vérifier dans pgAdmin4 ou Adminer après le démarrage, mais **aucune action manuelle n'est requise !**

## 1. Démarrer les Services Backend

### Étape 1 : Démarrer les services d'infrastructure
```bash
cd mmorpg-backend
docker-compose up -d
```

Cela démarre :
- PostgreSQL (port 5432)
- Redis (port 6379)
- NATS (port 4222)

### Étape 2 : Vérifier que les services sont actifs
```bash
docker-compose ps
```

Vous devriez voir tous les services "healthy" ou "running".

### Étape 3 : Démarrer le Gateway
Dans un nouveau terminal :
```bash
cd mmorpg-backend
go run cmd/gateway/main.go
```
Le gateway démarre sur le port **8080**.

### Étape 4 : Démarrer le Service Auth
Dans un autre terminal :
```bash
cd mmorpg-backend
go run cmd/auth/main.go
```
Le service auth démarre sur le port **8081**.

### Étape 5 : Vérifier que la DB est prête (Optionnel)
Si c'est votre première fois, vérifiez que les tables sont créées :
```bash
# Option 1 - Via Docker
docker exec -it mmorpg-postgres psql -U dev -d mmorpg -c "\dt"

# Option 2 - Via Adminer
# Ouvrez http://localhost:8090 dans votre navigateur
```

Vous devriez voir les tables `users` et `sessions`.

## 2. Désactiver le Mode Mock dans Unreal

### Modification du Code C++

Ouvrez le fichier : `MMORPGTemplate/Source/MMORPGCore/Private/Subsystems/UMMORPGAuthSubsystem.cpp`

1. **Changer l'URL du serveur** (ligne 17) :
```cpp
// Remplacer :
ServerURL = TEXT("http://localhost:3000");
// Par :
ServerURL = TEXT("http://localhost:8080");
```

2. **Désactiver le mode mock** (ligne 36) :
```cpp
// Remplacer :
bool bUseMockMode = true;
// Par :
bool bUseMockMode = false;
```

3. **Corriger les URLs de l'API** (IMPORTANT - 4 endroits) :
```cpp
// Ligne 87 - Remplacer :
TEXT("/api/auth/login")
// Par :
TEXT("/api/v1/auth/login")

// Ligne 136 - Remplacer :
TEXT("/api/auth/register")
// Par :
TEXT("/api/v1/auth/register")

// Ligne 165 - Remplacer :
TEXT("/api/auth/refresh")
// Par :
TEXT("/api/v1/auth/refresh")

// Ligne 210 - Remplacer :
TEXT("/api/auth/me")
// Par :
TEXT("/api/v1/auth/me")
```

4. **Recompiler le projet** dans Unreal Engine.

## 3. Tester l'Authentification

### Test 1 : Créer un Compte

1. Lancez le jeu dans l'éditeur Unreal
2. Sur l'écran de login, cliquez sur "Register"
3. Entrez :
   - Email : `test@example.com`
   - Username : `testuser`
   - Password : `password123`
   - Confirm Password : `password123`
4. Cliquez sur "Register"

**Succès attendu** : Message de succès et redirection vers l'écran de login.

### Test 2 : Se Connecter

1. Sur l'écran de login, entrez :
   - Email : `test@example.com`
   - Password : `password123`
2. Cliquez sur "Login"

**Succès attendu** : 
- Connexion réussie
- Tokens JWT stockés localement
- Transition vers le jeu principal

### Test 3 : Vérifier dans la Base de Données

Ouvrez Adminer dans votre navigateur : http://localhost:8090

- **Serveur** : `postgres`
- **Utilisateur** : `dev`
- **Mot de passe** : `dev`
- **Base de données** : `mmorpg`

Vérifiez les tables :
- `users` : Votre utilisateur créé
- `sessions` : Session active avec refresh token

## 4. Tests API Directs (Optionnel)

### Créer un compte via curl :
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "api@test.com",
    "username": "apiuser",
    "password": "testpass123"
  }'
```

### Se connecter via curl :
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "api@test.com",
    "password": "testpass123"
  }'
```

## 5. Dépannage

### Erreur "Service Unavailable"
- Vérifiez que le gateway ET le service auth sont lancés
- Vérifiez les logs dans les terminaux

### Erreur de connexion à la base de données
- Vérifiez que PostgreSQL est actif : `docker-compose ps`
- Vérifiez les logs : `docker-compose logs postgres`

### Erreur "Network Error" dans Unreal
- Vérifiez que l'URL est correcte : `http://localhost:8080`
- Vérifiez que le mode mock est désactivé
- Recompilez après les changements

### Voir les logs du backend
- Gateway : Regardez le terminal où vous avez lancé `go run cmd/gateway/main.go`
- Auth Service : Regardez le terminal où vous avez lancé `go run cmd/auth/main.go`

## 6. Arrêter les Services

Quand vous avez fini :
```bash
# Arrêter les services Go : Ctrl+C dans chaque terminal

# Arrêter Docker :
cd mmorpg-backend
docker-compose down
```

## Notes Importantes

- Les mots de passe sont hachés avec bcrypt dans la base de données
- Les tokens JWT expirent après 15 minutes (access) et 7 jours (refresh)
- Le système gère automatiquement le refresh des tokens
- Les sessions sont stockées dans Redis pour de meilleures performances