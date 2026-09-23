# Texass — Arknights-themed Discord bot

Texass is a Discord bot with an Arknights gacha simulator, a music player,
anime reactions and a handful of meme commands. It is a Go rewrite of the
original Python bot, [reven-n1/DiscordBot](https://github.com/reven-n1/DiscordBot).

Built with [disgo](https://github.com/disgoorg/disgo),
[Lavalink v4](https://lavalink.dev) (via [disgolink](https://github.com/disgoorg/disgolink))
and SQLite (via [GORM](https://gorm.io), pure-Go driver, no CGO).

User-facing text is in Russian, carried over from the original bot.

## Commands

### Arknights simulator

| Command | Description |
|---|---|
| `/ark` | Roll a random operator and add it to your collection. 4h cooldown per user per server, with six-star soft pity. |
| `/myark [character] [public]` | Show your collection, or your copies of one operator. |
| `/barter` | Exchange duplicate operators for rolls of the next rarity tier. |

### Music

| Command | Description |
|---|---|
| `/play <query>` | Play a YouTube/Spotify link or search result, or add it to the queue. |
| `/skip`, `/stop`, `/disconnect` | Skip the current track, stop and clear the queue, leave the voice channel. |
| `/queue`, `/shuffle`, `/repeat <mode>` | Show or shuffle the queue; repeat off / one track / whole queue. |
| `/skipto <index>`, `/pop <index>` | Jump to a track, or remove it from the queue. |
| `/controls` | Player message with previous / pause / stop / next buttons. |

Skipping, stopping, pausing, jumping and removing tracks is limited to the
member who requested the track or a server administrator.

### Fun

| Command | Description |
|---|---|
| `/sfw <type>` | Random anime picture (waifu.pics). |
| `/reaction <type> [member]` | Anime reaction GIF with a phrase, optionally aimed at a member. |
| `/nsfw <type>` | NSFW anime picture. |
| `/ger` | "Farts" at a random member (or at yourself). 22h cooldown. |
| `/f [member]`, `/o7 [member]` | Pay respects / salute. Also available as user context-menu commands. |
| `/avatar [member]` | Show a member's avatar. Also available as a user context-menu command. |
| `/say <message> [ephemeral]` | Make the bot repeat a message. |

### Utility

| Command | Description |
|---|---|
| `/help` | List all commands. |
| `/info` | Bot info and global statistics. |
| `/invite` | Invite link for this bot. |
| `/ping` | Gateway latency. |
| `/announce <message> [channel]` | Post a message as the bot (administrators only). |
| `/settings [nsfw-content] [toxic-greetings]` | Per-server settings (Manage Server or Administrator). |

## Opt-in content

Some of the original bot's content is off by default and has to be enabled
per server with `/settings`:

- `nsfw-content` enables `/ark`, `/myark`, `/barter`, `/ger` and `/nsfw`.
  These commands also only work in channels marked as NSFW.
- `toxic-greetings` makes the bot post a rude message in the `основной`
  channel when a member leaves the server.

Running `/settings` without options shows the current values.

## Requirements

- A Discord application with a bot token. Enable the **Server Members**
  privileged intent in the Developer Portal (the bot needs it for
  `/ger`, the welcome/leave messages and the member cache).
- [Lavalink v4](https://lavalink.dev) for music. The bot doesn't wait
  for it: it keeps retrying the connection in the background (backing off
  up to 30s), and music commands say the music server is unavailable
  until Lavalink is up. If Lavalink restarts later, the bot reconnects on
  its own.
- Go 1.26+ to build from source, or Docker to use the prebuilt image.

## Configuration

Each setting can be passed as a command-line flag or an environment
variable. The environment variable wins if both are set.

| Flag | Environment variable | Default | Description |
|---|---|---|---|
| `-token` | `BOT_TOKEN` | — | Discord bot token (required). |
| `-lavalink-host` | `LAVALINK_HOST` | — | Lavalink host. |
| `-lavalink-port` | `LAVALINK_PORT` | — | Lavalink port. |
| `-lavalink-password` | `LAVALINK_PASSWORD` | — | Lavalink password. |
| `-db-path` | `DB_PATH` | `data/texas_bot.db` | SQLite database file; created if missing. |

Lavalink itself is configured by [`application.yml`](application.yml). It
loads the [youtube-source](https://github.com/lavalink-devs/youtube-source)
and [LavaSrc](https://github.com/topi314/LavaSrc) plugins on startup. For
Spotify support, set `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET` on the
**Lavalink** container (not the bot). You can get them from the
[Spotify developer dashboard](https://developer.spotify.com/dashboard).

## Running

### Docker Compose

[`docker-compose.yml`](docker-compose.yml) runs Lavalink and the bot
together. The bot uses the image
[`ghcr.io/wladbelsky/go_texas_bot/texas_bot:main`](https://github.com/wladbelsky/go_texas_bot/pkgs/container/go_texas_bot%2Ftexas_bot),
which CI builds from `main`. Replace the placeholder token, Lavalink password
and Spotify credentials, then run:

```sh
docker compose up -d
```

The SQLite database is stored in `./data`.

`docker-compose.yml` needs Compose v2 (the `docker compose` plugin). The old
Python `docker-compose` v1 (e.g. 1.25 from the Ubuntu 20.04 repos) rejects it
with `Unsupported config option for services: ...`. For v1, use the
equivalent legacy file (format 2.4, same services, limits and healthcheck):

```sh
docker-compose -f docker-compose.old.yml up -d
```

Both services have resource limits (`deploy.resources.limits`):

| Service | CPUs | Memory | Notes |
|---|---|---|---|
| `lavalink` | 2 | 1536M | JVM heap is `-Xmx1G`. Keep it well below the memory limit, or the container gets OOM-killed. |
| `texas` | 0.5 | 256M | `GOMEMLIMIT=200MiB` makes the Go GC stay under the limit. |

If you raise the Lavalink heap for a bigger bot, raise its memory limit with it.

Lavalink has a healthcheck (`curl` against its `/version` endpoint), and
the bot starts only once Lavalink is healthy. If Lavalink never becomes
healthy (for example, a plugin fails to download), `docker compose up`
reports the dependency failure and the bot container isn't started. Check
`docker compose logs lavalink`.

### From source

```sh
go build -o texas_bot .
./texas_bot -token "$BOT_TOKEN" \
  -lavalink-host localhost -lavalink-port 2333 -lavalink-password youshallnotpass
```

Slash commands are registered globally on startup, so Discord can take a
while to show new or changed commands.

## Development

```sh
go vet ./...
go test ./...
```

CI ([`.github/workflows/build.yml`](.github/workflows/build.yml)) runs the
tests, builds `linux/amd64` and `linux/arm64` binaries, and pushes a
multi-arch image to GHCR.

See [CLAUDE.md](CLAUDE.md) for the code layout and conventions, such as
how to add a command.

## Credits

- Operator and skin data come from
  [Aceship/AN-EN-Tags](https://github.com/Aceship/AN-EN-Tags) (en_US
  tables, trimmed into [`arknights/data/`](arknights/data)).
- Images come from [waifu.pics](https://waifu.pics).
- The original Python bot is by [reven-n1](https://github.com/reven-n1/DiscordBot).

Arknights is a trademark of Hypergryph / Yostar. This project is not
affiliated with them.

## License

[GPL-3.0](LICENSE)
