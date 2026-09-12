package arknights

import (
	"errors"
	"fmt"

	"go_texas_bot/db"
	"gorm.io/gorm"
)

// Roll performs one gacha roll for the given user: rolls a rarity (applying
// the user's pity counter), picks a random character of that rarity, adds
// it to the user's collection, and persists the updated pity counter.
func Roll(userID string) (Character, error) {
	pity, err := getPity(userID)
	if err != nil {
		return Character{}, err
	}

	rarity, resetPity := RollRarity(pity.SixMiss)
	character, ok := RandomOfRarity(rarity)
	if !ok {
		return Character{}, fmt.Errorf("no characters available for rarity %d", rarity)
	}

	if resetPity {
		pity.SixMiss = 0
	} else {
		pity.SixMiss++
	}
	if err = db.DB.Save(&pity).Error; err != nil {
		return Character{}, err
	}

	if err = addToCollection(userID, character); err != nil {
		return Character{}, err
	}

	return character, nil
}

func getPity(userID string) (db.ArkPity, error) {
	var pity db.ArkPity
	err := db.DB.FirstOrCreate(&pity, db.ArkPity{UserID: userID}).Error
	return pity, err
}

func addToCollection(userID string, character Character) error {
	var entry db.ArkCollectionEntry
	err := db.DB.FirstOrCreate(&entry, db.ArkCollectionEntry{
		UserID:       userID,
		OperatorName: character.DisplayName(),
	}).Error
	if err != nil {
		return err
	}
	entry.Rarity = character.Rarity
	entry.Count++
	return db.DB.Save(&entry).Error
}

// Collection returns the calling user's collected operators grouped by
// rarity.
func Collection(userID string) (map[int][]db.ArkCollectionEntry, error) {
	var entries []db.ArkCollectionEntry
	if err := db.DB.Where("user_id = ?", userID).Order("rarity").Find(&entries).Error; err != nil {
		return nil, err
	}
	out := make(map[int][]db.ArkCollectionEntry)
	for _, e := range entries {
		out[e.Rarity] = append(out[e.Rarity], e)
	}
	return out, nil
}

// OwnedCount returns how many distinct operators a user has collected.
func OwnedCount(userID string) (int, error) {
	var count int64
	err := db.DB.Model(&db.ArkCollectionEntry{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

// OwnedEntry returns the user's collection entry for the given character,
// if they own at least one copy of it.
func OwnedEntry(userID string, character Character) (db.ArkCollectionEntry, bool, error) {
	var entry db.ArkCollectionEntry
	err := db.DB.Where(&db.ArkCollectionEntry{
		UserID:       userID,
		OperatorName: character.DisplayName(),
	}).First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.ArkCollectionEntry{}, false, nil
	}
	if err != nil {
		return db.ArkCollectionEntry{}, false, err
	}
	return entry, true, nil
}

// exchangeRate is how many duplicate copies convert into one roll of the
// next rarity tier.
const exchangeRate = 5

// Barter converts every complete group of duplicate copies of a sub-six-star
// operator into a roll of the next rarity tier and adds the results to the
// user's collection. It mirrors the original bot's exchange formula
// verbatim, including its quirk of using the post-exchange remainder (not
// the number of exchanged groups) as the granted roll count.
func Barter(userID string) ([]Character, error) {
	var entries []db.ArkCollectionEntry
	err := db.DB.Where(&db.ArkCollectionEntry{UserID: userID}).
		Where("count > ? AND rarity < ?", exchangeRate, 6).
		Order("rarity").Find(&entries).Error
	if err != nil {
		return nil, err
	}

	type exchange struct {
		rarity int
		count  int
	}
	exchanges := make([]exchange, 0, len(entries))
	for i := range entries {
		e := &entries[i]

		var remainder int
		if e.Count%exchangeRate != 0 {
			remainder = e.Count % exchangeRate
		} else {
			remainder = e.Count/exchangeRate - 1
		}
		exchanges = append(exchanges, exchange{rarity: e.Rarity + 1, count: remainder})

		e.Count = remainder
		if err = db.DB.Save(e).Error; err != nil {
			return nil, err
		}
	}

	var results []Character
	for _, ex := range exchanges {
		for i := 0; i < ex.count; i++ {
			character, ok := RandomOfRarity(ex.rarity)
			if !ok {
				continue
			}
			if err = addToCollection(userID, character); err != nil {
				return nil, err
			}
			results = append(results, character)
		}
	}
	return results, nil
}
