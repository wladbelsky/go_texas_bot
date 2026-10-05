package reactions

import "testing"

func TestPhrase_Format_WithTarget(t *testing.T) {
	p := Phrases["hug"]
	got := p.Format("Alice", "Bob")
	want := "Alice обнимает Bob"
	if got != want {
		t.Fatalf("Format(with target) = %q, want %q", got, want)
	}
}

func TestPhrase_Format_SelfOnly(t *testing.T) {
	p := Phrases["hug"]
	got := p.Format("Alice", "")
	want := "Alice обнимает сам себя"
	if got != want {
		t.Fatalf("Format(no target) = %q, want %q", got, want)
	}
}

func TestCategories_MatchPhrasesMap(t *testing.T) {
	if len(Categories) != 25 {
		t.Fatalf("len(Categories) = %d, want 25 (Discord's max choices per option)", len(Categories))
	}
	for _, c := range Categories {
		if _, ok := Phrases[c]; !ok {
			t.Errorf("category %q has no entry in Phrases", c)
		}
	}
	if len(Phrases) != len(Categories) {
		t.Fatalf("len(Phrases) = %d, len(Categories) = %d, want them equal", len(Phrases), len(Categories))
	}
}
