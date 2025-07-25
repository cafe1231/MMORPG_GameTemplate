# Plan de Réimplémentation - Phase 1B : Frontend Authentication System

## État Actuel
- **Phase 1A Complétée** : Backend authentication fonctionnel (JWT, login, register, refresh, logout)
- **Problèmes Phase 1B** : Erreurs de compilation avec les types `void*` et les templates dans les UFUNCTION

## Objectifs Phase 1B
1. Système d'authentification frontend dans UE5.6
2. Interface utilisateur de connexion/inscription
3. Gestion des tokens JWT côté client
4. Système de persistance des sessions
5. Préparation pour le système de personnages

## Plan d'Implémentation Structuré

### Étape 1 : Création du Projet UE5.6 Propre
```
1. Créer nouveau projet UE5.6 "MMORPGTemplate"
2. Type: C++ Game
3. Template: Blank
4. Settings:
   - Target Platform: Desktop
   - Quality Preset: Maximum
   - Raytracing: Disabled
   - Starter Content: No
```

### Étape 2 : Structure Modulaire (Éviter les Erreurs)

**Modules à créer:**
1. **MMORPGCore** - Types de données et logique métier
2. **MMORPGNetwork** - Communication HTTP/WebSocket
3. **MMORPGUI** - Widgets et interfaces

**Structure recommandée:**
```
MMORPGTemplate/
├── Source/
│   ├── MMORPGTemplate/         # Module principal
│   ├── MMORPGCore/            # Types et subsystems
│   │   ├── Public/
│   │   │   ├── Types/         # Structures de données
│   │   │   └── Subsystems/    # Game Instance Subsystems
│   │   └── Private/
│   ├── MMORPGNetwork/         # Réseau
│   │   ├── Public/
│   │   │   ├── Http/          # Client HTTP
│   │   │   └── WebSocket/     # Client WebSocket
│   │   └── Private/
│   └── MMORPGUI/             # Interface utilisateur
│       ├── Public/
│       │   └── Auth/          # Widgets d'authentification
│       └── Private/
```

### Étape 3 : Types de Données (Sans void*)

**FAuthTypes.h** - Structures Blueprint-friendly:
```cpp
USTRUCT(BlueprintType)
struct FLoginRequest
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadWrite)
    FString Email;
    
    UPROPERTY(BlueprintReadWrite)
    FString Password;
};

USTRUCT(BlueprintType)
struct FAuthTokens
{
    GENERATED_BODY()
    
    UPROPERTY(BlueprintReadOnly)
    FString AccessToken;
    
    UPROPERTY(BlueprintReadOnly)
    FString RefreshToken;
    
    UPROPERTY(BlueprintReadOnly)
    FDateTime ExpiresAt;
};
```

### Étape 4 : HTTP Client (Sans Templates Complexes)

**Design Pattern pour éviter les erreurs:**
- Pas de `void*` dans les UFUNCTION
- Pas de templates complexes dans les paramètres par défaut
- Utiliser des surcharges au lieu de paramètres par défaut

```cpp
// Au lieu de:
UFUNCTION()
void Connect(const FString& URL, const TMap<FString, FString>& Headers = TMap<FString, FString>());

// Faire:
UFUNCTION()
void Connect(const FString& URL);

UFUNCTION()
void ConnectWithHeaders(const FString& URL, const TMap<FString, FString>& Headers);
```

### Étape 5 : Ordre d'Implémentation

**Semaine 1 - Infrastructure de Base:**
1. Créer les modules
2. Implémenter les types de données
3. Créer le HTTP Client basique
4. Tester la compilation

**Semaine 2 - Système d'Authentification:**
1. Créer UMMORPGAuthSubsystem
2. Implémenter login/register/refresh/logout
3. Ajouter la gestion des erreurs
4. Créer le système de sauvegarde

**Semaine 3 - Interface Utilisateur:**
1. Créer WBP_LoginScreen
2. Créer WBP_RegisterScreen
3. Implémenter la navigation
4. Ajouter les validations

**Semaine 4 - Tests et Polish:**
1. Tests d'intégration complets
2. Gestion des erreurs réseau
3. Animations et feedback visuel
4. Documentation

## Bonnes Pratiques pour Éviter les Problèmes

1. **Compilation Fréquente**
   - Compiler après chaque nouveau fichier .h/.cpp
   - Vérifier les erreurs UHT immédiatement

2. **Types Blueprint-Safe**
   - Utiliser uniquement des types supportés par Blueprint
   - Éviter void*, templates complexes, std::string

3. **Séparation C++/Blueprint**
   - Logique complexe en C++ pur (sans UFUNCTION)
   - Wrappers simples pour Blueprint

4. **Tests Unitaires**
   - Créer des tests pour chaque composant
   - Vérifier la sérialisation JSON

## Prochaines Étapes

1. **Supprimer** le projet MMORPGTemplate actuel
2. **Créer** un nouveau projet UE5.6 propre
3. **Suivre** ce plan étape par étape
4. **Me notifier** quand le nouveau projet est créé pour commencer l'implémentation

## Notes Importantes

- Ce plan évite les erreurs de compilation rencontrées précédemment
- Chaque étape est testable individuellement
- La structure modulaire facilite le debugging
- Les types sont tous Blueprint-compatible dès le départ

---
**Date**: 2025-07-24
**Version**: 2.0 (Restart après problèmes de compilation)