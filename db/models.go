package db

// ArkCollectionEntry tracks how many copies of an operator a user has rolled.
type ArkCollectionEntry struct {
	UserID       string `gorm:"primaryKey;index:idx_ark_collection_user"`
	OperatorName string `gorm:"primaryKey"`
	Rarity       int
	Count        int
}

// ArkStats is a singleton row (ID is always 1) of global /ark counters.
// TotalGranted only ever grows, unlike summing ArkCollectionEntry.Count,
// which /barter reduces when it exchanges duplicates away.
type ArkStats struct {
	ID           uint `gorm:"primaryKey"`
	TotalGranted int  // operators ever added to any collection, by roll or barter
}

// ArkPity tracks the six-star soft-pity counter per user.
type ArkPity struct {
	UserID  string `gorm:"primaryKey"`
	SixMiss int
}

// GuildSettings holds per-guild feature toggles for optional/sensitive content.
type GuildSettings struct {
	GuildID               string `gorm:"primaryKey"`
	NSFWEnabled           bool
	ToxicGreetingsEnabled bool
}

// GerStats is a singleton row of global /ger counters (ID is always 1).
type GerStats struct {
	ID    uint `gorm:"primaryKey"`
	Total int  // every /ger call, including self-directed ones
	Self  int  // calls that landed on the caller themselves
	Bot   int  // calls whose target was a bot
	Me    int  // calls whose target was this bot
}

// UserGerCounter tracks each user's /ger involvement: how many times they
// called it (Uses) and how many times they were the target (Hits).
type UserGerCounter struct {
	UserID string `gorm:"primaryKey"`
	Uses   int
	Hits   int
}

// Cooldown is a running per-guild-per-user command cooldown. Rows are keyed
// by command so every command.cooldown.Tracker shares this one table.
// ExpiresAt is unix milliseconds, so expiry can be compared in SQL without
// depending on how the driver serializes times.
type Cooldown struct {
	GuildID   string `gorm:"primaryKey"`
	UserID    string `gorm:"primaryKey"`
	Command   string `gorm:"primaryKey"`
	ExpiresAt int64
}
