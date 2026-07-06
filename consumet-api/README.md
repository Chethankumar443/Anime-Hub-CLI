# Consumet & AniList GraphQL Node.js API Server

This Node.js API server wraps Gogoanime scraping and AniList GraphQL queries using `@consumet/extensions`. It is designed to run as a backend companion for the Anime-Hub-CLI.

## Features
- **Gogoanime endpoints** for searching, detailed metadata (episodes), and playing stream sources.
- **AniList Meta endpoints** using AniList GraphQL queries matching sources.
- **Localhost Binding:** Binds strictly to `127.0.0.1` to prevent triggering Windows Defender Firewall popups on launch.
- **Standalone Packaging:** Pre-configured with `pkg` to build compiled executables.

---

## Local Development

### 1. Install Dependencies
```bash
npm install
```

### 2. Run the Server
```bash
npm start
```
The server will start listening at `http://127.0.0.1:3000`.

---

## Compiling Standalone Binaries
To build executables for distribution (without needing Node.js installed on the user's target machine), use the `build` script which utilizes `pkg`:

```bash
npm run build
```
This will compile and output binaries into the `dist/` folder:
- `dist/consumet-api-win.exe` (Windows)
- `dist/consumet-api-macos` (macOS)
- `dist/consumet-api-linux` (Linux)
