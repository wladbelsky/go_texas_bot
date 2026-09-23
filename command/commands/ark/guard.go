package ark

import (
	"time"

	"go_texas_bot/command/cooldown"
)

// rollCooldown mirrors the original bot's per-guild-per-user /ark cooldown
// (config.json ark.ark_cooldown=14400s).
var rollCooldown = cooldown.New("ark", 4*time.Hour)
