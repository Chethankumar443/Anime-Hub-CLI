**What to test:** The data flow, the provider logic, and the process lifecycle. Ignore visual Bubbletea testing.

**Practical Test Cases:**
1.  **The "Expired URL" Test:** Mock the provider to return a valid URL on the first call, and a 404 on the second. Verify the `FallbackManager` routes to the second provider.
2.  **The "Missing Dependency" Test:** Run the app without `mpv` in the `$PATH`. Verify it catches `exec.LookPath` and prints a styled error with OS-specific install commands.
3.  **The "Zombie Process" Test:** Force-kill the Go CLI while Consumet is running. Restart the CLI. Verify it detects the orphaned `consumet-win.exe` process and either kills it or reuses the port without crashing.
4.  **The "Port Conflict" Test:** Occupy port 3000 with a dummy server. Start the CLI. Verify the `ConsumetManager` automatically falls back to port 3001.
