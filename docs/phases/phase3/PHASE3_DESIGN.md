# Phase 3 - Design - Core Gameplay Systems Architecture

## Architecture Philosophy

### Design Principles

1. **Modular Independence** - Each gameplay system operates independently while sharing common interfaces
2. **Data-Driven Design** - Gameplay configuration through external files, not hardcoded values
3. **Event-Driven Communication** - Systems communicate through events, reducing tight coupling
4. **Predictive Responsiveness** - Client-side prediction with server reconciliation for smooth gameplay
5. **Extensibility by Default** - Every system designed to be extended without modifying core code

### Technical Goals

- **Minimal Latency Impact** - Gameplay should feel responsive even on 200ms connections
- **Efficient Synchronization** - Only send necessary updates to relevant clients
- **Robust Error Handling** - Graceful degradation when systems fail
- **Developer Productivity** - Clear APIs and comprehensive debugging tools
- **Performance at Scale** - Maintain 60+ FPS with all systems active

## System Component Design

### 1. Inventory Management System

#### Frontend Components

```
UE5 Inventory Architecture
├── Core Components
│   ├── UInventoryComponent           # Actor component for inventory
│   ├── UItemDatabase                 # Singleton item definition manager
│   ├── FInventorySlot               # Struct for slot data
│   └── FItemInstance                # Runtime item with unique properties
├── UI Components
│   ├── WBP_Inventory                # Main inventory widget
│   ├── WBP_InventorySlot           # Individual slot widget
│   ├── WBP_ItemTooltip             # Hover information display
│   └── WBP_EquipmentPanel          # Character equipment view
└── Controllers
    ├── AInventoryController         # Handles drag/drop logic
    ├── UInventorySubsystem         # Game instance subsystem
    └── UItemInteractionComponent   # World item pickup
```

#### Backend Services

```
Inventory Service Architecture
├── API Endpoints
│   ├── GET  /inventory/{characterId}      # Fetch full inventory
│   ├── POST /inventory/move               # Move item between slots
│   ├── POST /inventory/equip              # Equip/unequip item
│   ├── POST /inventory/drop               # Drop item to world
│   └── POST /inventory/use                # Consume/use item
├── Domain Logic
│   ├── InventoryManager                   # Core inventory operations
│   ├── ItemValidator                      # Validate item operations
│   ├── EquipmentCalculator               # Calculate equipment stats
│   └── InventorySerializer               # Optimize network payload
└── Data Models
    ├── Inventory                         # Container with slots
    ├── Item                              # Base item definition
    ├── ItemInstance                      # Unique item in world
    └── EquipmentSlots                    # Character equipment
```

#### Data Flow Diagram

```
┌─────────────┐     ┌───────────────┐     ┌─────────────┐
│   Client    │     │    Server     │     │  Database   │
│    (UE5)    │     │   (Node.js)   │     │ (PostgreSQL)│
└─────┬───────┘     └───────┬───────┘     └──────┬──────┘
      │                     │                     │
      │ 1. Drag Item        │                     │
      ├────────────────────>│                     │
      │                     │ 2. Validate Move    │
      │                     ├────────────────────>│
      │                     │                     │
      │                     │ 3. Update DB        │
      │                     │<────────────────────┤
      │ 4. Confirm Move     │                     │
      │<────────────────────┤                     │
      │                     │                     │
      │ 5. Update UI        │                     │
      │                     │                     │
```

### 2. Combat System

#### Frontend Components

```
UE5 Combat Architecture
├── Core Systems
│   ├── UCombatComponent             # Main combat logic
│   ├── UTargetingComponent          # Target selection/management
│   ├── UDamageCalculator           # Client-side damage preview
│   └── UCombatAnimationComponent   # Animation state machine
├── UI Elements
│   ├── WBP_TargetFrame             # Target health/info
│   ├── WBP_DamageNumbers           # Floating damage text
│   ├── WBP_CombatLog               # Combat event history
│   └── WBP_HealthBar               # Player/NPC health bars
└── Effects
    ├── UCombatEffectsManager       # VFX/SFX coordination
    ├── FHitReaction                # Hit response system
    └── UDeathHandler               # Death/respawn logic
```

#### Backend Services

```
Combat Service Architecture
├── Real-time Processing
│   ├── CombatEngine                # Core combat calculations
│   ├── TargetValidator             # Line of sight, range checks
│   ├── DamageProcessor             # Apply damage with modifiers
│   └── CombatLogger                # Track combat statistics
├── Combat Flow
│   ├── AttackRequest               # Validate attack attempt
│   ├── DamageCalculation          # Server-side damage math
│   ├── EffectApplication          # Buffs/debuffs/conditions
│   └── DeathProcessing            # Handle entity death
└── Synchronization
    ├── CombatStateSync            # Sync combat states
    ├── HealthUpdater              # Broadcast health changes
    └── CombatEventStream          # Real-time combat events
```

#### Combat Sequence Diagram

```
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│ Player │ │ Client │ │ Server │ │Combat  │ │ Target │
│        │ │  (UE5) │ │ (Node) │ │Engine  │ │ (NPC)  │
└───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘
    │          │          │          │          │
    │ Attack   │          │          │          │
    ├─────────>│          │          │          │
    │          │ Predict  │          │          │
    │          ├──────────┤          │          │
    │          │ Send Req │          │          │
    │          ├─────────>│          │          │
    │          │          │ Validate │          │
    │          │          ├─────────>│          │
    │          │          │          │ Calculate│
    │          │          │          ├─────────>│
    │          │          │          │ Damage   │
    │          │          │          │<─────────┤
    │          │          │ Apply    │          │
    │          │          │<─────────┤          │
    │          │ Result   │          │          │
    │          │<─────────┤          │          │
    │ Update   │          │          │          │
    │<─────────┤          │          │          │
    │          │          │          │          │
```

### 3. Chat System

#### Frontend Components

```
UE5 Chat Architecture
├── Core Components
│   ├── UChatSubsystem              # Game instance subsystem
│   ├── FChatMessage                # Message data structure
│   ├── UChatCommandParser          # Parse slash commands
│   └── UChatChannelManager         # Manage active channels
├── UI Components
│   ├── WBP_ChatWindow              # Main chat interface
│   ├── WBP_ChatTab                 # Channel tab widget
│   ├── WBP_ChatInput               # Input field with completion
│   └── WBP_ChatOptions             # Settings and filters
└── Integration
    ├── UChatNotificationComponent  # System messages
    ├── UChatHistoryManager         # Persistent history
    └── UChatFilterComponent        # Profanity/spam filter
```

#### Backend Services

```
Chat Service Architecture
├── Message Handling
│   ├── MessageRouter               # Route to appropriate channel
│   ├── ChannelManager              # Manage channel subscriptions
│   ├── MessageValidator            # Spam/profanity checks
│   └── CommandProcessor            # Handle slash commands
├── Channels
│   ├── GlobalChannel               # Zone/world chat
│   ├── PartyChannel                # Group members only
│   ├── WhisperChannel              # Private messages
│   └── SystemChannel               # Server announcements
└── Features
    ├── MessageHistory              # Store recent messages
    ├── RateLimiter                 # Prevent spam
    └── NotificationService         # Offline message alerts
```

### 4. NPC System

#### Frontend Components

```
UE5 NPC Architecture
├── Core Systems
│   ├── ANPCCharacter               # Base NPC actor class
│   ├── UNPCInteractionComponent    # Handle player interaction
│   ├── UNPCBehaviorComponent       # AI behavior tree runner
│   └── UNPCDialogComponent         # Dialog tree processor
├── NPC Types
│   ├── AShopNPC                    # Merchant NPCs
│   ├── AQuestNPC                   # Quest giver NPCs
│   ├── AGuardNPC                   # Combat NPCs
│   └── AAmbientNPC                 # Background NPCs
└── UI Components
    ├── WBP_NPCDialog               # Dialog interface
    ├── WBP_ShopInterface           # Buy/sell UI
    ├── WBP_QuestDialog             # Quest offer/complete
    └── WBP_NPCNameplate            # Overhead display
```

#### Backend Services

```
NPC Service Architecture
├── NPC Management
│   ├── NPCSpawnManager             # Control NPC spawning
│   ├── NPCStateManager             # Track NPC states
│   ├── NPCInventoryManager         # Shop inventories
│   └── NPCDialogManager            # Dialog tree logic
├── Interaction Handling
│   ├── InteractionValidator        # Check interaction range
│   ├── ShopTransactionProcessor    # Handle buy/sell
│   ├── DialogProgressTracker       # Track dialog choices
│   └── NPCQuestInterface           # Quest-related interactions
└── AI Systems
    ├── PathfindingService          # NPC movement
    ├── BehaviorTreeRunner          # Execute AI behaviors
    └── NPCScheduler                # Time-based activities
```

### 5. Quest System

#### Frontend Components

```
UE5 Quest Architecture
├── Core Systems
│   ├── UQuestSubsystem             # Main quest manager
│   ├── FQuest                      # Quest data structure
│   ├── UQuestObjective             # Individual objectives
│   └── UQuestTracker               # Progress tracking
├── UI Components
│   ├── WBP_QuestLog                # Full quest journal
│   ├── WBP_QuestTracker            # Active quest HUD
│   ├── WBP_QuestDetails            # Detailed quest view
│   └── WBP_QuestRewards            # Reward preview
└── Integration
    ├── UQuestGiverComponent        # NPC quest offering
    ├── UQuestTriggerVolume         # Location-based quests
    └── UQuestItemComponent         # Quest item tracking
```

#### Backend Services

```
Quest Service Architecture
├── Quest Management
│   ├── QuestDefinitionLoader       # Load quest templates
│   ├── QuestInstanceManager        # Player quest instances
│   ├── QuestProgressTracker        # Track objectives
│   └── QuestCompletionHandler      # Handle completion
├── Objective Types
│   ├── KillObjective               # Eliminate X enemies
│   ├── CollectObjective            # Gather Y items
│   ├── DeliveryObjective           # Bring item to NPC
│   ├── LocationObjective           # Reach specific area
│   └── InteractionObjective        # Use/activate object
└── Features
    ├── QuestChainManager           # Sequential quests
    ├── QuestPrerequisiteChecker    # Check requirements
    └── RewardDistributor           # Grant quest rewards
```

## UI/UX Design

### Inventory Interface

```
┌─────────────────────────────────────────────────────────┐
│ Inventory                                    [X] Close   │
├─────────────────────────────────────────────────────────┤
│ ┌─────────────────────┐ ┌────────────────────────────┐ │
│ │   Character Model    │ │  Grid Inventory (8x10)     │ │
│ │                      │ │ ┌──┐┌──┐┌──┐┌──┐┌──┐┌──┐ │ │
│ │  [Head   ] [    ]    │ │ │SW││  ││  ││HP││  ││  │ │ │
│ │  [Chest  ] [Ring1]   │ │ └──┘└──┘└──┘└──┘└──┘└──┘ │ │
│ │  [Legs   ] [Ring2]   │ │ ┌──┐┌──┐┌──┐┌──┐┌──┐┌──┐ │ │
│ │  [Feet   ] [    ]    │ │ │  ││  ││  ││  ││  ││  │ │ │
│ │  [Main   ] [Off ]    │ │ └──┘└──┘└──┘└──┘└──┘└──┘ │ │
│ │                      │ │         ...                 │ │
│ │ Stats:               │ │                             │ │
│ │ ATK: 125  DEF: 89    │ │ Weight: 45/100 kg          │ │
│ │ HP: 1200  MP: 340    │ │ Gold: 1,547                │ │
│ └─────────────────────┘ └────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘

Legend: SW=Sword, HP=Health Potion, [ ]=Empty Slot
```

### Chat Window Design

```
┌─────────────────────────────────────────────────────────┐
│ [Global] [Party] [Combat] [System]           [-][□][X]  │
├─────────────────────────────────────────────────────────┤
│ [14:32] [Global] Player1: Anyone want to party?         │
│ [14:32] [System] Player2 has joined the zone.           │
│ [14:33] [Party] TankGuy: Ready for dungeon              │
│ [14:33] [Global] Merchant: Selling potions cheap!       │
│ [14:34] [Combat] You hit Goblin for 45 damage.          │
│ [14:34] [Combat] Goblin hits you for 12 damage.         │
│ [14:35] [Whisper] From Friend: Hey, where are you?      │
│                                                          │
│                                                          │
├─────────────────────────────────────────────────────────┤
│ [/w Friend │                                    ] [Send] │
└─────────────────────────────────────────────────────────┘
```

### Quest Log Layout

```
┌─────────────────────────────────────────────────────────┐
│ Quest Log                                    [X] Close   │
├─────────────────────────────────────────────────────────┤
│ ┌─────────────────┐ ┌─────────────────────────────────┐ │
│ │ Active Quests   │ │ The Missing Merchant           │ │
│ │                 │ │ ────────────────────────────── │ │
│ │ ▼ Main Story    │ │ The merchant hasn't returned   │ │
│ │   • Missing     │ │ from his journey. His wife is  │ │
│ │     Merchant    │ │ worried. Find out what         │ │
│ │                 │ │ happened.                      │ │
│ │ ▼ Side Quests   │ │                                │ │
│ │   • Wolf Pelts  │ │ Objectives:                    │ │
│ │   • Herb Garden │ │ ☑ Talk to merchant's wife      │ │
│ │                 │ │ ☐ Search the road (0/3)        │ │
│ │ ▼ Daily         │ │ ☐ Find the merchant            │ │
│ │   • Monster     │ │                                │ │
│ │     Hunt        │ │ Rewards:                       │ │
│ │                 │ │ • 500 XP                       │ │
│ │ Completed (5)   │ │ • 50 Gold                      │ │
│ └─────────────────┘ │ • Merchant's Favor (Rep)       │ │
│                     └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### NPC Interaction Flow

```
┌─────────────────────────────────────────────────────────┐
│                  Blacksmith Johnson                      │
├─────────────────────────────────────────────────────────┤
│ "Welcome to my shop! I've got the finest weapons in     │
│ the land. What can I do for you?"                      │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ [1] Show me your wares                              │ │
│ │ [2] Can you repair my equipment?                    │ │
│ │ [3] Tell me about the town                          │ │
│ │ [4] I have a quest item for you                     │ │
│ │ [5] Goodbye                                         │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Network Protocol Design

### Message Formats

#### Inventory Update Message
```protobuf
message InventoryUpdate {
  enum UpdateType {
    FULL_SYNC = 0;
    ITEM_ADD = 1;
    ITEM_REMOVE = 2;
    ITEM_MOVE = 3;
    ITEM_UPDATE = 4;
  }
  
  UpdateType type = 1;
  repeated InventorySlot slots = 2;
  uint32 version = 3;  // For conflict resolution
  int64 timestamp = 4;
}

message InventorySlot {
  uint32 slot_id = 1;
  ItemInstance item = 2;
  uint32 quantity = 3;
}
```

#### Combat Action Message
```protobuf
message CombatAction {
  enum ActionType {
    BASIC_ATTACK = 0;
    SKILL_USE = 1;
    DEFEND = 2;
    FLEE = 3;
  }
  
  string attacker_id = 1;
  string target_id = 2;
  ActionType action = 3;
  uint32 skill_id = 4;  // If skill
  int64 timestamp = 5;
}

message CombatResult {
  string action_id = 1;
  bool success = 2;
  repeated DamageEvent damages = 3;
  repeated StatusEffect effects = 4;
}
```

### Update Frequencies

| System | Update Rate | Priority | Conditions |
|--------|-------------|----------|------------|
| Movement | 10 Hz | High | When moving |
| Combat | 20 Hz | Critical | During combat |
| Inventory | On-demand | Medium | On change |
| Chat | Instant | Medium | On message |
| Quest | On-demand | Low | On progress |
| NPC State | 2 Hz | Low | When in view |

### Optimization Strategies

1. **Delta Compression**
   - Send only changed data, not full states
   - Use bitfields for efficient packing
   - Compress similar updates together

2. **Interest Management**
   - Only send updates for nearby entities
   - Reduce update frequency for distant objects
   - Prioritize updates based on relevance

3. **Message Batching**
   - Combine multiple small updates
   - Send batches at fixed intervals
   - Reduce packet overhead

4. **Client Prediction**
   - Predict movement and actions locally
   - Reconcile with server authoritative state
   - Smooth interpolation for other players

## Database Design

### Schema Overview

```sql
-- Items and Inventory
CREATE TABLE items (
  id UUID PRIMARY KEY,
  item_template_id VARCHAR(50) NOT NULL,
  owner_id UUID REFERENCES characters(id),
  container_id UUID,  -- NULL if in world
  slot_id INT,
  quantity INT DEFAULT 1,
  durability INT,
  custom_stats JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Combat logs for analysis
CREATE TABLE combat_logs (
  id UUID PRIMARY KEY,
  attacker_id UUID,
  target_id UUID,
  action_type VARCHAR(20),
  damage_dealt INT,
  combat_data JSONB,
  timestamp TIMESTAMP DEFAULT NOW()
);

-- Chat messages (recent only)
CREATE TABLE chat_messages (
  id UUID PRIMARY KEY,
  channel VARCHAR(20),
  sender_id UUID REFERENCES characters(id),
  message TEXT,
  timestamp TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_chat_timestamp ON chat_messages(timestamp);

-- NPC states
CREATE TABLE npc_states (
  npc_id UUID PRIMARY KEY,
  position POINT,
  health INT,
  behavior_state VARCHAR(50),
  custom_data JSONB,
  last_updated TIMESTAMP DEFAULT NOW()
);

-- Quest progress
CREATE TABLE character_quests (
  character_id UUID REFERENCES characters(id),
  quest_id VARCHAR(50),
  status VARCHAR(20),
  objectives_data JSONB,
  completed_at TIMESTAMP,
  PRIMARY KEY (character_id, quest_id)
);
```

### Indexing Strategy

```sql
-- Performance critical indexes
CREATE INDEX idx_items_owner ON items(owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX idx_items_container ON items(container_id) WHERE container_id IS NOT NULL;
CREATE INDEX idx_combat_logs_timestamp ON combat_logs(timestamp DESC);
CREATE INDEX idx_quest_status ON character_quests(character_id, status);

-- Spatial index for NPCs
CREATE INDEX idx_npc_position ON npc_states USING GIST(position);

-- Full text search for chat
CREATE INDEX idx_chat_message_search ON chat_messages USING GIN(to_tsvector('english', message));
```

### Query Patterns

```sql
-- Get player inventory (optimized)
SELECT i.*, t.* 
FROM items i
JOIN item_templates t ON i.item_template_id = t.id
WHERE i.owner_id = $1 AND i.container_id IS NULL
ORDER BY i.slot_id;

-- Find NPCs in area (spatial query)
SELECT * FROM npc_states
WHERE position <@ box(point($1,$2), point($3,$4))
AND last_updated > NOW() - INTERVAL '5 minutes';

-- Get active quests with progress
SELECT q.*, cq.objectives_data, cq.status
FROM quests q
JOIN character_quests cq ON q.id = cq.quest_id
WHERE cq.character_id = $1 AND cq.status IN ('active', 'ready_to_complete');
```

## Integration Points

### With Authentication System (Phase 1)

```typescript
// All gameplay requests include JWT
interface AuthorizedRequest {
  headers: {
    authorization: string; // Bearer JWT
  };
  userId: string;        // Extracted from JWT
  characterId: string;   // Active character
}

// Middleware validates every request
async function validateGameplayRequest(req: Request) {
  const token = validateJWT(req.headers.authorization);
  const character = await getActiveCharacter(token.userId);
  if (!character) throw new Error('No active character');
  req.characterId = character.id;
}
```

### With Networking Layer (Phase 2)

```typescript
// Extend WebSocket connection with gameplay context
interface GameplayConnection extends WSConnection {
  characterId: string;
  subscribedChannels: Set<string>;
  combatState?: CombatSession;
  nearbyEntities: Set<string>;
}

// Message routing based on gameplay state
function routeGameplayMessage(conn: GameplayConnection, msg: Message) {
  switch(msg.type) {
    case 'inventory_action':
      return inventoryService.handle(conn, msg);
    case 'combat_action':
      return combatService.handle(conn, msg);
    case 'chat_message':
      return chatService.handle(conn, msg);
  }
}
```

### Between Gameplay Systems

```typescript
// Event-driven communication between systems
class GameplayEventBus {
  // Quest system listens to combat events
  on('enemy_killed', (event: EnemyKilledEvent) => {
    questSystem.updateKillObjectives(event);
  });
  
  // Inventory system notifies quest system
  on('item_acquired', (event: ItemAcquiredEvent) => {
    questSystem.updateCollectObjectives(event);
  });
  
  // NPC system triggers quest updates
  on('npc_interaction', (event: NPCInteractionEvent) => {
    if (event.type === 'quest_complete') {
      questSystem.completeQuest(event);
      inventorySystem.grantRewards(event.rewards);
    }
  });
}
```

## Extension Points

### How to Add New Item Types

```typescript
// 1. Define item template
{
  "id": "custom_sword_01",
  "name": "Flaming Sword",
  "type": "weapon",
  "subtype": "sword",
  "stats": {
    "damage": 50,
    "fire_damage": 10
  },
  "requirements": {
    "level": 10,
    "strength": 15
  }
}

// 2. Extend item behavior
class FlamingSword extends BaseWeapon {
  onEquip(character: Character) {
    super.onEquip(character);
    character.addAura('fire_aura');
  }
  
  onHit(target: Character, damage: number) {
    // Apply burn effect
    target.addDebuff('burn', 5, 3); // 5 damage for 3 seconds
  }
}

// 3. Register with item factory
ItemFactory.register('flaming_sword', FlamingSword);
```

### How to Create Custom Quests

```typescript
// 1. Define quest in JSON
{
  "id": "custom_quest_01",
  "name": "The Haunted Forest",
  "description": "Investigate strange sounds in the forest",
  "objectives": [
    {
      "type": "location",
      "target": "haunted_forest_center",
      "description": "Reach the center of the forest"
    },
    {
      "type": "custom",
      "handler": "investigate_sounds",
      "description": "Investigate the mysterious sounds"
    }
  ]
}

// 2. Implement custom objective handler
class InvestigateSoundsObjective extends QuestObjective {
  async checkProgress(character: Character): Promise<boolean> {
    // Custom logic for investigation
    const nearbyEvents = await getEventsNearCharacter(character);
    return nearbyEvents.some(e => e.type === 'mysterious_sound');
  }
}

// 3. Register objective type
QuestSystem.registerObjectiveType('investigate_sounds', InvestigateSoundsObjective);
```

### How to Extend Combat Formulas

```typescript
// 1. Create custom damage calculator
class CustomDamageCalculator extends BaseDamageCalculator {
  calculate(attacker: Character, target: Character, skill?: Skill): DamageResult {
    let damage = super.calculate(attacker, target, skill);
    
    // Add custom mechanics
    if (attacker.hasBlessing('warrior_spirit')) {
      damage.physical *= 1.2;
    }
    
    if (target.isType('undead') && damage.holy > 0) {
      damage.holy *= 2; // Double holy damage vs undead
    }
    
    // Weather effects
    const weather = getZoneWeather(attacker.zoneId);
    if (weather === 'rain' && damage.fire > 0) {
      damage.fire *= 0.5; // Reduce fire damage in rain
    }
    
    return damage;
  }
}

// 2. Register calculator
CombatSystem.setDamageCalculator(new CustomDamageCalculator());
```

---

*This design document provides the technical blueprint for implementing Phase 3 gameplay systems. All implementation should follow these architectural patterns while maintaining flexibility for game-specific customization.*