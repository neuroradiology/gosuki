<p align="center">
<img src="internal/webui/static/favicon.svg" height=50>
<h3 align="center">GoSuki</h3>
<h4 align="center">Multi-browser, standalone, bookmark manager</h4>
<h5 align="center">No subscription or extension required</h5>

 <h4 align="center">
  <a href="#-in-action">Demo</a> |
  <a href="https://gosuki.net/docs/getting_started/quickstart">Documentation</a> |
  <a href="https://gosuki.net/">Website</a>
</h4>
 <h5 align="center">
  <a href="#installation">Install</a> |
  <a href="#-features">Feautres</a>
  <!-- <a href="#rationale">Rationale</a> -->
</h5>


<br>
</p>



<h3 align="center">What's this ? Oh, just another bookmark manager. No big deal.</h3>

- [x] Ever feel like your bookmarks are a [chaotic mess](https://news.ycombinator.com/item?id=13734253) ?
- [x] Ever find yourself swapping browsers when they can't keep up with your demands?
- [x] Have past bookmark managers [let you down](https://support.mozilla.org/en-US/kb/future-of-pocket), vendor locking or abandoning you in your time of need?
- [x] Maybe you would rather avoid entrusting your bookmarks to cloud companies.
- [x] Perhaps you keep multiple browser profiles for work, research, streaming, and development?
- [x] Do you use some ["obscure"](https://github.com/qutebrowser/qutebrowser) browser that doesn't support extensions?

- [ ] If you're nodding your head to any of the above, then look no further.

**GoSuki** is an **extension-free**, **self-contained**, **real time** bookmark tracker and organizer. It packs everything in a **single binary** and captures all your bookmarks in a **portable database**.
<br>
<br>

## ✨ Features

- 📦 **Standalone**: Gosuki is a single binary with no dependencies or external extensions necessary. It's designed to just work right out of the box
- ⌨️ **Ctrl+D**: Use the universal shortcut to add bookmarks and call [custom commands](/docs/features/marktab-actions)
- 🏷️ **Tag Everything**: Tag with **#hashtags** even if your browser does not support it. You can even add tags in the Title. Your folders become tags
- 🔎 **Real time**: Gosuki keeps track of your bookmarks, spotting any changes as they happen
- 🖥️ **Web UI + CLI** Builtin, local Web UI. Also works without Javascript. dmenu/rofi compatible CLI.
- 🧪 **Hackable**: Modular and extensible. Custom scripts and actions per tags and folders.
- 🌎 **Browser Agnostic**: Detects which browsers you have installed and watch changes in all of them
- 👤 **Profile Support**: Also handles multiple profiles for each browser
- 💾 **Buku Compatibility**: Gosuki is compatible with the [Buku](https://github.com/jarun/buku) sqlite database, you can use any program that was made for buku
- 📡 **External APIs** Consolidate your curated content from external APIs (github, reddit ...)


## 📸 In Action

<div align="center">
  <p>
    <h3><a href="https://github.com/user-attachments/assets/bb5c52f8-4413-4f91-88c7-445834728952">Realtime Bookmark Tracker</a></h3>
    <video controls muted src="https://github.com/user-attachments/assets/bb5c52f8-4413-4f91-88c7-445834728952"></video>
  </p>


  <p>
    <h3><a href="https://github.com/user-attachments/assets/2e69940a-2fc3-4108-9b4c-ef324b3d08cd">Marktab Scripts</a></h3>
    <video controls muted src="https://github.com/user-attachments/assets/2e69940a-2fc3-4108-9b4c-ef324b3d08cd"></video>
    <p>Note: you can also drop bookmarks in a folder matching an action to execute the script. Folders are tags</p>
  </p>

    
  <p>
    <h3><a href="https://github.com/user-attachments/assets/bf1e7c87-5775-4c54-a428-cfe84757c43e">CLI - Suki</a></h3>
    <video controls muted src="https://github.com/user-attachments/assets/bf1e7c87-5775-4c54-a428-cfe84757c43e"></video>
  </p>

</div>

<br>
<p align="center"><a href="https://youtu.be/lxrzR4cHgmI" target="_blank">Full Demo on YT</a></p>

## Installation

Checkout the [quick start guide](https://gosuki.net/docs/getting_started/quickstart).

### Releases


#### From source

##### Dependencies:

- `sqlite3` development library

```console
 go install github.com/blob42/gosuki/cmd/gosuki@latest
 go install github.com/blob42/gosuki/cmd/suki@latest
```

Gosuki currently supports Linux and MacOS<sub>beta</sub> . More platforms are [planned](#roadmap).

## Running GoSuki

GoSuki is designed to run as a background service for real-time bookmark monitoring. Below are the recommended ways to start and interact with the application.

### As a Service
Start GoSuki as a persistent service:
```bash
gosuki start
```
This initializes all configured browsers and begins real-time bookmark tracking.

### Terminal UI (TUI)
Launch with an interactive terminal interface for real-time parsing overview:
```bash
gosuki --tui start
```
The TUI displays module status and bookmark processing metrics.

### Debugging
Enable detailed logging with:
```bash
gosuki --debug=2
```
*note*: Avoid using `--tui` with debug mode.

## How does it work ?

Gosuki monitors the browser's internal bookmark files for changes. It uses the native OS event notification system to detect changes as they happen. This allows it to be fast and efficient, without the need for any plugins or extensions.

The application maintains a portable database of all tracked bookmarks, accessible via the built-in web UI or CLI.

Curious for more details on the internals ? Checkout the [Architecture](docs/internal/architecture.md) file.

<!-- ## Rationale -->
<!-- TODO -->
<!-- Explain the reasons I made gosuki -->
<!-- Redirect to the comparaison matrix with other projects -->

## Roadmap

- [ ] **Archival** (top most priority) - Archive bookmarks in a portable format for offline access.
- [ ] **Linkrot** - Automatically identify broken links and replace with web.archive.org alternatives
- [ ] **Built-in Sync** - Enable secure synchronization between multiple devices
- [ ] **Tag Refresh** - Automatically clean and update tags/metadata for existing bookmarks
- [ ] **Simple Cleanup UI** - Intuitive interface for organizing and pruning bookmarks
- [ ] **Browser Sync** - Push changes back to browsers for consistent bookmark management
- [ ] **Platforms** - FreeBSD, Windows and other platforms

## Support

GoSuki is a one-man project. If you find value in this tool, consider supporting its development through:

- Contributions via [GitHub Sponsors](https://github.com/sponsors/blob42) or [Patreon](https://www.patreon.com/c/GoSuki)
- Reporting issues and suggesting features
- Contributing code or documentation
- Sharing the project with others who might benefit

Your support helps maintain existing features and allows me to develop new capabilities for all users.

## Contributing
We welcome contributions from the community! To get started:
1. Fork the repository
2. Create a new branch for your changes
3. Submit a pull request with clear documentation
4. Follow our [code of conduct](CODE_OF_CONDUCT.md)

For bug reports, please provide detailed steps to reproduce the issue.


## Related Projects 

Read the ["how does it compare to"](docs/how-does-it-compare-to.md) guide.

- [Buku](https://github.com/jarun/buku): Gosuki is compatible with Buku
- [Shiori](https://github.com/go-shiori/shiori): Simple bookmark manager built with Go
- [bmm](https://github.com/dhth/bmm): get to your bookmarks in a flash
- [wallabag](https://github.com/wallabag/wallabag): self hostable application for saving web pages

## Links & Discussions

- [Ask HN: Do you still use browser bookmarks?](https://news.ycombinator.com/item?id=14064096)
- [Ask HN: How to handle bookmarks so you can find them again?](https://news.ycombinator.com/item?id=13734253)
- [Reddit: Does anyone actually use mobile bookmarks](https://www.reddit.com/r/firefox/comments/dez7hh/does_anyone_actually_use_mobile_bookmarks/)
- [You are the dead internet](https://www.youtube.com/watch?v=aoTQPoz9_As)

