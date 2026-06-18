# Freebusy

> Hasura - `postgresql://postgres:postgrespassword@local.hasura.dev:5432/freebusydb`
> Prisma - `postgresql://postgres:postgrespassword@localhost:5432/freebusydb`

The pure availability engine: given the facts about a bookable resource and a
date range, it computes what's bookable — deterministic functions over plain Go
structs. No database, no auth, no payments, no network.

## What's open vs. closed

This repository contains two things:

- **The engine** — the functional core. It expands recurring availability
  (RRULE) against the resource's timezone, applies blackout/closure exceptions,
  honors stay rules (min/max nights, check-in days, advance windows, buffers,
  gaps), and computes per-night free counts, bookable ranges, and free-unit
  picks over a pool of interchangeable units.
- **The API contract** (`protobuf/`) — the protobuf definitions for the *whole*
  system, including the parts whose implementation is closed. The contract is
  open so clients and tooling can be generated against it; the stateful shell
  that implements bookings, holds, multi-tenancy/RLS, auth, pricing, promo
  codes, payments, and external calendar feeds is not part of this repository.

One line: freebusy answers "is this free, and which unit?" deterministically —
everything stateful or external is the shell's job.

## Code generation

Generated artifacts under `protobuf/generated/` are produced from the protobuf
definitions via [`buf`](https://buf.build) and recipes in the `justfile`:

- `just generate` — regenerate everything (Go, OpenAPI, ORM, docs)
- `just orm` — regenerate the ORM schemas (Prisma + GORM) only
- `just docs` — regenerate the protobuf docs only

ORM generation requires the `protoc-gen-protorm` plugin to be installed and on
your `PATH`. Install it globally, e.g. via Homebrew or `go install`, before
running `just orm` / `just generate`. The generated ORM output reflects whatever
`protoc-gen-protorm` version is installed, so keep it up to date.

---

© 2026 oh-tarnished | Apache 2.0 License
