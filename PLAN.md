# Rebuild plan: noisy passcode entry tool (`passgen`)

## Context

`passgen` generates a random passcode you deliberately don't memorize, stores it, and later walks
you through **typing it onto another device** (e.g. an iPhone Screen Time prompt) as a stream of
keystroke instructions — real keystrokes interleaved with noise — so the act of entering it never
teaches you the code. Accessing the stored secret is acceptable; this is an anti-memorization aid,
not cryptography.

The current `internal/noise_steps.go` algorithm is broken: it patches a target string with ad-hoc
deadline handling and post-hoc cleanup passes, uses a `REPLACE` op that isn't a real keystroke on a
passcode field, `stringifyAndHide` reveals the buffer verbatim (defeating the purpose), and the
author's own commit notes it fails "on some cases". We are replacing the core algorithm and the I/O
flow from scratch. Tests asserting the old free-text step format will be rewritten.

### Locked decisions
- **Targeted generation everywhere.** Pick the random code `S` first, then engineer a keystroke walk
  that lands exactly on `S`. (Chosen over the "emergent" model because re-entry must reproduce a
  specific stored code with fresh noise — only a targeted generator can do that, so one code path.)
- **Auto-submit field** (iPhone numeric passcode): the field submits the instant the buffer reaches
  length `N`, so the buffer must stay `<= N-1` until the single final correct keystroke. → *capped
  noise model.*
- **Default run = combined** (generate + save + walk through once). Flags split into generate-only
  and enter-only so a saved code can be re-entered later with fresh noise.
- **Press Enter** to reveal each next keystroke; the resulting code is never shown.

## Core logic

### Keystroke model
A passcode field supports only two primitives: **ADD `<char>`** (append) and **BACKSPACE** (remove
last char) — no "replace". The buffer behaves like a stack; replaying any ADD/BACKSPACE stream
deterministically yields a final string. Our job: emit a stream that ends exactly at `S` (length
`N`), padded with noise.

### State — committed-prefix invariant (provably correct)
- `k` = number of leading correct chars permanently committed; buffer's first `k` chars == `S[0:k]`.
- `g` = count of garbage (noise) chars currently on top. Buffer == `S[0:k]` + `g` garbage chars.
- **Invariant: never BACKSPACE a committed char** — backspaces only remove garbage. Correctness is
  then trivial: drive `g→0` and `k→N` and the buffer equals `S`.

Operations and when they're allowed:
- `ADD_CORRECT` — only when `g == 0` and `k < N`: append `S[k]`, `k++`. (real progress)
- `ADD_NOISE` — append a random char `!= S[k]`, `g++`. (noise)
- `BACKSPACE` — only when `g > 0`: remove a garbage char, `g--`.

### Auto-submit cap
The field submits when buffer length hits `N`. Enforce **`k + g <= N-1` at all times**, except the
final keystroke (an `ADD_CORRECT` taking `k` from `N-1` to `N`). Concretely: `ADD_NOISE` is allowed
only when `k + g <= N-2`.

### Budget / guaranteed termination in exactly `steps` keystrokes
Accounting: `#ADD - #BACKSPACE = N` and `#ADD + #BACKSPACE = steps`, so `#BACKSPACE = (steps-N)/2`.
Thus `steps >= N` and `steps ≡ N (mod 2)`; each noise excursion costs 2 keystrokes (add + matching
backspace). When deriving `steps`, round up to the nearest value with correct parity.

Per-step controller (`i` from 0, `R = steps - i` remaining):
1. `minNeeded = g + (N - k)` — backspaces to clear current garbage + remaining correct adds.
2. **Forced finish:** if `R == minNeeded` → drain mode: if `g > 0` do `BACKSPACE`, else `ADD_CORRECT`.
3. **Otherwise** (`R - minNeeded >= 2`, guaranteed by parity) randomly choose among currently allowed
   ops: `ADD_NOISE` (only if `k+g <= N-2`), `BACKSPACE` (only if `g>0`), `ADD_CORRECT` (only if
   `g==0 && k<N`). Weight to taste (e.g. favor noise early, progress late).
4. After the loop assert `k == N && g == 0`; replay the emitted steps internally and verify the
   result equals `S` before anything is shown to the user (defensive — should be impossible to fail).

This replaces the entire current generator, including the deadline hacks, the cleanup passes, the
`REPLACE` path, and the dead `isCorrectDigitInCorrectPosition`.

### What the operator sees (anti-memorization + masking)
Each step prints **only the bare keystroke** — e.g. `Press 7` or `Backspace`. It must **never** reveal
the building buffer, the final code, a code-revealing counter, or any label/marking/coloring that
distinguishes a real keystroke from a noise one. Real and noise adds render identically (`Press X` for
both); the operator cannot tell which presses survive. Press Enter to advance; the previous line is
cleared via the existing `deleteLine()` ANSI helper in `main.go`. Because the stream is re-randomized
every entry for the same stored code, no fixed muscle-memory pattern forms.

Accepted limitation: anyone who *records* the keystroke stream can replay the stack to recover the
code, and a backspace reveals the immediately prior add was noise. This is an aid, not crypto.

### Optional future enrichment (not in initial build)
To remove the faint structural tell (correct chars only added when `g==0`), allow the walk to
occasionally backspace into committed chars and retype them, so "correct" no longer implies
"permanent". Out of scope for now; noted for later.

### Randomness
Use `crypto/rand` for passcode generation and all action/char choices (replacing the current unseeded
`math/rand`). Provide a small `randInt(n int) int` helper so the generator depends on it rather than a
`*rand.Rand`.

## Commands / I/O

Default action `passgen` (combined): parse flags → build charset → generate `S` → save (base64) →
immediately walk through entry once.

Flags to split the two moments:
- `--generate-only` / `-g`: generate + save, no walkthrough.
- `--enter <file>` / `-e <file>`: skip generation; read + base64-decode the saved code from `<file>`,
  generate fresh noise, walk through. Re-runnable any time for the same code.

Existing flags kept/adapted: `--charset/-c`, `--length/-l`, `--min/-m`, `--max/-x` (step range),
`--output/-o`, `--lowercase/-L`. Step count derived as today (random in `[min,max]`, floored to `>N`),
then **rounded up to satisfy `steps >= N` and the parity rule**. For Screen Time, `--charset numeric`.

## Files to change
- `internal/noise_steps.go` — **rewrite**. New generator (e.g. `GenerateKeystrokes`) implementing the
  committed-prefix capped model; emit **structured steps** (`[]Step{Action, Char}`, Action ∈
  {AddCorrect, AddNoise, Backspace}) rather than free-text, so `main.go` formats the display and tests
  assert cleanly. Delete `stringifyAndHide` and `isCorrectDigitInCorrectPosition`.
- `internal/passcode_gen.go` — switch `GeneratePassword` to `crypto/rand`; add a helper to draw a noise
  char `!= S[k]`.
- `main.go` — add `--generate-only` / `--enter` handling and the read+decode path; format `Press X` /
  `Backspace` display from structured steps; keep `deleteLine()` and base64 storage.
- `internal/charset.go` — unchanged (reused as-is).
- `test/passcodeGen_test.go`, `test/charset_test.go` — keep; adjust only if signatures change.
- `test/integration_test.go` — **rewrite** the simulator: replay the structured ADD/BACKSPACE stream
  through a stack and assert: final == generated code, the auto-submit cap (`len <= N-1` until the
  final keystroke) is never violated, and total keystrokes == `steps`. Sweep many seeds × lengths
  (1..20) × charsets.

## Verification
1. `go build ./...` and `go vet ./...` clean.
2. `go test ./...` — integration test replays the keystroke stream over thousands of random seeds ×
   lengths × charsets, asserting final == code, cap never violated, count == steps, parity correct.
3. Manual generate: `go run . -c numeric -l 4` → follow the `Press X`/`Backspace` steps on a scratch
   buffer; confirm you land on the saved code (`base64 -d passcode.enc`).
4. Manual re-entry: `go run . -g -c numeric -l 6 -o code.enc`, then `go run . -e code.enc` twice —
   confirm the two keystroke sequences differ but both reconstruct the same code.
