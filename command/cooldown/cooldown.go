// Package cooldown implements the per-guild-per-user command cooldowns
// ported from the original bot's user_guild_cooldown. They're kept in
// memory (reset on restart) rather than persisted, since losing a partial
// cooldown on redeploy is a fine trade-off for a personal bot.
package cooldown

import (
	"sync"
	"time"
)

type Tracker struct {
	duration time.Duration
	mu       sync.Mutex
	last     map[string]time.Time // keyed by "guildID:userID"
}

func New(duration time.Duration) *Tracker {
	return &Tracker{duration: duration, last: make(map[string]time.Time)}
}

// Reserve starts the cooldown for the user in the guild and returns 0, or,
// if one is already running, returns how much of it is left. The check and
// the reservation happen under one lock so two concurrent invocations can't
// both get through. Callers that then fail to do anything should Release.
func (t *Tracker) Reserve(guildID, userID string) time.Duration {
	key := guildID + ":" + userID
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()
	if last, ok := t.last[key]; ok {
		if remaining := t.duration - now.Sub(last); remaining > 0 {
			return remaining
		}
	}
	t.last[key] = now
	return 0
}

// Release cancels a reservation, so a command that failed doesn't cost the
// user a full cooldown.
func (t *Tracker) Release(guildID, userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, guildID+":"+userID)
}
