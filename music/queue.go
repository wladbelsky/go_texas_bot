// Package music implements the Lavalink-backed music player ported from
// the original Python bot's music.py cog: a per-guild queue with
// repeat-modes, shuffle, and track-requester tracking, played through
// Lavalink v4 via disgolink.
package music

import (
	"encoding/json"
	"math/rand"
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

// RepeatMode mirrors the original bot's Queue.RepeatMode.
type RepeatMode string

const (
	RepeatNone RepeatMode = "none"
	RepeatOne  RepeatMode = "one"
	RepeatAll  RepeatMode = "all"
)

// WithRequester returns a copy of track with the requesting user's ID
// attached, so ownership (who's allowed to skip/stop it) survives in
// Lavalink's own track data instead of a side map that could drift out of
// sync with the queue.
func WithRequester(track lavalink.Track, requesterID snowflake.ID) lavalink.Track {
	withData, err := track.WithUserData(requesterID.String())
	if err != nil {
		// Marshaling a string can't realistically fail; fall back to the
		// track without requester data rather than dropping it entirely.
		return track
	}
	return withData
}

// Requester returns the user ID stashed by WithRequester, if any.
func Requester(track lavalink.Track) (string, bool) {
	if len(track.UserData) == 0 {
		return "", false
	}
	var requesterID string
	if err := json.Unmarshal(track.UserData, &requesterID); err != nil {
		return "", false
	}
	return requesterID, requesterID != ""
}

// Queue is a per-guild playback queue. It keeps the full track history (like
// the original bot's position-indexed queue) rather than consuming a plain
// FIFO, so a "previous track" control is possible.
type Queue struct {
	mu         sync.Mutex
	tracks     []lavalink.Track
	position   int // index of the current track; -1 before anything has played
	repeatMode RepeatMode
}

// NewQueue returns an empty queue.
func NewQueue() *Queue {
	return &Queue{position: -1, repeatMode: RepeatNone}
}

// IsEmpty reports whether the queue has no tracks at all.
func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks) == 0
}

// Length returns the total number of tracks in the queue.
func (q *Queue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks)
}

// Position returns the index of the current track (0-based), or -1 if
// nothing has started playing yet.
func (q *Queue) Position() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.position
}

// All returns a copy of every track in the queue, in order.
func (q *Queue) All() []lavalink.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]lavalink.Track(nil), q.tracks...)
}

// Current returns the track at the current position, if any.
func (q *Queue) Current() (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.currentLocked()
}

func (q *Queue) currentLocked() (lavalink.Track, bool) {
	if q.position < 0 || q.position >= len(q.tracks) {
		return lavalink.Track{}, false
	}
	return q.tracks[q.position], true
}

// Upcoming returns every track after the current position.
func (q *Queue) Upcoming() []lavalink.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	start := q.position + 1
	if start >= len(q.tracks) {
		return nil
	}
	return append([]lavalink.Track(nil), q.tracks[start:]...)
}

// History returns every track before the current position.
func (q *Queue) History() []lavalink.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	end := q.position
	if end < 0 {
		end = 0
	}
	if end > len(q.tracks) {
		end = len(q.tracks)
	}
	return append([]lavalink.Track(nil), q.tracks[:end]...)
}

// Add appends tracks to the end of the queue.
func (q *Queue) Add(tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

// RepeatMode returns the queue's current repeat mode.
func (q *Queue) RepeatMode() RepeatMode {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.repeatMode
}

// SetRepeatMode changes the queue's repeat mode.
func (q *Queue) SetRepeatMode(mode RepeatMode) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.repeatMode = mode
}

// Clear empties the queue and resets its position.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = nil
	q.position = -1
}

// Shuffle randomizes the order of the upcoming (not yet played) tracks.
func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()
	start := q.position + 1
	if start >= len(q.tracks) {
		return
	}
	upcoming := q.tracks[start:]
	rand.Shuffle(len(upcoming), func(i, j int) {
		upcoming[i], upcoming[j] = upcoming[j], upcoming[i]
	})
}

// Advance moves to the next track, honoring the repeat mode, and returns
// it. ok is false if there's nothing left to play (queue empty, or past the
// end with repeat off).
func (q *Queue) Advance() (track lavalink.Track, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) == 0 {
		return lavalink.Track{}, false
	}

	if q.repeatMode == RepeatOne && q.position >= 0 {
		return q.currentLocked()
	}

	q.position++
	if q.position >= len(q.tracks) {
		if q.repeatMode == RepeatAll {
			q.position = 0
		} else {
			// Stay on the last played track rather than parking past the end:
			// tracks appended later land at position+1 and get picked up by
			// the next Advance, and Previous still steps back correctly.
			q.position--
			return lavalink.Track{}, false
		}
	}
	return q.currentLocked()
}

// Previous moves back one track and returns it, mirroring the original
// bot's "prev" button (position -= 2 then advance).
func (q *Queue) Previous() (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) == 0 || q.position <= 0 {
		return lavalink.Track{}, false
	}
	q.position--
	return q.currentLocked()
}

// At returns the track at the given 0-based index without moving the
// position.
func (q *Queue) At(index int) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 || index >= len(q.tracks) {
		return lavalink.Track{}, false
	}
	return q.tracks[index], true
}

// SkipTo jumps directly to the 0-based index and returns that track.
func (q *Queue) SkipTo(index int) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if index < 0 || index >= len(q.tracks) {
		return lavalink.Track{}, false
	}
	q.position = index
	return q.currentLocked()
}

// RemoveAt removes the track at the given 0-based index. It cannot remove
// the currently playing track (index == position), matching the original
// bot's behavior of only letting you pop queued-but-not-playing tracks.
func (q *Queue) RemoveAt(index int) (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if index < 0 || index >= len(q.tracks) || index == q.position {
		return lavalink.Track{}, false
	}
	track := q.tracks[index]
	q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)
	if index < q.position {
		q.position--
	}
	return track, true
}

// Manager owns one Queue per guild.
type Manager struct {
	mu     sync.Mutex
	queues map[snowflake.ID]*Queue
}

// NewManager returns an empty Manager.
func NewManager() *Manager {
	return &Manager{queues: make(map[snowflake.ID]*Queue)}
}

// Get returns the guild's queue, creating an empty one if it doesn't exist
// yet.
func (m *Manager) Get(guildID snowflake.ID) *Queue {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, ok := m.queues[guildID]
	if !ok {
		q = NewQueue()
		m.queues[guildID] = q
	}
	return q
}

// Delete removes the guild's queue entirely, e.g. once the bot leaves its
// voice channel.
func (m *Manager) Delete(guildID snowflake.ID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.queues, guildID)
}
