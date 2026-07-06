*This document is designed to be fed into an AI IDE (Cursor, Windsurf, Aider) as context, or used by a human developer to onboard. It is structured to be reusable by swapping the `[Bracketed]` context for future projects.*

**1. Core Engineering Skills (The "How to Build" Matrix)**
*   **Language Mastery:** [Go 1.26.4]. Focus on interfaces, concurrency (goroutines/channels), and error handling. *Rule: Never use `panic()` in application logic; always return errors.*
*   **UI Paradigm:** [Terminal User Interface (TUI)]. Framework: [Bubbletea]. *Rule: All UI state must be immutable updates. Never mutate the model directly in the View function.*
*   **External Execution:** [Process Management]. *Rule: When spawning external binaries (like `mpv`), always capture stderr and handle non-zero exit codes gracefully.*

**2. AI-Assisted Development Workflow (How to use Cursor/Windsurf/Aider)**
*   **Context Loading:** Always attach `architecture.md` and `trd.md` to the AI chat before generating code. 
*   **Rule Setting:** Create a `.cursorrules` file in the root directory. 
    *   *Content:* "You are an expert Go developer. Always use `golangci-lint` standards. Prefer composition over inheritance. Do not write raw HTML scrapers; use the defined Provider interfaces. Keep functions under 40 lines."
*   **Iterative Prompting:** Do not ask the AI to "build the whole app." Ask it to "implement the `Search` state in Bubbletea based on `architecture.md`."

**3. DevOps & CI/CD Skills (GitHub Actions)**
*   **Pipeline Design:** Use GitHub Actions for automated testing and GoReleaser for binary compilation.
*   **Caching:** Always cache Go modules (`actions/setup-go@v5` with `cache: true`) to reduce CI build times from 3 minutes to 30 seconds.
