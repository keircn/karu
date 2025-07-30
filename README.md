# Karu

ani-cli clone written in Go

## Installation

### From Source

```bash
git clone https://github.com/keircn/karu
cd karu
go build -o karu cmd/karu/main.go
```

## Usage

### Search and Watch Anime

```bash
# Interactive search
./karu search

# Direct search
./karu search "bocchi the rock"
```

### Options

```bash
# Show version
./karu version

# Show help
./karu --help
```

## Supported Players

Karu automatically detects and uses available video players:

- **mpv** (recommended)
- **iina** (macOS)
- **vlc**
- **flatpak mpv**

## Dependencies

- Go
- A supported video player (mpv recommended)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## Disclaimer

This project is developed under the MIT license and you are free to do whatever you please with the code. This software is intended for educational purposes. Please respect content creators and use legal streaming services when available. We do not host any illegal content, all sources are independent of Karu. (the lawyers made me say that :<)

## Special thanks

- Inspired by [ani-cli](https://github.com/pystardust/ani-cli)
- Built with [Cobra](https://github.com/spf13/cobra) and [Bubble Tea](https://github.com/charmbracelet/bubbletea)
