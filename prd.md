### Product Vision

A blazing-fast, terminal-based anime hub built in Go that allows users to search, select, and stream content directly from their command line, completely bypassing the web browser. The tool focuses on extreme stability, elegant terminal UI presentation, and seamless handoffs to native system players.

### Core Philosophy

Do not build advanced features until the core loop (**Search $\rightarrow$ Select $\rightarrow$ Play**) is bulletproof. The product must prioritize resilience against fragile upstream data sources over feature bloat.

### Scope & Boundaries

* **In Scope:**
* Interactive text-based Terminal User Interface (TUI).
* Smart provider routing with transparent failover behavior.
* Sub/Dub and language profile toggling.
* Native video playback handoff via external utilities.
* Local stream session tracking and history storage.


* **Out of Scope:**
* In-terminal raw video frame rendering (poor user experience and low frame rates).
* Local video file downloading or persistent block storage management.
* Centralized user account creation (the system remains completely local-first).



### Version Roadmap

#### Version 1 (The Working Prototype)

* **Goal:** Establish a robust core application loop.
* **Scope:** Text-only TUI, single metadata provider, manual sub/dub toggle, direct handoff to media player, and local JSON tracking.
* **Engineering Focus:** Validate terminal state management and process handoffs before introducing ecosystem complexity.

#### Version 2 (The Resilient Tool)

* **Goal:** Provide an uninterrupted user experience when external providers fail.
* **Scope:** Multi-provider fallback routing, automatic background URL token refreshing, and advanced terminal graphics protocol support (Kitty/Sixel) for cover art layouts.
* **Engineering Focus:** Shield the user from stream URL expirations and upstream maintenance outages.

#### Version 3 (The Power User Tool)

* **Goal:** Serve as a total desktop web browser replacement.
* **Scope:** Two-way tracking platform synchronization (AniList OAuth2 mutations), custom layout keybindings, and automated picture-in-picture workspace orchestration.
* **Engineering Focus:** Handle external authentication state and complex asynchronous API integrations.
