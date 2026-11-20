# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Static HTML/CSS/JS (no build step), embedded in and served by a Go single binary. Runs locally with docker compose (2+ app replicas behind a reverse proxy).

## Users

One player (the owner) on the Albion Online **Asia** server who is new to the game and wants to try combat, gathering, and crafting/trading. Uses the tool on a second screen next to the game: a side monitor on the Mac, or a phone propped beside it.

## Product Purpose

Help a new player progress efficiently without automating the game. Two jobs:

1. **Coach**: answer "what should I do next?" in Korean, grounded in a researched early-game guide, live market prices, and (optionally) the player's public fame stats.
2. **Refining calculator**: show which raw→refined resource step is profitable in which Asia city right now.

Success = the player makes a decision (where to gather, what to refine, where to sell) in under a minute without leaving the game for long.

## Positioning

An advisory tool only. It never sends input to the game client and never reads game memory or packets. Data comes from public APIs (Albion Online Data Project market prices, official gameinfo API).

## Operating Context

- Glanced at between game actions; often read on a phone at arm's length.
- The game itself is dark and busy; the tool sits beside it.
- Market data is crowd-sourced and can be stale; data age matters for every price shown.

## Capabilities and Constraints

- Korean UI copy; game item and city names may appear in English as the API returns them.
- Chat history lives in the browser; the server is stateless (any replica can serve any request).
- Claude API key is supplied by the owner at runtime (`ANTHROPIC_API_KEY`); the calculator must work without it.
- Undecided: station usage fee in profit math (excluded unless a verified formula is found).

## Evidence on Hand

No screenshots, brand assets, or logos. Do not use Albion Online's logo or official art (not ours to use).

## Product Principles

- Decision first: every view leads with the answer (best city, profit per unit), then the numbers behind it.
- Show data freshness next to every price; never present stale data as current.
- Never automate gameplay.
- Works one-handed on a phone.
