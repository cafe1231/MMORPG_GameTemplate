package character

import (
	"time"

	"github.com/google/uuid"
)

// Stats represents the character's combat and gameplay statistics
type Stats struct {
	ID                   uuid.UUID
	CharacterID          uuid.UUID
	// Primary stats
	Strength             int
	Dexterity            int
	Intelligence         int
	Wisdom               int
	Constitution         int
	Charisma             int
	// Combat stats
	HealthCurrent        int
	HealthMax            int
	ManaCurrent          int
	ManaMax              int
	StaminaCurrent       int
	StaminaMax           int
	// Derived stats
	AttackPower          int
	SpellPower           int
	Defense              int
	CriticalChance       float32
	CriticalDamage       float32
	DodgeChance          float32
	BlockChance          float32
	// Movement and other stats
	MovementSpeed        float32
	AttackSpeed          float32
	CastSpeed            float32
	// Resource regeneration
	HealthRegen          float32
	ManaRegen            float32
	StaminaRegen         float32
	// Points for allocation
	StatPointsAvailable  int
	SkillPointsAvailable int
	// Metadata
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewStats creates new character stats with class-specific defaults
func NewStats(characterID uuid.UUID, classType ClassType) *Stats {
	now := time.Now()
	stats := &Stats{
		ID:                   uuid.New(),
		CharacterID:          characterID,
		Strength:             10,
		Dexterity:            10,
		Intelligence:         10,
		Wisdom:               10,
		Constitution:         10,
		Charisma:             10,
		HealthCurrent:        100,
		HealthMax:            100,
		ManaCurrent:          50,
		ManaMax:              50,
		StaminaCurrent:       100,
		StaminaMax:           100,
		AttackPower:          0,
		SpellPower:           0,
		Defense:              0,
		CriticalChance:       5.0,
		CriticalDamage:       150.0,
		DodgeChance:          5.0,
		BlockChance:          0.0,
		MovementSpeed:        100.0,
		AttackSpeed:          100.0,
		CastSpeed:            100.0,
		HealthRegen:          1.0,
		ManaRegen:            1.0,
		StaminaRegen:         5.0,
		StatPointsAvailable:  0,
		SkillPointsAvailable: 0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Apply class-specific starting stats
	stats.ApplyClassDefaults(classType)
	stats.CalculateDerivedStats(classType)
	
	return stats
}

// ApplyClassDefaults sets class-specific starting primary stats
func (s *Stats) ApplyClassDefaults(classType ClassType) {
	switch classType {
	case ClassWarrior:
		s.Strength = 15
		s.Constitution = 13
		s.Dexterity = 8
		s.Intelligence = 6
		s.Wisdom = 6
		s.Charisma = 8
	case ClassMage:
		s.Strength = 6
		s.Constitution = 8
		s.Dexterity = 8
		s.Intelligence = 15
		s.Wisdom = 10
		s.Charisma = 8
	case ClassRogue:
		s.Strength = 8
		s.Constitution = 8
		s.Dexterity = 15
		s.Intelligence = 8
		s.Wisdom = 8
		s.Charisma = 10
	case ClassPriest:
		s.Strength = 6
		s.Constitution = 8
		s.Dexterity = 8
		s.Intelligence = 10
		s.Wisdom = 15
		s.Charisma = 10
	case ClassRanger:
		s.Strength = 10
		s.Constitution = 10
		s.Dexterity = 13
		s.Intelligence = 8
		s.Wisdom = 10
		s.Charisma = 8
	case ClassPaladin:
		s.Strength = 13
		s.Constitution = 12
		s.Dexterity = 8
		s.Intelligence = 8
		s.Wisdom = 10
		s.Charisma = 10
	case ClassWarlock:
		s.Strength = 6
		s.Constitution = 8
		s.Dexterity = 8
		s.Intelligence = 13
		s.Wisdom = 8
		s.Charisma = 13
	case ClassDruid:
		s.Strength = 8
		s.Constitution = 10
		s.Dexterity = 8
		s.Intelligence = 10
		s.Wisdom = 13
		s.Charisma = 8
	}
}

// CalculateDerivedStats recalculates all derived stats based on primary stats and class
func (s *Stats) CalculateDerivedStats(classType ClassType) {
	// Health calculation
	switch classType {
	case ClassWarrior, ClassPaladin:
		s.HealthMax = 100 + (s.Constitution * 10) + (s.Strength * 2)
	case ClassRogue, ClassRanger:
		s.HealthMax = 100 + (s.Constitution * 8) + (s.Dexterity * 2)
	case ClassMage, ClassWarlock:
		s.HealthMax = 100 + (s.Constitution * 6) + (s.Intelligence * 1)
	case ClassPriest, ClassDruid:
		s.HealthMax = 100 + (s.Constitution * 7) + (s.Wisdom * 2)
	default:
		s.HealthMax = 100 + (s.Constitution * 8)
	}

	// Mana calculation
	switch classType {
	case ClassMage, ClassWarlock:
		s.ManaMax = 50 + (s.Intelligence * 10) + (s.Wisdom * 2)
	case ClassPriest, ClassDruid:
		s.ManaMax = 50 + (s.Wisdom * 10) + (s.Intelligence * 2)
	case ClassPaladin:
		s.ManaMax = 50 + (s.Wisdom * 5)
	default:
		s.ManaMax = 50 + (s.Intelligence * 2)
	}

	// Stamina calculation
	s.StaminaMax = 100 + (s.Constitution * 5) + (s.Strength * 2)

	// Attack power calculation
	switch classType {
	case ClassWarrior, ClassPaladin:
		s.AttackPower = s.Strength*2 + s.Dexterity
	case ClassRogue, ClassRanger:
		s.AttackPower = s.Dexterity*2 + s.Strength
	default:
		s.AttackPower = s.Strength + s.Dexterity
	}

	// Spell power calculation
	switch classType {
	case ClassMage, ClassWarlock:
		s.SpellPower = s.Intelligence * 3
	case ClassPriest, ClassDruid:
		s.SpellPower = s.Wisdom * 3
	case ClassPaladin:
		s.SpellPower = (s.Wisdom*2 + s.Intelligence) / 2
	default:
		s.SpellPower = 0
	}

	// Defense calculation
	s.Defense = s.Constitution*2 + (s.Strength+s.Dexterity)/2

	// Critical and dodge chances
	s.CriticalChance = 5.0 + float32(s.Dexterity)*0.1
	s.DodgeChance = 5.0 + float32(s.Dexterity)*0.2

	// Regeneration rates
	s.HealthRegen = 1.0 + float32(s.Constitution)*0.1
	switch classType {
	case ClassMage, ClassWarlock, ClassPriest, ClassDruid:
		s.ManaRegen = 1.0 + float32(s.Wisdom)*0.2
	default:
		s.ManaRegen = 1.0 + float32(s.Wisdom)*0.05
	}
	s.StaminaRegen = 5.0 + float32(s.Constitution)*0.2

	// Ensure current values don't exceed maximums
	if s.HealthCurrent > s.HealthMax {
		s.HealthCurrent = s.HealthMax
	}
	if s.ManaCurrent > s.ManaMax {
		s.ManaCurrent = s.ManaMax
	}
	if s.StaminaCurrent > s.StaminaMax {
		s.StaminaCurrent = s.StaminaMax
	}
}

// AllocateStatPoint allocates a stat point to a primary stat
func (s *Stats) AllocateStatPoint(stat string) error {
	if s.StatPointsAvailable <= 0 {
		return ErrNoStatPointsAvailable
	}

	switch stat {
	case "strength":
		s.Strength++
	case "dexterity":
		s.Dexterity++
	case "intelligence":
		s.Intelligence++
	case "wisdom":
		s.Wisdom++
	case "constitution":
		s.Constitution++
	case "charisma":
		s.Charisma++
	default:
		return ErrInvalidStatType
	}

	s.StatPointsAvailable--
	s.UpdatedAt = time.Now()
	return nil
}

// AddStatPoints adds stat points (usually from leveling up)
func (s *Stats) AddStatPoints(points int) {
	s.StatPointsAvailable += points
	s.UpdatedAt = time.Now()
}

// AddSkillPoints adds skill points (usually from leveling up)
func (s *Stats) AddSkillPoints(points int) {
	s.SkillPointsAvailable += points
	s.UpdatedAt = time.Now()
}

// TakeDamage reduces health by the specified amount
func (s *Stats) TakeDamage(damage int) {
	s.HealthCurrent -= damage
	if s.HealthCurrent < 0 {
		s.HealthCurrent = 0
	}
	s.UpdatedAt = time.Now()
}

// Heal increases health by the specified amount
func (s *Stats) Heal(amount int) {
	s.HealthCurrent += amount
	if s.HealthCurrent > s.HealthMax {
		s.HealthCurrent = s.HealthMax
	}
	s.UpdatedAt = time.Now()
}

// UseMana reduces mana by the specified amount
func (s *Stats) UseMana(amount int) error {
	if s.ManaCurrent < amount {
		return ErrInsufficientMana
	}
	s.ManaCurrent -= amount
	s.UpdatedAt = time.Now()
	return nil
}

// RestoreMana increases mana by the specified amount
func (s *Stats) RestoreMana(amount int) {
	s.ManaCurrent += amount
	if s.ManaCurrent > s.ManaMax {
		s.ManaCurrent = s.ManaMax
	}
	s.UpdatedAt = time.Now()
}

// UseStamina reduces stamina by the specified amount
func (s *Stats) UseStamina(amount int) error {
	if s.StaminaCurrent < amount {
		return ErrInsufficientStamina
	}
	s.StaminaCurrent -= amount
	s.UpdatedAt = time.Now()
	return nil
}

// RestoreStamina increases stamina by the specified amount
func (s *Stats) RestoreStamina(amount int) {
	s.StaminaCurrent += amount
	if s.StaminaCurrent > s.StaminaMax {
		s.StaminaCurrent = s.StaminaMax
	}
	s.UpdatedAt = time.Now()
}

// IsDead checks if the character is dead
func (s *Stats) IsDead() bool {
	return s.HealthCurrent <= 0
}

// FullRestore restores all resources to maximum
func (s *Stats) FullRestore() {
	s.HealthCurrent = s.HealthMax
	s.ManaCurrent = s.ManaMax
	s.StaminaCurrent = s.StaminaMax
	s.UpdatedAt = time.Now()
}