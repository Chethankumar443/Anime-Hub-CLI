**V1 Testing:**
*   **Unit Tests:** Mock the HTTP layer using `httptest`. Verify that the AniList client correctly parses the GraphQL JSON response into Go structs.
*   **TUI Tests:** Use Bubbletea's `teatest` package. Simulate key presses (Type "Naruto", press Enter) and assert the final model state.

**V2 Testing:**
*   **Integration Tests:** Test the `FallbackManager`. Mock Provider A to return an error, and Provider B to return success. Assert that the final URL comes from Provider B.
*   **Database Tests:** Write tests that spin up an ephemeral SQLite database, write history, and read it back to ensure no data corruption.

**V3 Testing:**
*   **Security/Update Tests:** Mock the GitHub release API. Feed it a binary with a mismatched SHA256 checksum. Assert that the updater aborts and returns an error.
