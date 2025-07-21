package ports

import (
	"context"
	"time"
)

// Repository is the base interface for all repositories
type Repository interface {
	// Transaction support
	WithTx(tx Transaction) Repository
}

// UserRepository defines operations for user management
type UserRepository interface {
	Repository
	
	// CRUD operations
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	
	// Queries
	List(ctx context.Context, filter UserFilter, pagination Pagination) ([]*User, int64, error)
	Exists(ctx context.Context, email, username string) (bool, error)
	
	// Authentication
	UpdateLastLogin(ctx context.Context, id string, timestamp time.Time) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
}

// CharacterRepository defines operations for character management
type CharacterRepository interface {
	Repository
	
	// CRUD operations
	Create(ctx context.Context, character *Character) error
	GetByID(ctx context.Context, id string) (*Character, error)
	GetByName(ctx context.Context, name string) (*Character, error)
	GetByUserID(ctx context.Context, userID string) ([]*Character, error)
	Update(ctx context.Context, character *Character) error
	Delete(ctx context.Context, id string) error
	
	// Queries
	CountByUserID(ctx context.Context, userID string) (int, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	
	// Game-specific operations
	UpdatePosition(ctx context.Context, id string, position Position) error
	UpdateStats(ctx context.Context, id string, stats CharacterStats) error
	UpdatePlaytime(ctx context.Context, id string, additionalSeconds int64) error
}

// SessionRepository defines operations for session management
type SessionRepository interface {
	Repository
	
	// CRUD operations
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByToken(ctx context.Context, tokenHash string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	
	// Queries
	GetActiveByUserID(ctx context.Context, userID string) ([]*Session, error)
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByUserID(ctx context.Context, userID string) error
	
	// Session management
	UpdateLastActive(ctx context.Context, id string, timestamp time.Time) error
	ExtendExpiration(ctx context.Context, id string, newExpiration time.Time) error
}

// InventoryRepository defines operations for inventory management
type InventoryRepository interface {
	Repository
	
	// CRUD operations
	AddItem(ctx context.Context, characterID string, item *InventoryItem) error
	RemoveItem(ctx context.Context, characterID string, slotIndex int) error
	UpdateItem(ctx context.Context, characterID string, slotIndex int, item *InventoryItem) error
	GetItems(ctx context.Context, characterID string) ([]*InventoryItem, error)
	
	// Queries
	GetItem(ctx context.Context, characterID string, slotIndex int) (*InventoryItem, error)
	CountItems(ctx context.Context, characterID string) (int, error)
	HasSpace(ctx context.Context, characterID string, requiredSlots int) (bool, error)
	
	// Bulk operations
	MoveItem(ctx context.Context, characterID string, fromSlot, toSlot int) error
	SwapItems(ctx context.Context, characterID string, slot1, slot2 int) error
	Clear(ctx context.Context, characterID string) error
}

// Domain models

// User represents a user account
type User struct {
	ID               string
	Email            string
	Username         string
	PasswordHash     string
	AccountStatus    string
	EmailVerified    bool
	IsPremium        bool
	PremiumExpiresAt *time.Time
	MaxCharacters    int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLoginAt      *time.Time
	DeletedAt        *time.Time
}

// Character represents a game character
type Character struct {
	ID                     string
	UserID                 string
	Name                   string
	Class                  string
	Race                   string
	Gender                 string
	Level                  int
	Experience             int64
	LastZoneID             string
	LastPosition           Position
	Appearance             map[string]interface{}
	Stats                  CharacterStats
	Attributes             CharacterAttributes
	PlaytimeSeconds        int64
	CreatedAt              time.Time
	UpdatedAt              time.Time
	LastPlayedAt           time.Time
	DeletedAt              *time.Time
}

// Position represents a 3D position with rotation
type Position struct {
	X     float32
	Y     float32
	Z     float32
	Yaw   float32
	Pitch float32
	Roll  float32
}

// CharacterStats represents character combat stats
type CharacterStats struct {
	Health         int
	MaxHealth      int
	Mana           int
	MaxMana        int
	Stamina        int
	MaxStamina     int
	AttackPower    int
	SpellPower     int
	Defense        int
	MagicResistance int
}

// CharacterAttributes represents character base attributes
type CharacterAttributes struct {
	Strength     int
	Agility      int
	Intelligence int
	Wisdom       int
	Constitution int
	Charisma     int
	UnspentPoints int
}

// Session represents a user session
type Session struct {
	ID           string
	UserID       string
	TokenHash    string
	DeviceID     string
	IPAddress    string
	UserAgent    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	LastActiveAt time.Time
}

// InventoryItem represents an item in inventory
type InventoryItem struct {
	SlotIndex  int
	ItemID     string
	Quantity   int
	ItemData   map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Filter and pagination

// UserFilter defines filters for user queries
type UserFilter struct {
	AccountStatus *string
	IsPremium     *bool
	EmailVerified *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// Pagination defines pagination parameters
type Pagination struct {
	Offset int
	Limit  int
	Sort   string
	Order  string // ASC or DESC
}