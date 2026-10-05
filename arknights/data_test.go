package arknights

import "testing"

func TestAll_NotEmpty(t *testing.T) {
	if Count() == 0 {
		t.Fatal("expected embedded character data to be non-empty")
	}
}

func TestByRarity_CoversThreeToSix(t *testing.T) {
	for rarity := 3; rarity <= 6; rarity++ {
		if len(ByRarity(rarity)) == 0 {
			t.Errorf("expected at least one character of rarity %d", rarity)
		}
	}
}

func TestByRarity_RejectsOutOfRange(t *testing.T) {
	for _, rarity := range []int{0, 1, 2, 7} {
		if len(ByRarity(rarity)) != 0 {
			t.Errorf("ByRarity(%d) = non-empty, want empty (only 3-6 star are collectible)", rarity)
		}
	}
}

func TestByName_KnownCharacter(t *testing.T) {
	c, ok := ByName("Amiya")
	if !ok {
		t.Fatal("expected to find Amiya")
	}
	if c.DisplayName() != "Amiya" {
		t.Fatalf("DisplayName() = %q, want %q", c.DisplayName(), "Amiya")
	}
}

func TestByName_NormalizesSpacesAndApostrophes(t *testing.T) {
	for _, c := range All() {
		if c.Name == c.DisplayName() {
			continue
		}
		got, ok := ByName(c.Name)
		if !ok {
			t.Fatalf("ByName(%q) not found", c.Name)
		}
		if got.DisplayName() != c.DisplayName() {
			t.Fatalf("ByName(%q) = %q, want %q", c.Name, got.DisplayName(), c.DisplayName())
		}
		return
	}
	t.Skip("no character with a space/apostrophe in its name to test normalization against")
}

func TestByName_Missing(t *testing.T) {
	if _, ok := ByName("Definitely Not A Real Operator"); ok {
		t.Fatal("expected ByName to report a missing character")
	}
}

func TestCharacter_CleanTraits_StripsMarkup(t *testing.T) {
	c := Character{Traits: "Deals <@ba.kw>Arts damage</> to a single target"}
	want := "Deals Arts damage to a single target"
	if got := c.CleanTraits(); got != want {
		t.Fatalf("CleanTraits() = %q, want %q", got, want)
	}
}

func TestSkin_DisplayName_PrefersSkinName(t *testing.T) {
	s := Skin{SkinName: "Winter Solstice", ModelName: "Amiya", PortraitID: "char_002_amiya_2"}
	if got := s.DisplayName(); got != "Winter Solstice" {
		t.Fatalf("DisplayName() = %q, want %q", got, "Winter Solstice")
	}
}

func TestSkin_DisplayName_FallsBackToModelAndSuffix(t *testing.T) {
	s := Skin{ModelName: "Amiya", PortraitID: "char_002_amiya_2"}
	if got := s.DisplayName(); got != "Amiya#2" {
		t.Fatalf("DisplayName() = %q, want %q", got, "Amiya#2")
	}
}

func TestSkinsFor_KnownCharacterHasDefaultSkin(t *testing.T) {
	c, ok := ByName("Amiya")
	if !ok {
		t.Fatal("expected to find Amiya")
	}
	if len(SkinsFor(c.ID)) == 0 {
		t.Fatalf("expected at least one skin for %s", c.ID)
	}
}
