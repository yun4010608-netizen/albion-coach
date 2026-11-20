---
version: 1
slug: "web-index-html"
primary_target: "web/index.html"
related_targets: []
---

# Surface: 알비온 코치 (single page: 코치 + 정제 계산기)

Scope: the whole web UI (web/index.html). Mode: operate.
Audience/job: one new Albion player on the Asia server, second screen (side monitor or phone) next to the game, deciding where to refine/sell and what to do next.
Constraints: Korean copy; stateless server; chat history in the browser; calculator must work without an API key; data age shown for every price.
Unresolved: station usage fee (excluded until a verified formula exists).

## Direction contract

THESIS: Refining profit read like a Korean stock quote. Cities are ranked as a 호가 ladder around a centre axis, gain in red and loss in blue (the Korean market convention). Refuses the category default of grey panels, gold accent and a plain sortable table.

OWN-WORLD: Night scene beside a dark game, so a graphite ground (#16181D) with numeral white (#ECE8DF). Red #E8413A = gain, blue #3B7BEA = loss, amber #F2B33D only for freshly refreshed values. Every rule and bar edge at 1 device pixel; no card chrome, no radius beyond 2px on inputs. Tabular numerals in fixed-width cells, so refreshes change digits, never layout. Each city row carries fixed furniture (city tincture chip, data age) in the same slot. Nothing labelled twice.

STORY: The player picks a resource and tier, reads one sentence that names the best city and the profit per unit, scans the ladder for alternatives, then asks the coach a follow-up in Korean.

FIRST VIEWPORT: 56px top bar: app name left, tab nav (코치 / 정제 계산기) used under 1100px, "Asia" server label right. ≥1100px: two panes, calculator left (filters row, answer sentence with the profit figure at ~56px, then the city ladder table), coach right (440px: messages, input pinned bottom). Under 1100px: one pane at a time via tabs; ladder rows drop price columns into a sub-line.

FORM: Korean brokerage quote board, candidate 3 of 7 on my ordered list. seed:c18d6d93e8f385d8. Signature move: profit bars extending left (blue) or right (red) from a 1px centre axis per city row, bars growing from the axis when a recipe loads.

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
