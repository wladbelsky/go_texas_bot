package arknights

import "testing"

func TestRollRarity_AlwaysWithinRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		rarity, _ := RollRarity(0)
		if rarity < 3 || rarity > 6 {
			t.Fatalf("RollRarity(0) = %d, want a value in [3,6]", rarity)
		}
	}
}

func TestRollRarity_ResetsPityOnlyOnSixStar(t *testing.T) {
	for i := 0; i < 1000; i++ {
		rarity, resetPity := RollRarity(0)
		if resetPity && rarity != 6 {
			t.Fatalf("resetPity = true but rarity = %d, want resetPity only on a six-star", rarity)
		}
		if rarity == 6 && !resetPity {
			t.Fatalf("rolled a six-star but resetPity = false")
		}
	}
}

func TestRollRarity_HighPityGuaranteesSixStar(t *testing.T) {
	// Once the bonus alone reaches the full roll range, every roll must be
	// a six-star, whatever the random draw is.
	const sixMiss = 100
	for i := 0; i < 100; i++ {
		rarity, resetPity := RollRarity(sixMiss)
		if rarity != 6 || !resetPity {
			t.Fatalf("RollRarity(%d) = (%d, %v), want a guaranteed six-star", sixMiss, rarity, resetPity)
		}
	}
}

func TestRandomOfRarity_ReturnsCharacterOfThatRarity(t *testing.T) {
	for rarity := 3; rarity <= 6; rarity++ {
		c, ok := RandomOfRarity(rarity)
		if !ok {
			t.Fatalf("RandomOfRarity(%d) reported not found", rarity)
		}
		if c.Rarity != rarity {
			t.Fatalf("RandomOfRarity(%d) returned a character of rarity %d", rarity, c.Rarity)
		}
	}
}

func TestRandomOfRarity_UnknownRarity(t *testing.T) {
	if _, ok := RandomOfRarity(1); ok {
		t.Fatal("expected RandomOfRarity(1) to report not found (1-2 star operators aren't collectible)")
	}
}
