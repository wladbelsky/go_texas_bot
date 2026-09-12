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
