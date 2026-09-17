package ark

import (
	"sync"
	"time"
)

// cooldown mirrors the original bot's per-guild-per-user 4h cooldown on
// /ark. It's kept in memory (reset on restart) rather than persisted, since
// losing a partial cooldown on redeploy is a fine trade-off for a personal
// bot.
const cooldown = 4 * time.Hour

var lastRoll sync.Map // map[string]time.Time, keyed by "guildID:userID"

func checkCooldown(guildID, userID string) time.Duration {
	key := guildID + ":" + userID
	now := time.Now()
	if v, ok := lastRoll.Load(key); ok {
		if remaining := cooldown - now.Sub(v.(time.Time)); remaining > 0 {
			return remaining
		}
	}
	lastRoll.Store(key, now)
	return 0
}
