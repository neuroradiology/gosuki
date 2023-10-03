<h1 align="center">🔖 GoSuki</h1>

## Blazing fast, Plugin-Free, Multi-Browser Bookmark Manager

**ALPHA**: This is an early preview work in progress, expect bugs and API changes.

### What's this ? Oh, just another bookmark manager. No big deal.

- Ever feel like your bookmarks are a chaotic mess ?
- Ever find yourself swapping browsers when they can't keep up with your demands?
- Have past bookmark managers let you down, vendor locking or abandoning you in your time of need?
- Maybe you're cautious about entrusting your bookmarks to unreliable cloud companies?
- Or perhaps you keep multiple browser profiles for work, research, streaming, and web development?
- Even pondering about monks who may be using an "obscure" browser that doesn't support extensions?

If you're nodding your head to any of the above, then look no further. 

GoSuki is a blazing fast real time bookmark manager with zero plugin
dependencies and a single binary. It's designed to work in the background and
manage your bookmarks across all your browsers, profiles.

## Features in a nutshell

- 🔌 **Standalone**: Gosuki is a single binary with no dependencies or external extensions necessary. It's designed to just work right out of the box.
- 🔖 **Quick Bookmarking**: Gosuki leverages the  universal shortcut `ctrl+d` with native bookmarks UI that exists in all browsers.
- 🏷️ **Tagging**: You can tag your bookmarks in any browser. In Chrome, for example, you can include `#tag1 #tag2` in your bookmark title.
- 🔎 **Constant Monitoring**: Gosuki keeps track of your bookmarks, spotting any changes as they happen.
- 🖌️ **Customizable**: You can add commands in your bookmark title to initiate certain Gosuki actions, like archiving a bookmark with `:archive`.
- 🌎 **Browser Agnostic**: Detects which browsers you have installed and watch changes in all of them.
- 👤🔀 **Profile Support**: Also handles multiple profiles for each browser.
- 💾 **Buku Compatibility**: Gosuki is compatible with the [Buku](https://github.com/jarun/buku) sqlite database, you can use any program that was made for buku.

## Quick Start

### Installation

`go install github.com/blob42/gosuki/cmd/gosuki@latest`

Gosuki currently supports Linux and WSL on Windows. MacOS support is planned for the future.


## How does it work ?

Gosuki monitors the browser's internal bookmark files for changes. It uses the native OS event notification system to detect changes as they happen. This allows it to be fast and efficient, without the need for any plugins or extensions.

Curious for more details on the internals ? Checkout the [Architecture](docs/artchitecture.md) file.
