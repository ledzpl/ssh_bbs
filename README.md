# SSH Text BBS

Simple SSH-accessible text BBS implemented in Go.

## Run

- Start the server: `go run ./cmd/bbs -addr :2323`
- Connect from another terminal: `ssh guest@localhost -p 2323` (any username works; no password; add `-o StrictHostKeyChecking=no` on first connect if needed).

Commands inside the shell: arrow keys to navigate boards/posts, Enter to select, `w` to write, `b`/Left to go back, `q` to quit. Default boards: `general`, `tech`.

Persistence:
- Board list JSON (default `data/boards.json`) keeps board names and order.
- Posts per board are saved as JSON in `data/posts/<board>.json` with a versioned wrapper so future post field changes stay compatible.

Authentication (optional):
- Provide `-auth path/to/auth.json` where the file is `{"users":[{"username":"alice","password":"secret"}]}`.
- If present, connections must log in with one of the users (SSH username is used if it matches). Without the flag, auth is disabled.

Encryption (optional):
- Set the `BBS_ENCRYPTION_KEY` environment variable to encrypt post storage files.
- Key must be 32 bytes (64 hex characters). Generate with: `openssl rand -hex 32`
- Example: `export BBS_ENCRYPTION_KEY=0000000000000000000000000000000000000000000000000000000000000000`
- When enabled, post files are encrypted using AES-GCM. Existing plain-text files can still be read (backward compatible).
- ⚠️ **Important**: Keep your encryption key safe. Lost keys cannot recover encrypted data.

## Development

- Format: `go fmt ./...`
- Test: `GOCACHE=$(pwd)/.gocache go test ./...`
