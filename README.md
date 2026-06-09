# passgen — Noisy Passcode

🚧 This project is currently under testing, use at your own risk.

A CLI tool for when you need a passcode you deliberately *don't* memorize, but
still have to type by hand from time to time (e.g. an iPhone Screen Time
passcode).

It generates a random passcode, saves it, and then walks you through entering it
on the target device as a stream of single-keystroke instructions — real
keystrokes interleaved with noise (`Press 7`, `Backspace`, `Press 2`, …). Real
and noise presses look identical and the sequence is **re-randomized every time
you enter the code**, so the act of typing it never teaches it to your fingers.

> This is an anti-memorization aid, not cryptography. The saved code is stored
> Base64-encoded and is recoverable by design.

## Usage

```sh
# Generate, save, and immediately walk through entering a 4-digit code (default):
go run . -c numeric -l 4

# Generate and save only (no walkthrough):
go run . -g -c numeric -l 6 -o code.enc

# Re-enter a previously saved code, with a fresh noise sequence each run:
go run . -e code.enc
```

Press Enter after performing each keystroke on your device; the instruction is
cleared before the next one appears, and the passcode itself is never shown.

### Flags

| Flag | Alias | Default | Description |
| --- | --- | --- | --- |
| `--charset` | `-c` | `numeric` | `alphabet` \| `numeric` \| `alphanumeric` \| `all` |
| `--length` | `-l` | `4` | Passcode length |
| `--min` | `-m` | `12` | Minimum keystroke steps |
| `--max` | `-x` | `18` | Maximum keystroke steps |
| `--output` | `-o` | `passcode.enc` | File to save the generated code to |
| `--lowercase` | `-L` | off | Lowercase characters only |
| `--generate-only` | `-g` | off | Generate + save, skip the walkthrough |
| `--enter <file>` | `-e` | — | Skip generation; enter a saved code |

## How it works

The tool targets an **auto-submitting** field (the passcode submits the instant
it reaches the required length), so the on-device buffer never reaches the full
length until the very last keystroke. Each entry, a fresh ADD/BACKSPACE sequence
is built that provably reconstructs the exact stored passcode while staying under
that cap. See `PLAN.md` for the full algorithm (committed-prefix invariant, step
budgeting, and the cap).

## Tests

```sh
go test ./...
```

The integration test replays tens of thousands of generated sequences
(seeds × charsets × lengths) and asserts each one reconstructs its passcode,
never violates the auto-submit cap, and uses exactly the requested keystroke
budget.
