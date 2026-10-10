# terminalscore

[Türkçe](README.tr.md)

Live football scores in your terminal. Every match of the day, match events and league
tables, from all over the world. Data comes from ESPN, no API key needed.

On first launch it asks for your favorite team; the competition your team is playing in
and its league go to the top of the list. Other leagues follow by size.

## Install

Linux / macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.sh | sh
```

Installs to `~/.local/bin` (or `/usr/local/bin` when run as root), no sudo needed.
The download is verified against the release checksums.

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.ps1 | iex
```

`.deb` / `.rpm` packages and archives for other systems are on the
[releases](https://github.com/demirdenizargun/terminalscore/releases/latest) page:

```sh
sudo apt install ./terminalscore_*_linux_amd64.deb
sudo dnf install ./terminalscore_*_linux_amd64.rpm
```

With Go:

```sh
go install github.com/demirdenizargun/terminalscore@latest
```

## Usage

```
terminalscore
```

| Key | |
|---|---|
| ← → | change day (league on the standings tab) |
| t | today |
| ↑ ↓ | select match / scroll |
| PgUp PgDn | jump 10 matches |
| enter | match details |
| esc | back |
| tab | matches / standings |
| f | favorite team |
| r | refresh |
| q | quit |

While matches are live the list refreshes every 10 seconds and an open match every 5.
A match with a new goal blinks GOAL! for a minute.

The interface is in English or Turkish, picked from your system language. To override it,
use `terminalscore --lang tr` or set `"language": "en"` in the config file:
`~/.config/terminalscore/config.json` (`%AppData%\terminalscore\config.json` on Windows).
Deleting that file makes it ask for your favorite team again.

## Building from source

```sh
go build -o terminalscore .
```

## Contributing

Translations are welcome: add a `texts` value in `internal/ui/text.go` and league names
in `internal/espn/leagues.go`.

## License

MIT. Match data belongs to ESPN; this project is not affiliated with ESPN. The endpoints
used are not an officially documented API and may change at any time.
