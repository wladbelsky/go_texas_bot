package db

// ArkCollectionEntry tracks how many copies of an operator a user has rolled.
type ArkCollectionEntry struct {
	UserID       string `gorm:"primaryKey;index:idx_ark_collection_user"`
	OperatorName string `gorm:"primaryKey"`
	Rarity       int
	Count        int
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
