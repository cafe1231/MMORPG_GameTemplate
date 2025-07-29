# Phase 3: Core Gameplay Systems - Technical Architecture

## Executive Summary

This document provides the comprehensive technical architecture for Phase 3, implementing core gameplay systems using Go microservices following hexagonal architecture patterns. The architecture supports inventory management, combat, chat, NPCs, and quests while maintaining scalability, security, and performance.

## 1. Service Architecture

### 1.1 Microservices Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         API Gateway                               │
│                    (WebSocket & REST)                             │
└────────────┬────────────────────────────────────────┬────────────┘
             │                                        │
             ▼                                        ▼
┌─────────────────────┐                   ┌─────────────────────┐
│   Game Service      │                   │   Chat Service      │
│  - Combat Engine    │                   │  - Channel Mgmt     │
│  - Skill System     │                   │  - Message Routing  │
│  - Damage Calc      │                   │  - History Storage  │
│  - Death/Respawn    │                   │  - Profanity Filter │
└──────────┬──────────┘                   └──────────┬──────────┘
           │                                         │
           ▼                                         ▼
┌─────────────────────┐                   ┌─────────────────────┐
│  Inventory Service  │                   │   World Service     │
│  - Item Management  │                   │  - NPC Management   │
│  - Equipment System │                   │  - Quest System     │
│  - Loot Generation  │                   │  - Dialog Trees     │
│  - Trade Validation │                   │  - Spawn Control    │
└─────────────────────┘                   └─────────────────────┘
           │                                         │
           └────────────────┬────────────────────────┘
                           ▼
              ┌─────────────────────────┐
              │    Shared Infrastructure │
              │  - PostgreSQL Cluster    │
              │  - Redis Cache          │
              │  - NATS Message Bus     │
              │  - Prometheus Metrics    │
              └─────────────────────────┘
```

### 1.2 Service Responsibilities

#### Game Service
- **Combat Engine**: Damage calculation, hit detection, combat formulas
- **Skill System**: Skill validation, cooldown management, resource consumption
- **Status Effects**: Buffs/debuffs, damage over time, crowd control
- **Death/Respawn**: Death handling, respawn logic, experience loss

#### Inventory Service
- **Item Management**: CRUD operations for items, item validation
- **Equipment System**: Equipment slots, stat calculation, restrictions
- **Loot System**: Loot table processing, drop rate calculation
- **Storage**: Bank/vault management, item stacking logic

#### Chat Service
- **Channel Management**: Global, party, guild, whisper channels
- **Message Routing**: Real-time message delivery via WebSocket
- **History**: Message persistence and retrieval
- **Moderation**: Profanity filter, rate limiting, mute/ban support

#### World Service
- **NPC Management**: NPC spawning, AI behavior, state management
- **Quest System**: Quest state machine, objective tracking, rewards
- **Dialog System**: Branching dialogs, condition checking
- **Shop System**: NPC vendor logic, buy/sell transactions

### 1.3 Inter-Service Communication

```go
// Event-driven communication via NATS
type ServiceEvent struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Service   string                 `json:"service"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// Example: Combat damage event
type CombatDamageEvent struct {
    AttackerID string  `json:"attacker_id"`
    TargetID   string  `json:"target_id"`
    Damage     int32   `json:"damage"`
    DamageType string  `json:"damage_type"`
    Critical   bool    `json:"critical"`
}

// Example: Item acquired event
type ItemAcquiredEvent struct {
    CharacterID string `json:"character_id"`
    ItemID      string `json:"item_id"`
    Quantity    int32  `json:"quantity"`
    Source      string `json:"source"` // "loot", "quest", "trade", etc.
}
```

## 2. Data Models and Schemas

### 2.1 Item System

```go
// Core item model
type Item struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    Description string                 `json:"description" db:"description"`
    Type        ItemType               `json:"type" db:"type"`
    Rarity      ItemRarity             `json:"rarity" db:"rarity"`
    Level       int32                  `json:"level" db:"level"`
    Stats       map[string]int32       `json:"stats" db:"stats"`
    Flags       []string               `json:"flags" db:"flags"`
    MaxStack    int32                  `json:"max_stack" db:"max_stack"`
    Value       int64                  `json:"value" db:"value"`
    Icon        string                 `json:"icon" db:"icon"`
    Model       string                 `json:"model" db:"model"`
}

// Inventory slot
type InventorySlot struct {
    SlotID      int32     `json:"slot_id" db:"slot_id"`
    ItemID      string    `json:"item_id" db:"item_id"`
    Quantity    int32     `json:"quantity" db:"quantity"`
    BoundType   BoundType `json:"bound_type" db:"bound_type"`
    Enchantment *string   `json:"enchantment" db:"enchantment"`
    Durability  *int32    `json:"durability" db:"durability"`
}

// Character inventory
type Inventory struct {
    CharacterID string           `json:"character_id" db:"character_id"`
    Slots       []InventorySlot  `json:"slots" db:"slots"`
    Equipment   map[string]string `json:"equipment" db:"equipment"` // slot_type -> item_id
    Currency    map[string]int64  `json:"currency" db:"currency"`
    UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}
```

### 2.2 Combat State Management

```go
// Combat session
type CombatSession struct {
    ID           string              `json:"id"`
    Participants []CombatParticipant `json:"participants"`
    StartTime    time.Time           `json:"start_time"`
    State        CombatState         `json:"state"`
    TurnOrder    []string            `json:"turn_order"` // participant IDs
    CurrentTurn  int                 `json:"current_turn"`
}

// Combat participant
type CombatParticipant struct {
    ID           string                 `json:"id"`
    Type         ParticipantType        `json:"type"` // "player", "npc", "pet"
    Stats        CombatStats            `json:"stats"`
    StatusEffects []StatusEffect        `json:"status_effects"`
    Cooldowns    map[string]time.Time   `json:"cooldowns"`
    Resources    map[string]int32       `json:"resources"` // hp, mp, energy, etc.
}

// Status effect
type StatusEffect struct {
    ID         string    `json:"id"`
    Type       string    `json:"type"`
    Source     string    `json:"source_id"`
    Duration   int32     `json:"duration"` // in ticks
    Stacks     int32     `json:"stacks"`
    Modifiers  []StatModifier `json:"modifiers"`
    AppliedAt  time.Time `json:"applied_at"`
}
```

### 2.3 Quest State Machine

```go
// Quest definition
type Quest struct {
    ID           string         `json:"id" db:"id"`
    Name         string         `json:"name" db:"name"`
    Description  string         `json:"description" db:"description"`
    Level        int32          `json:"level" db:"level"`
    Type         QuestType      `json:"type" db:"type"`
    Objectives   []QuestObjective `json:"objectives" db:"objectives"`
    Requirements QuestRequirements `json:"requirements" db:"requirements"`
    Rewards      QuestRewards   `json:"rewards" db:"rewards"`
    NPCGiver     string         `json:"npc_giver" db:"npc_giver"`
    NPCCompleter string         `json:"npc_completer" db:"npc_completer"`
}

// Quest progress
type QuestProgress struct {
    CharacterID  string                    `json:"character_id"`
    QuestID      string                    `json:"quest_id"`
    State        QuestState                `json:"state"`
    Objectives   map[string]ObjectiveProgress `json:"objectives"`
    AcceptedAt   time.Time                 `json:"accepted_at"`
    CompletedAt  *time.Time                `json:"completed_at"`
}

// Quest state machine
type QuestState string

const (
    QuestStateAvailable  QuestState = "available"
    QuestStateActive     QuestState = "active"
    QuestStateComplete   QuestState = "complete"
    QuestStateTurnedIn   QuestState = "turned_in"
    QuestStateFailed     QuestState = "failed"
)
```

### 2.4 NPC Data Models

```go
// NPC definition
type NPC struct {
    ID           string            `json:"id" db:"id"`
    Name         string            `json:"name" db:"name"`
    Type         NPCType           `json:"type" db:"type"`
    Level        int32             `json:"level" db:"level"`
    Stats        NPCStats          `json:"stats" db:"stats"`
    Behavior     NPCBehavior       `json:"behavior" db:"behavior"`
    Dialog       []DialogNode      `json:"dialog" db:"dialog"`
    LootTable    *string           `json:"loot_table" db:"loot_table"`
    ShopInventory *ShopInventory   `json:"shop_inventory" db:"shop_inventory"`
    SpawnPoints  []SpawnPoint      `json:"spawn_points" db:"spawn_points"`
}

// NPC instance (spawned in world)
type NPCInstance struct {
    ID          string       `json:"id"`
    TemplateID  string       `json:"template_id"`
    Position    Vector3      `json:"position"`
    Rotation    Quaternion   `json:"rotation"`
    Health      int32        `json:"health"`
    State       NPCState     `json:"state"`
    Target      *string      `json:"target"`
    LastAction  time.Time    `json:"last_action"`
    RespawnTime *time.Time   `json:"respawn_time"`
}

// Dialog system
type DialogNode struct {
    ID          string           `json:"id"`
    Text        string           `json:"text"`
    Conditions  []DialogCondition `json:"conditions"`
    Options     []DialogOption    `json:"options"`
    Actions     []DialogAction    `json:"actions"`
}
```

### 2.5 Chat Message Structure

```go
// Chat message
type ChatMessage struct {
    ID          string       `json:"id" db:"id"`
    Channel     ChatChannel  `json:"channel" db:"channel"`
    SenderID    string       `json:"sender_id" db:"sender_id"`
    SenderName  string       `json:"sender_name" db:"sender_name"`
    Content     string       `json:"content" db:"content"`
    Timestamp   time.Time    `json:"timestamp" db:"timestamp"`
    Recipients  []string     `json:"recipients" db:"recipients"` // for whisper/party
    Flags       []string     `json:"flags" db:"flags"` // "system", "gm", etc.
}

// Chat channel types
type ChatChannel string

const (
    ChatChannelGlobal  ChatChannel = "global"
    ChatChannelZone    ChatChannel = "zone"
    ChatChannelParty   ChatChannel = "party"
    ChatChannelGuild   ChatChannel = "guild"
    ChatChannelWhisper ChatChannel = "whisper"
    ChatChannelSystem  ChatChannel = "system"
    ChatChannelCombat  ChatChannel = "combat"
)
```

## 3. Technical Design Patterns

### 3.1 Event Sourcing for Game State

```go
// Game event base
type GameEvent struct {
    ID          string                 `json:"id"`
    AggregateID string                 `json:"aggregate_id"`
    Type        string                 `json:"type"`
    Version     int64                  `json:"version"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        map[string]interface{} `json:"data"`
}

// Event store interface
type EventStore interface {
    Append(ctx context.Context, events []GameEvent) error
    Load(ctx context.Context, aggregateID string, fromVersion int64) ([]GameEvent, error)
    LoadSnapshot(ctx context.Context, aggregateID string) (*AggregateSnapshot, error)
    SaveSnapshot(ctx context.Context, snapshot AggregateSnapshot) error
}

// Event sourced aggregate
type EventSourcedAggregate struct {
    ID      string
    Version int64
    Events  []GameEvent
}

// Apply events to rebuild state
func (a *EventSourcedAggregate) Apply(event GameEvent) error {
    switch event.Type {
    case "ItemEquipped":
        return a.applyItemEquipped(event)
    case "DamageTaken":
        return a.applyDamageTaken(event)
    // ... other event types
    }
    return nil
}
```

### 3.2 CQRS for Read-Heavy Operations

```go
// Command handler interface
type CommandHandler interface {
    Handle(ctx context.Context, command Command) error
}

// Query handler interface
type QueryHandler interface {
    Handle(ctx context.Context, query Query) (interface{}, error)
}

// Example: Inventory commands and queries
type EquipItemCommand struct {
    CharacterID string
    ItemID      string
    SlotType    string
}

type GetInventoryQuery struct {
    CharacterID string
}

// Read model for inventory (denormalized for performance)
type InventoryReadModel struct {
    CharacterID    string                 `json:"character_id"`
    TotalItems     int32                  `json:"total_items"`
    EquippedItems  map[string]ItemSummary `json:"equipped_items"`
    BagItems       []ItemSummary          `json:"bag_items"`
    TotalValue     int64                  `json:"total_value"`
    LastUpdated    time.Time              `json:"last_updated"`
}
```

### 3.3 State Synchronization Strategies

```go
// State synchronization manager
type StateSyncManager struct {
    updateQueue    chan StateUpdate
    subscriptions  map[string][]Subscription
    conflictResolver ConflictResolver
}

// State update
type StateUpdate struct {
    EntityID   string
    EntityType string
    OldState   interface{}
    NewState   interface{}
    Timestamp  time.Time
    Source     string
}

// Conflict resolution
type ConflictResolver interface {
    Resolve(updates []StateUpdate) (StateUpdate, error)
}

// Last-write-wins resolver
type LastWriteWinsResolver struct{}

func (r *LastWriteWinsResolver) Resolve(updates []StateUpdate) (StateUpdate, error) {
    if len(updates) == 0 {
        return StateUpdate{}, errors.New("no updates to resolve")
    }
    
    // Sort by timestamp and return the latest
    sort.Slice(updates, func(i, j int) bool {
        return updates[i].Timestamp.After(updates[j].Timestamp)
    })
    
    return updates[0], nil
}
```

### 3.4 Caching Strategies

```go
// Multi-level cache
type GameCache struct {
    l1Cache     *ristretto.Cache  // In-memory L1
    l2Cache     *redis.Client     // Redis L2
    persistence Database          // PostgreSQL
}

// Cache-aside pattern
func (c *GameCache) GetItem(ctx context.Context, itemID string) (*Item, error) {
    // Check L1 cache
    if val, found := c.l1Cache.Get(itemID); found {
        return val.(*Item), nil
    }
    
    // Check L2 cache
    var item Item
    err := c.l2Cache.Get(ctx, fmt.Sprintf("item:%s", itemID)).Scan(&item)
    if err == nil {
        c.l1Cache.Set(itemID, &item, 1)
        return &item, nil
    }
    
    // Load from database
    item, err = c.persistence.GetItem(ctx, itemID)
    if err != nil {
        return nil, err
    }
    
    // Update caches
    c.l2Cache.Set(ctx, fmt.Sprintf("item:%s", itemID), &item, 5*time.Minute)
    c.l1Cache.Set(itemID, &item, 1)
    
    return &item, nil
}
```

## 4. API Design

### 4.1 RESTful Endpoints

```yaml
# Inventory Service
/api/v1/inventory:
  /{character_id}:
    get:
      summary: Get character inventory
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Inventory'
    
  /{character_id}/equip:
    post:
      summary: Equip an item
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                item_id: string
                slot_type: string
    
  /{character_id}/items/{item_id}:
    delete:
      summary: Drop/destroy an item
    put:
      summary: Move/split item stack

# Game Service
/api/v1/combat:
  /attack:
    post:
      summary: Perform basic attack
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                target_id: string
                skill_id: string
  
  /skills/{skill_id}/use:
    post:
      summary: Use a skill
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                target_id: string
                position: object

# World Service
/api/v1/npcs:
  /{npc_id}/interact:
    post:
      summary: Interact with NPC
  
  /{npc_id}/dialog:
    post:
      summary: Select dialog option
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                option_id: string

/api/v1/quests:
  get:
    summary: Get available quests
  
  /{quest_id}/accept:
    post:
      summary: Accept a quest
  
  /{quest_id}/complete:
    post:
      summary: Complete quest objectives
  
  /{quest_id}/turn-in:
    post:
      summary: Turn in completed quest
```

### 4.2 WebSocket Events

```go
// WebSocket message types
const (
    // Client -> Server
    WSMsgTypeMove           = "move"
    WSMsgTypeAttack         = "attack"
    WSMsgTypeUseSkill       = "use_skill"
    WSMsgTypeChat           = "chat"
    WSMsgTypeInteract       = "interact"
    WSMsgTypeInventoryAction = "inventory_action"
    
    // Server -> Client
    WSMsgTypeStateUpdate    = "state_update"
    WSMsgTypeCombatEvent    = "combat_event"
    WSMsgTypeChatMessage    = "chat_message"
    WSMsgTypeInventoryUpdate = "inventory_update"
    WSMsgTypeQuestUpdate    = "quest_update"
    WSMsgTypeSystemMessage  = "system_message"
)

// WebSocket message structure
type WSMessage struct {
    Type      string          `json:"type"`
    Timestamp int64           `json:"timestamp"`
    Data      json.RawMessage `json:"data"`
}

// Example: Combat event
type CombatEventData struct {
    Type        string `json:"type"` // "damage", "heal", "miss", "crit"
    Source      string `json:"source"`
    Target      string `json:"target"`
    Value       int32  `json:"value"`
    DamageType  string `json:"damage_type,omitempty"`
    IsCritical  bool   `json:"is_critical,omitempty"`
}
```

### 4.3 Protocol Buffer Definitions

```protobuf
// combat.proto
syntax = "proto3";
package mmorpg.combat;

message CombatAction {
    string action_id = 1;
    string source_id = 2;
    string target_id = 3;
    oneof action {
        BasicAttack basic_attack = 4;
        SkillUse skill_use = 5;
        ItemUse item_use = 6;
    }
    int64 timestamp = 7;
}

message BasicAttack {
    string weapon_id = 1;
}

message SkillUse {
    string skill_id = 1;
    repeated string modifiers = 2;
}

message CombatResult {
    string action_id = 1;
    bool success = 2;
    repeated DamageResult damages = 3;
    repeated StatusEffectResult effects = 4;
    string failure_reason = 5;
}

message DamageResult {
    string target_id = 1;
    int32 damage = 2;
    string damage_type = 3;
    bool is_critical = 4;
    bool is_blocked = 5;
    bool is_dodged = 6;
}
```

## 5. Performance Considerations

### 5.1 Database Sharding Strategy

```go
// Sharding configuration
type ShardConfig struct {
    TotalShards int
    Strategy    ShardStrategy
}

// Shard key calculation
func GetShardKey(entityType string, entityID string) int {
    hash := fnv.New32a()
    hash.Write([]byte(entityID))
    return int(hash.Sum32() % uint32(shardConfig.TotalShards))
}

// Database router
type DatabaseRouter struct {
    shards map[int]Database
}

func (r *DatabaseRouter) GetDatabase(entityType string, entityID string) Database {
    shardKey := GetShardKey(entityType, entityID)
    return r.shards[shardKey]
}

// Sharding by entity type
var shardingRules = map[string]ShardingRule{
    "character": {Strategy: "hash", Key: "character_id"},
    "inventory": {Strategy: "hash", Key: "character_id"},
    "quest":     {Strategy: "hash", Key: "character_id"},
    "chat":      {Strategy: "range", Key: "timestamp"},
}
```

### 5.2 Redis Caching Patterns

```go
// Cache warming
type CacheWarmer struct {
    cache    Cache
    database Database
}

func (w *CacheWarmer) WarmCache(ctx context.Context) error {
    // Preload frequently accessed data
    items, err := w.database.GetPopularItems(ctx, 1000)
    if err != nil {
        return err
    }
    
    for _, item := range items {
        key := fmt.Sprintf("item:%s", item.ID)
        w.cache.Set(ctx, key, item, 1*time.Hour)
    }
    
    return nil
}

// Cache invalidation
type CacheInvalidator struct {
    cache Cache
    bus   MessageBus
}

func (i *CacheInvalidator) Start(ctx context.Context) {
    sub := i.bus.Subscribe("cache.invalidate.*")
    for msg := range sub {
        var event CacheInvalidateEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            continue
        }
        
        i.cache.Delete(ctx, event.Key)
    }
}
```

### 5.3 Interest Management

```go
// Interest management for position updates
type InterestManager struct {
    grid       *SpatialGrid
    interests  map[string]*InterestArea
    updateRate time.Duration
}

// Interest area
type InterestArea struct {
    EntityID string
    Position Vector3
    Radius   float64
    Entities map[string]bool
}

// Update interest areas
func (m *InterestManager) UpdateInterests(entityID string, position Vector3) {
    area := m.interests[entityID]
    if area == nil {
        area = &InterestArea{
            EntityID: entityID,
            Radius:   100.0, // 100 units
            Entities: make(map[string]bool),
        }
        m.interests[entityID] = area
    }
    
    // Find entities in range
    oldEntities := area.Entities
    newEntities := m.grid.GetEntitiesInRadius(position, area.Radius)
    
    // Send enter events
    for id := range newEntities {
        if !oldEntities[id] {
            m.sendEnterEvent(entityID, id)
        }
    }
    
    // Send leave events
    for id := range oldEntities {
        if !newEntities[id] {
            m.sendLeaveEvent(entityID, id)
        }
    }
    
    area.Position = position
    area.Entities = newEntities
}
```

### 5.4 Load Balancing

```go
// Service discovery and load balancing
type ServiceRegistry struct {
    consul *consul.Client
}

func (r *ServiceRegistry) GetService(name string) ([]ServiceInstance, error) {
    services, _, err := r.consul.Health().Service(name, "", true, nil)
    if err != nil {
        return nil, err
    }
    
    instances := make([]ServiceInstance, len(services))
    for i, svc := range services {
        instances[i] = ServiceInstance{
            ID:      svc.Service.ID,
            Address: svc.Service.Address,
            Port:    svc.Service.Port,
            Weight:  getWeight(svc.Service.Meta),
        }
    }
    
    return instances, nil
}

// Weighted round-robin load balancer
type LoadBalancer struct {
    instances []ServiceInstance
    current   int
    mu        sync.RWMutex
}

func (lb *LoadBalancer) Next() *ServiceInstance {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    if len(lb.instances) == 0 {
        return nil
    }
    
    instance := &lb.instances[lb.current]
    lb.current = (lb.current + 1) % len(lb.instances)
    
    return instance
}
```

## 6. Security Architecture

### 6.1 Anti-Cheat Considerations

```go
// Action validation
type ActionValidator struct {
    rules map[string]ValidationRule
}

func (v *ActionValidator) Validate(action GameAction) error {
    rule, exists := v.rules[action.Type]
    if !exists {
        return ErrUnknownAction
    }
    
    // Check rate limiting
    if err := v.checkRateLimit(action); err != nil {
        return err
    }
    
    // Validate action parameters
    if err := rule.Validate(action); err != nil {
        return err
    }
    
    // Check for impossible actions
    if err := v.checkPhysics(action); err != nil {
        return err
    }
    
    return nil
}

// Movement validation
func (v *ActionValidator) checkMovement(oldPos, newPos Vector3, deltaTime float64) error {
    distance := oldPos.Distance(newPos)
    maxSpeed := 10.0 // units per second
    maxDistance := maxSpeed * deltaTime
    
    if distance > maxDistance*1.1 { // 10% tolerance
        return ErrImpossibleMovement
    }
    
    return nil
}
```

### 6.2 Input Validation

```go
// Input sanitization middleware
func InputValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Validate content length
        if c.Request.ContentLength > MaxRequestSize {
            c.AbortWithStatus(http.StatusRequestEntityTooLarge)
            return
        }
        
        // Validate headers
        if err := validateHeaders(c.Request.Header); err != nil {
            c.AbortWithStatus(http.StatusBadRequest)
            return
        }
        
        c.Next()
    }
}

// Chat message validation
func ValidateChatMessage(msg *ChatMessage) error {
    // Length check
    if len(msg.Content) > MaxChatMessageLength {
        return ErrMessageTooLong
    }
    
    // Character validation
    if !utf8.ValidString(msg.Content) {
        return ErrInvalidCharacters
    }
    
    // Command injection prevention
    if containsSQLKeywords(msg.Content) {
        return ErrSuspiciousContent
    }
    
    return nil
}
```

### 6.3 Rate Limiting

```go
// Rate limiter
type RateLimiter struct {
    store     RateLimitStore
    rules     map[string]RateLimitRule
}

type RateLimitRule struct {
    Window    time.Duration
    MaxCount  int
    BurstSize int
}

func (rl *RateLimiter) Allow(key string, action string) (bool, error) {
    rule, exists := rl.rules[action]
    if !exists {
        return true, nil // No limit defined
    }
    
    count, err := rl.store.Increment(key, rule.Window)
    if err != nil {
        return false, err
    }
    
    return count <= rule.MaxCount, nil
}

// Game action rate limits
var gameActionLimits = map[string]RateLimitRule{
    "combat.attack":     {Window: 1 * time.Second, MaxCount: 2},
    "combat.skill":      {Window: 1 * time.Second, MaxCount: 1},
    "chat.message":      {Window: 1 * time.Minute, MaxCount: 30},
    "inventory.move":    {Window: 1 * time.Second, MaxCount: 10},
    "quest.accept":      {Window: 1 * time.Second, MaxCount: 1},
    "trade.request":     {Window: 1 * time.Minute, MaxCount: 5},
}
```

### 6.4 Item/Currency Security

```go
// Secure item transaction
type ItemTransaction struct {
    ID          string
    Type        TransactionType
    Source      string
    Destination string
    Items       []ItemTransfer
    Currency    map[string]int64
    Timestamp   time.Time
    Signature   string
}

// Transaction processor
type TransactionProcessor struct {
    validator TransactionValidator
    ledger    TransactionLedger
}

func (p *TransactionProcessor) Process(tx ItemTransaction) error {
    // Validate transaction
    if err := p.validator.Validate(tx); err != nil {
        return err
    }
    
    // Begin distributed transaction
    dtx := p.ledger.Begin()
    defer dtx.Rollback()
    
    // Lock involved inventories
    if err := dtx.LockInventories(tx.Source, tx.Destination); err != nil {
        return err
    }
    
    // Verify source has items
    if err := dtx.VerifySource(tx); err != nil {
        return err
    }
    
    // Transfer items
    if err := dtx.Transfer(tx); err != nil {
        return err
    }
    
    // Record in ledger
    if err := dtx.RecordTransaction(tx); err != nil {
        return err
    }
    
    // Commit
    return dtx.Commit()
}
```

## 7. Implementation Guidelines

### 7.1 Service Structure

Each microservice follows hexagonal architecture:

```
service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities.go
│   │   ├── services.go
│   │   └── errors.go
│   ├── ports/
│   │   ├── repositories.go
│   │   ├── services.go
│   │   └── handlers.go
│   ├── adapters/
│   │   ├── postgres/
│   │   ├── redis/
│   │   ├── http/
│   │   └── grpc/
│   └── application/
│       └── service.go
├── pkg/
│   └── proto/
└── config/
    └── config.yaml
```

### 7.2 Development Workflow

1. **Define Domain Models**: Start with core entities and business logic
2. **Create Port Interfaces**: Define repository and service contracts
3. **Implement Adapters**: Build database, cache, and API adapters
4. **Wire Application**: Connect ports and adapters in main.go
5. **Add Observability**: Integrate logging, metrics, and tracing
6. **Write Tests**: Unit tests for domain, integration tests for adapters

### 7.3 Testing Strategy

```go
// Unit test example
func TestCombatDamageCalculation(t *testing.T) {
    calc := NewDamageCalculator()
    
    attacker := &CombatStats{
        AttackPower: 100,
        CritChance:  0.2,
    }
    
    defender := &CombatStats{
        Defense: 50,
        DodgeChance: 0.1,
    }
    
    result := calc.Calculate(attacker, defender)
    
    assert.GreaterOrEqual(t, result.Damage, 25)
    assert.LessOrEqual(t, result.Damage, 75)
}

// Integration test example
func TestInventoryService_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Create service
    repo := postgres.NewInventoryRepository(db)
    cache := redis.NewInventoryCache(setupTestRedis(t))
    service := NewInventoryService(repo, cache)
    
    // Test inventory operations
    inv, err := service.GetInventory(ctx, "test-char-id")
    assert.NoError(t, err)
    assert.NotNil(t, inv)
}
```

## 8. Deployment Architecture

```yaml
# docker-compose.yml for development
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: mmorpg
      POSTGRES_USER: mmorpg
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  nats:
    image: nats:2.9-alpine
    command: -js -m 8222
    ports:
      - "4222:4222"
      - "8222:8222"

  game-service:
    build: ./services/game
    environment:
      DATABASE_URL: postgres://mmorpg:secret@postgres/mmorpg
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
    ports:
      - "8081:8080"

  inventory-service:
    build: ./services/inventory
    environment:
      DATABASE_URL: postgres://mmorpg:secret@postgres/mmorpg
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
    ports:
      - "8082:8080"

  chat-service:
    build: ./services/chat
    environment:
      DATABASE_URL: postgres://mmorpg:secret@postgres/mmorpg
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
    ports:
      - "8083:8080"

  world-service:
    build: ./services/world
    environment:
      DATABASE_URL: postgres://mmorpg:secret@postgres/mmorpg
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
    ports:
      - "8084:8080"

volumes:
  postgres_data:
  redis_data:
```

## 9. Monitoring and Observability

### 9.1 Metrics

```go
// Prometheus metrics
var (
    combatDamageTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_combat_damage_total",
            Help: "Total damage dealt in combat",
        },
        []string{"damage_type", "source_type", "target_type"},
    )
    
    inventoryOperationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "inventory_operation_duration_seconds",
            Help: "Duration of inventory operations",
        },
        []string{"operation", "status"},
    )
    
    activeConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "websocket_active_connections",
            Help: "Number of active WebSocket connections",
        },
        []string{"service"},
    )
)
```

### 9.2 Distributed Tracing

```go
// OpenTelemetry tracing
func (s *GameService) Attack(ctx context.Context, req *AttackRequest) (*AttackResponse, error) {
    ctx, span := tracer.Start(ctx, "GameService.Attack")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("attacker_id", req.AttackerID),
        attribute.String("target_id", req.TargetID),
        attribute.String("skill_id", req.SkillID),
    )
    
    // Validate request
    if err := s.validator.ValidateAttack(ctx, req); err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    // Calculate damage
    damage, err := s.calculator.Calculate(ctx, req)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    span.SetAttributes(
        attribute.Int64("damage", int64(damage.Amount)),
        attribute.Bool("critical", damage.IsCritical),
    )
    
    return &AttackResponse{Damage: damage}, nil
}
```

## 10. Migration Path

### Phase 3A Implementation Order:
1. **Week 1**: Set up infrastructure (databases, message bus, service scaffolding)
2. **Week 2**: Implement Inventory Service (core CRUD, equipment system)
3. **Week 3**: Implement Game Service (combat engine, damage calculation)
4. **Week 4**: Implement Chat Service (message routing, channel management)
5. **Week 5**: Implement World Service (NPC spawning, quest state machine)

### Phase 3B Integration Order:
1. **Week 1**: Integrate Inventory UI with backend
2. **Week 2**: Integrate Combat System with animations
3. **Week 3**: Integrate Chat UI and commands
4. **Week 4**: Integrate NPC interactions and quest UI
5. **Week 5**: End-to-end testing and optimization

## Conclusion

This architecture provides a scalable, secure, and maintainable foundation for Phase 3's core gameplay systems. The microservices approach allows independent scaling and development, while the hexagonal architecture ensures clean separation of concerns. The use of Go provides excellent performance and concurrency support, essential for real-time gameplay.

Key architectural decisions:
- **Event-driven communication** via NATS for loose coupling
- **CQRS pattern** for optimized read/write operations
- **Multi-level caching** for performance
- **Event sourcing** for game state auditability
- **Comprehensive security** measures against common exploits

This design supports the immediate Phase 3 requirements while providing a foundation for future features like PvP, guilds, and advanced combat systems.