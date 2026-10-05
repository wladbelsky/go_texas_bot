// Package cooldown implements the per-guild-per-user command cooldowns
// ported from the original bot's user_guild_cooldown. Unlike the original
// (discord.py keeps them in memory), they're stored in the database so a
// restart or redeploy doesn't reset them.
package cooldown

import (
	"errors"
	"log/slog"
	"time"

	"go_texas_bot/db"
	"gorm.io/gorm"
)

// Tracker is one command's cooldown. Trackers for different commands share
// the db.Cooldown table and are told apart by name.
type Tracker struct {
	name     string
	duration time.Duration
}

// New returns a cooldown of the given duration for the named command. The
// name is part of the stored key, so it must be unique per command and
// shouldn't change, or running cooldowns are lost.
func New(name string, duration time.Duration) *Tracker {
	return &Tracker{name: name, duration: duration}
}

// Reserve starts the cooldown for the user in the guild and returns 0, or,
// if one is already running, returns how much of it is left. The check and
// the reservation are a single SQL statement, so two concurrent invocations
// can't both get through. Callers that then fail to do anything should
// Release.
func (t *Tracker) Reserve(guildID, userID string) (time.Duration, error) {
	remaining, err := t.tryReserve(guildID, userID)
	if errors.Is(err, errGone) {
		// The running cooldown disappeared between the two statements in
		// tryReserve (released, or it just expired), so the slot is free
		// now: try to take it once more.
		remaining, err = t.tryReserve(guildID, userID)
	}
	return remaining, err
}

// errGone means the running cooldown vanished while being read.
var errGone = errors.New("cooldown: running cooldown vanished while reserving")

func (t *Tracker) tryReserve(guildID, userID string) (time.Duration, error) {
	now := time.Now()
	nowMs := now.UnixMilli()
	expiresAt := now.Add(t.duration).UnixMilli()

	// Insert the cooldown, or take over an existing row only if it has
	// already expired. RowsAffected is 0 exactly when a cooldown is running.
	res := db.DB.Exec(`INSERT INTO cooldowns (guild_id, user_id, command, expires_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (guild_id, user_id, command) DO UPDATE SET expires_at = excluded.expires_at
		WHERE cooldowns.expires_at <= ?`,
		guildID, userID, t.name, expiresAt, nowMs)
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected > 0 {
		return 0, nil
	}

	var running db.Cooldown
	err := db.DB.Where(&db.Cooldown{GuildID: guildID, UserID: userID, Command: t.name}).First(&running).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errGone
	}
	if err != nil {
		return 0, err
	}
	remaining := time.Duration(running.ExpiresAt-nowMs) * time.Millisecond
	if remaining <= 0 {
		return 0, errGone
	}
	return remaining, nil
}

// Release cancels a reservation, so a command that failed doesn't cost the
// user a full cooldown. It's a best-effort refund: a failure is only logged.
func (t *Tracker) Release(guildID, userID string) {
	err := db.DB.Where(&db.Cooldown{GuildID: guildID, UserID: userID, Command: t.name}).Delete(&db.Cooldown{}).Error
	if err != nil {
		slog.Error("cooldown: failed to release", "command", t.name, "guild", guildID, "user", userID, "err", err)
	}
}
