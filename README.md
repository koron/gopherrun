# Gopher Run!

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/gopherrun)](https://pkg.go.dev/github.com/koron/gopherrun)
[![Actions/Go](https://github.com/koron/gopherrun/workflows/Go/badge.svg)](https://github.com/koron/gopherrun/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/gopherrun)](https://goreportcard.com/report/github.com/koron/gopherrun)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/gopherrun)

Sample game using [Ebitengine](https://ebitengine.org/).

Blog post about this project in Japanese: [GolangとSDL2でゲームを作る / Making a game with Golang and SDL2](https://www.kaoriya.net/blog/2016/12/24/).

For the jumping sound effect, I used `jump07` from the [無料効果音で遊ぼう! / "Play with free sound effects!](https://taira-komori.jpn.org/game01.html) site.

## Build on Fedora Linux 42

Dependent Packages:

*   alsa-lib-devel
*   libXcursor-devel
*   libXi-devel
*   libXinerama-devel
*   libXrandr-devel
*   libXt-devel
*   libXxf86vm-devel
*   libglvnd-devel

How to install dependent packages:

```
sudo dnf install -y alsa-lib-devel libXcursor-devel libXi-devel libXinerama-devel libXrandr-devel libXt-devel libXxf86vm-devel libglvnd-devel
```
