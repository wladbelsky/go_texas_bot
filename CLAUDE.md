# CLAUDE.md

This repo is a Discord bot ("Texass") written in Go with disgo v0.19, disgolink v3 (Lavalink v4) and GORM + SQLite (`github.com/glebarez/sqlite`, pure Go, no CGO). It is a port of the Python bot [reven-n1/DiscordBot](https://github.com/reven-n1/DiscordBot). See README.md for what the bot does and how to run it.

## Commands

```sh
GOTOOLCHAIN=go1.26.0+auto go build ./...
GOTOOLCHAIN=go1.26.0+auto go vet ./...
GOTOOLCHAIN=go1.26.0+auto go test ./...
gofmt -l .   # must print nothing
```

CI (`.github/workflows/build.yml`) runs `go test ./...` and then builds the amd64/arm64 binaries and the GHCR image.

## Layout

| Path | What lives there |
|---|---|
| `main.go` | Wires everything: `db.Init`, the disgo client (intents, member + voice-state cache), event listeners, `music.Connect` (in a background goroutine), global command registration from `registry.Commands`. |
| `config/` | Flag/env lookup (`BOT_TOKEN`, `LAVALINK_*`, `DB_PATH`). An env var overrides its flag. |
| `db/` | `db.DB` global, models, `AutoMigrate`. |
| `arknights/` | Embedded operator and skin data, gacha rarity roll with soft pity, collection and barter logic. |
| `music/` | Per-guild `Queue` (history, repeat modes, requester stored in `Track.UserData`), the disgolink client and track-end auto-advance. Get the client with `music.Client()`: it is nil until Lavalink connects, so always nil-check it and treat nil as "music unavailable". Playback failures (`TrackException`/`TrackStuck`) are posted through `music.SetNotifier` to the channel `/play` was last used in, and `Queue.SkipFailed` moves past failed tracks without looping. |
| `reactions/` | waifu.pics client and the `/reaction` phrase table. |
| `guildsettings/` | Per-guild opt-in flags; `RequireGuildNSFW` guard. |
| `registry/` | `Commands`: every slash/user command definition sent to Discord. |
| `command/` | `Listener` dispatches interactions by `(name, type)` key; `sendError`; `register.go` blank-imports every command package. |
| `command/command_selector/` | `Key(name, cmdType)`, the handler map key. |
| `command/cooldown/` | Per-guild-per-user cooldown `Tracker`, stored in the `db.Cooldown` table (keyed by command name) so cooldowns survive restarts. |
| `command/embedutil/` | `ChunkLines` for splitting long text across embed fields. |
| `command/commands/*` | One package per command group. `default/` is package `defaultcmd` (`default` is a keyword). `welcome/` holds gateway member join/leave listeners, not commands. |

## Adding a command

1. Add its definition to `registry.Commands` in `registry/registry.go`. User context-menu commands are `discord.UserCommandCreate`.
2. In a package under `command/commands/`, register the handler from `init()` under `command_selector.Key(name, discord.ApplicationCommandTypeSlash)` (or `...TypeUser`). Follow the existing packages.
3. If the package is new, blank-import it in `command/register.go`.

`/help` is generated from `registry.Commands`, so it picks the command up automatically.

**Import rule:** command packages import `command`, so shared code must live in leaf packages (`registry`, `guildsettings`, `command/cooldown`, `command/embedutil`, domain packages). These must never import `command`, or you get an import cycle.

## disgo / Discord gotchas

- `Permissions.Has(a, b)` means *all* bits. For "any of", use `p.Has(a) || p.Has(b)`.
- `SlashCommandOption.String()` panics if the option isn't a string. Use `Int`/`OptInt` etc. to match the registered type.
- After `DeferCreateMessage`, you can't call `CreateMessage` again. Respond with `Rest.UpdateInteractionResponse` (see `sendError` in `command/commands.go`).
- Embed fields are limited to 1024 characters, and embeds to 25 fields; option choices max out at 25. Split long lists with `embedutil.ChunkLines`.
- `UserCommandInteractionData().TargetMember()` is zero-valued in DMs; fall back to `TargetUser()`.
- Messages use the immutable builder: `discord.NewMessageCreate().WithContent(...).WithEphemeral(true)`.
- `client.Rest` and `client.ApplicationID` are fields, not methods.
- Component handlers must check `event.Data.Type()` before `ButtonInteractionData()` (it's a bare type assertion).

## State and data

- Every new model must be added to `AutoMigrate` in `db/db.go`.
- Global counters are singleton rows with `ID: 1`, loaded with `FirstOrCreate` (`GerStats`, `ArkStats`).
- `ArkCollectionEntry` is keyed by `Character.DisplayName()`. Renaming an operator in the data orphans existing collection rows.
- Cooldowns: `Reserve` at the start of the command (a single upsert, so concurrent double use is blocked; it returns an error on DB failure), and `Release` on every error/refusal path so a failed command doesn't burn the cooldown. The `cooldown.New` name is part of the stored key: keep it unique and stable.
- `arknights/data/characters.json` and `skins.json` are trimmed copies of the en_US `character_table` / `skin_table` from [Aceship/AN-EN-Tags](https://github.com/Aceship/AN-EN-Tags). Only obtainable 3–6★ operators are included, with the fields of `arknights.Character` / `arknights.Skin`. There is no generator script; when refreshing, keep the same shape and run the `arknights` tests.

## Content and behavior policy

- NSFW and hostile content (`/nsfw`, the Arknights commands, `/ger`, the leave message) must stay behind `guildsettings` flags. These are off by default and set with `/settings`.
- User-facing strings are Russian, matching the original bot. Code, comments and docs are English.
- The Python original is the behavior reference; don't modify that repo. Some quirks are kept on purpose and documented in code, e.g. `arknights.Barter` grants the post-exchange remainder as the roll count. Don't "fix" these without asking.
- Music controls (skip/stop/pause/prev/next/skipto/pop) are limited to the track's requester or an administrator (`requireTrackOwnerOrAdmin`).

## Tests

- Domain logic is unit-tested. Tests that touch the DB open a temp database with `db.Init(filepath.Join(t.TempDir(), "test.db"))` and call `db.Close()` in `t.Cleanup`.
- HTTP is stubbed with `httptest` by overriding package-level URLs (e.g. `reactions.baseURL`).
- Discord event glue isn't unit-tested. Pull logic out into pure helpers (e.g. `canManageGuild`, `targetDisplayName`, `findChannelByName`) and test those instead.
