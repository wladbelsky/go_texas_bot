package arknights

import "math/rand"

// Default gacha rates, ported from the original bot's config.json
// (default_settings.ark.chance). Values are in permille (chance/1000 of
// the roll range) so a roll of the six-star bucket needs
// roll <= sixStarChance*1000 out of a 0-100000 range.
const (
	sixStarChance   = 2
	fiveStarChance  = 10
	fourStarChance  = 60
	threeStarChance = 100
)

// RollRarity picks a rarity (3-6) using the same cumulative-threshold gacha
// as the original bot. sixMiss is the number of consecutive rolls without a
// six-star; past 50 misses it adds an increasing soft-pity bonus to the
// six-star chance, exactly like the Python implementation. resetPity
// reports whether the caller should reset the pity counter (a six-star was
// rolled).
func RollRarity(sixMiss int) (rarity int, resetPity bool) {
	bonus := 0
	if sixMiss >= 50 {
		bonus = sixMiss*2 - 50
	}

	roll := rand.Intn(100000)
	switch {
	case roll <= (sixStarChance+bonus)*1000:
		return 6, true
	case roll <= fiveStarChance*1000:
		return 5, false
	case roll <= fourStarChance*1000:
		return 4, false
	case roll <= threeStarChance*1000:
		return 3, false
	default:
		// Unreachable as long as threeStarChance*1000 covers the full
		// 0-100000 roll range, same as in the original bot's config.
		return 3, false
	}
}

// RandomOfRarity returns a random character of the given star rarity.
func RandomOfRarity(rarity int) (Character, bool) {
	pool := ByRarity(rarity)
	if len(pool) == 0 {
		return Character{}, false
	}
	return pool[rand.Intn(len(pool))], true
}
