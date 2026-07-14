# cursor-account-switcher

CLI to switch between saved accounts across AI coding platforms. Swaps auth sessions, optionally force-quits and restarts the host app.

## Supported platforms

| Platform | ID | Auth storage | Restarts app |
|----------|-----|--------------|--------------|
| Cursor | `cursor` | state.vscdb | yes |
| Claude Code | `claude` | ~/.claude/.credentials.json + keychain | no |
| Codex | `codex` | ~/.codex/auth.json | no |
| Grok | `grok` | ~/.grok/auth.json | no |
| VS Code / GitHub Copilot | `vscode` | state.vscdb + keychain | yes |

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/reloadlife/cursor-account-switcher/main/install.sh | bash
```

Requires `~/.local/bin` on your `PATH`.

## Setup

Pick a platform (defaults to Cursor):

```bash
cursor-switch platform list
cursor-switch platform use claude
```

Default accounts: **personal** and **work**.

```bash
cursor-switch save personal
cursor-switch save work
```

Add more with custom labels:

```bash
cursor-switch account add freelance --label "Freelance Client"
cursor-switch switch freelance          # clear live login → empty slot
# sign into the tool as the new account, then:
cursor-switch save freelance

cursor-switch save side-project --label "Side Project"
```

## Usage

```bash
cursor-switch                         # interactive TUI menu
cursor-switch switch work             # by id or label (restores saved profile)
cursor-switch switch freelance        # empty slot → clears live auth for a fresh sign-in
cursor-switch save personal
cursor-switch status
cursor-switch account list
cursor-switch account add <id> --label "Display Name"
cursor-switch account remove freelance

cursor-switch p list                    # alias for platform list
cursor-switch p use claude              # set default platform
cursor-switch --platform codex save work
cursor-switch --platform grok save personal
cursor-switch --platform grok switch work

# Isolated HOME for parallel agents (does not change global login)
cursor-switch --platform grok materialize work --dest /tmp/grok-work-home
HOME=/tmp/grok-work-home grok
```

Each platform keeps its own account list and saved profiles.

## Platform notes

**Claude Code** — saves `claudeAiOauth` tokens and `oauthAccount` from `~/.claude.json` (email lives there, not in credentials). Keychain restore merges tokens without wiping MCP plugin auth.

**Codex** — requires file-based auth (`cli_auth_credentials_store = "file"` in `~/.codex/config.toml`). Keyring-only mode is not supported yet.

**Grok** — swaps `~/.grok/auth.json` (OIDC session). Use `materialize` + `HOME=…` for parallel sessions without clobbering the global login.

**VS Code / Copilot** — swaps GitHub-related state.vscdb keys and Keychain entries. Sign out/in once after switching if Copilot doesn't pick up the new session.

## Build from source

```bash
make install
```

## Storage

Global config: `~/.cursor-account-switcher/config.json` (default platform)

Per platform:
```
~/.cursor-account-switcher/cursor/config.json
~/.cursor-account-switcher/cursor/profiles/
~/.cursor-account-switcher/claude/...
~/.cursor-account-switcher/codex/...
~/.cursor-account-switcher/vscode/...
```

Profiles contain auth tokens — keep private.
