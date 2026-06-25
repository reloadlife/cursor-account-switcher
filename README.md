# cursor-account-switcher

CLI to switch between Cursor accounts. Swaps auth session in `state.vscdb`, force-quits Cursor, restarts.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/reloadlife/cursor-account-switcher/main/install.sh | bash
```

Requires `~/.local/bin` on your `PATH`.

## Setup

Default accounts: **personal** and **work**.

```bash
cursor-switch save personal
cursor-switch save work
```

Add more with custom labels:

```bash
cursor-switch account add freelance --label "Freelance Client"
cursor-switch save freelance

# or register on first save:
cursor-switch save side-project --label "Side Project"
```

## Usage

```bash
cursor-switch                         # interactive TUI menu
cursor-switch switch work             # by id or label
cursor-switch save personal
cursor-switch status
cursor-switch account list
cursor-switch account add <id> --label "Display Name"
cursor-switch account remove freelance
```

## Build from source

```bash
make install
```

## Storage

Config: `~/.cursor-account-switcher/config.json`  
Profiles: `~/.cursor-account-switcher/profiles/` (auth tokens — keep private)
