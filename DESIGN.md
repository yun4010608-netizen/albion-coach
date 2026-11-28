---
name: 알비온 코치
description: A second-screen refining profit board and Korean coach for a new Albion Online Asia player, read like a Korean stock quote.
colors:
  up: "#e8413a"
  up-ink: "#ff6a5f"
  down: "#3b7bea"
  down-ink: "#7aa6ff"
  amber: "#f2b33d"
  ground: "#16181d"
  raise: "#1d2026"
  ink: "#ece8df"
  mute: "#9a9da5"
  rule: "#2c3038"
  edge: "#6b6f78"
  city-thetford: "#6b3f8e"
  city-martlock: "#2c5aa0"
  city-lymhurst: "#3e7d3a"
  city-bridgewatch: "#d98a1f"
  city-fort-sterling: "#e6e6e1"
  city-caerleon: "#9e2a2b"
  city-brecilien: "#3a8c8c"
typography:
  display:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "56px"
    fontWeight: 700
    lineHeight: 1.35
    letterSpacing: "-0.03em"
    fontFeature: "tnum"
  display-phone:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "40px"
    fontWeight: 700
    lineHeight: 1.35
    letterSpacing: "-0.03em"
    fontFeature: "tnum"
  headline:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "30px"
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: "-0.015em"
  headline-phone:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "22px"
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: "-0.015em"
  title:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "17px"
    fontWeight: 700
    lineHeight: 1.55
    letterSpacing: "-0.01em"
  body:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.55
  body-long:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.65
  numeric:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.55
    fontFeature: "tnum"
  meta:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.55
  label:
    fontFamily: '"Pretendard Variable", system-ui, -apple-system, "Apple SD Gothic Neo", "Noto Sans KR", sans-serif'
    fontSize: "12px"
    fontWeight: 500
    lineHeight: 1.55
rounded:
  none: "0px"
  sm: "2px"
spacing:
  control-y: "6px"
  gap: "8px"
  cell-x: "10px"
  control-x: "12px"
  row-y: "13px"
  gutter-phone: "16px"
  section: "20px"
  gutter-coach: "24px"
  gutter-calc: "28px"
components:
  top-bar:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.ink}"
    typography: "{typography.title}"
    height: "56px"
    padding: "0 16px"
  tab:
    textColor: "{colors.mute}"
    typography: "{typography.body}"
    height: "56px"
    padding: "0 12px"
  tab-selected:
    textColor: "{colors.ink}"
  button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.ground}"
    rounded: "{rounded.sm}"
    padding: "8px 18px"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.mute}"
    rounded: "{rounded.sm}"
    padding: "6px 12px"
  button-ghost-hover:
    textColor: "{colors.ink}"
  segment:
    textColor: "{colors.mute}"
    rounded: "{rounded.none}"
    padding: "7px 13px"
  segment-selected:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.ground}"
  field:
    backgroundColor: "{colors.raise}"
    textColor: "{colors.ink}"
    rounded: "{rounded.sm}"
    padding: "6px 8px"
  field-ask:
    backgroundColor: "{colors.raise}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.sm}"
    padding: "10px 12px"
  answer-headline:
    textColor: "{colors.ink}"
    typography: "{typography.headline}"
  answer-figure-gain:
    textColor: "{colors.up-ink}"
    typography: "{typography.display}"
  answer-figure-loss:
    textColor: "{colors.down-ink}"
    typography: "{typography.display}"
  ladder-head:
    textColor: "{colors.mute}"
    typography: "{typography.label}"
    padding: "12px 10px"
  ladder-cell:
    textColor: "{colors.ink}"
    typography: "{typography.numeric}"
    padding: "13px 10px"
  profit-bar-gain:
    backgroundColor: "{colors.up}"
    rounded: "{rounded.none}"
    height: "12px"
  profit-bar-loss:
    backgroundColor: "{colors.down}"
    rounded: "{rounded.none}"
    height: "12px"
  city-chip:
    rounded: "{rounded.none}"
    size: "9px"
  starter-prompt:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.sm}"
    padding: "8px 12px"
  chat-user:
    backgroundColor: "{colors.raise}"
    textColor: "{colors.ink}"
    typography: "{typography.body-long}"
    rounded: "{rounded.sm}"
    padding: "10px 14px"
---

# Design System: 알비온 코치

## Overview

**Creative North Star: "The Night Quote Board"**

The tool sits on a second screen beside a dark, busy game and reads like a Korean brokerage quote screen (시세판). One graphite ground, numerals in warm white, and two market colours used the Korean way: red for gain, blue for loss. Cities stand in a ranked ladder, and each row's profit is drawn as a bar growing left or right from a hairline centre axis. The board is dense but calm. Nothing is boxed, nothing floats, and a refresh changes digits without moving the layout.

Hierarchy comes from one sentence. Each view opens with a Korean sentence that names the decision, with the profit figure set at display size inside it, and the evidence (recipe, return rate, fees, data age) follows in muted text underneath. Below that, the ladder gives the alternatives, and the coach pane takes follow-up questions in Korean. Colour is held back so that the gain/loss pair always reads as data and never as decoration.

The world rejects the category defaults of game-tool UIs: grey panels, a gold accent, card chrome and a plain sortable table. It uses no game logo or official art.

**Key Characteristics:**
- Graphite night ground, warm numeral white, the Korean red-up / blue-down market convention.
- 1px lines carry all structure (dim hairlines divide, a brighter edge outlines controls); no cards, no elevation shadows.
- A single variable typeface (Pretendard Variable) with tabular numerals everywhere numbers appear.
- An answer sentence with an embedded display-size figure leads every result.
- A centre-axis profit ladder with fixed furniture in every row: tincture chip, name, prices, signed profit, bar, data age.
- Amber appears only for the moment a price changes.

## Colors

A near-neutral graphite system with one semantic pair (gain red, loss blue), one transient signal (amber), and seven city tinctures confined to tiny chips.

### Primary
- **Gain Red** (up): fills the profit bar of a city where refining makes money, and its colour rides on the ▲ triangle. It is a fill colour; it is not used for text.
- **Gain Red, Lifted** (up-ink): the text form of gain, used for the signed profit in the ladder, the display figure in the answer sentence, and the route line. It is lifted in lightness so text clears 4.5:1 on the ground (6.3:1) where the fill colour sits at 4.4:1.

### Secondary
- **Loss Blue** (down): fills the bar of a city where refining loses money, extending left of the axis.
- **Loss Blue, Lifted** (down-ink): the text form of loss, used for signed losses and the loss figure in the answer sentence (7.4:1 on the ground).

### Tertiary
- **Refresh Amber** (amber): appears only as a 22% wash behind a price cell whose value changed on refresh, fading to transparent. Never a static fill, border or text colour.

### Neutral
- **Night Graphite** (ground): the single page background for both panes, the top bar and the sticky ask form; also the text colour on Numeral White fills.
- **Raised Graphite** (raise): the one tonal step up, reserved for places the player types into or wrote: inputs, selects, the question textarea, and the player's own chat bubble.
- **Numeral White** (ink): all primary text and numerals (14.5:1 on the ground), the fill of the selected segment and the send button, the selected tab underline, the hover border of ghost buttons and starter prompts, the coach error line's top rule, the caret, the focus ring and text selection.
- **Ledger Grey** (mute): secondary text (evidence line, column heads, data age, labels, unselected tabs and segments, missing values, the textarea placeholder) at 6.5:1 on the ground; also the centre axis line.
- **Hairline Graphite** (rule): every 1px divider (under the top bar, between the panes, under the answer block, under every ladder row, above the ask form) and the scrollbar thumb. It is a separator tone (1.3:1) and never outlines a control.
- **Control Edge** (edge): the 1px border of every control: the segmented control and its internal dividers, selects, the fee input, the question textarea, ghost buttons and starter prompts. It holds 3.5:1 against the ground and 3.2:1 against Raised Graphite, so a control's outline stays visible at arm's length.

### City Tinctures
Heraldic colours for the seven royal cities, shown only as a 9px square chip before the city name: **Thetford Purpure** (city-thetford), **Martlock Azure** (city-martlock), **Lymhurst Vert** (city-lymhurst), **Bridgewatch Or** (city-bridgewatch), **Fort Sterling Argent** (city-fort-sterling), **Caerleon Gules** (city-caerleon), **Brecilien Teal** (city-brecilien).

### Named Rules
**The Market Convention Rule.** Gain Red means profit and Loss Blue means loss, as on a Korean brokerage screen, and they mean nothing else. Each always arrives with its drawn triangle (▲ gain, ▼ loss); zero stays in Numeral White with no triangle. Errors speak in Numeral White, text and rule alike.

**The Lifted Ink Rule.** Bars take the fill pair; numbers and words take the lifted inks. Never set text in the fill red or fill blue.

**The Edge-Or-Rule Rule.** A line that separates content is Hairline Graphite; a line that outlines something the player can press or type into is Control Edge. Never swap them.

**The Fresh Amber Rule.** Amber exists only in motion: a 22% mix of Refresh Amber washed behind a value that just changed, fading out over 2.4s. If amber is visible on a screen at rest, it is a bug.

**The Tincture Chip Rule.** A city tincture lives only in the 9px chip and always sits next to the city name. It is never a text colour, bar colour or row background, and it never identifies a city on its own (Thetford, Martlock and Caerleon sit below 3:1 on the ground).

## Typography

**Display Font:** Pretendard Variable (self-hosted variable woff2, weights 45–920), falling back to system-ui, Apple SD Gothic Neo, Noto Sans KR
**Body Font:** Pretendard Variable (same family)
**Label/Mono Font:** none; numbers use the same family with tabular numerals

**Character:** One neutral Korean-Latin grotesque does everything. The quote-board voice comes from size and weight steps, tight tracking at large sizes, and fixed-width digits, not from a second face.

### Hierarchy
- **Display** (700, 56px, 1.35, -0.03em, tabular): the profit figure inside the answer sentence, never on its own line as a stat tile. Coloured by the Lifted Ink pair and preceded by its triangle. Scales fluidly with the viewport (5vw) between Display Phone and Display.
- **Display Phone** (700, 40px): the floor of the display figure, reached at viewports narrower than about 800px, so every phone sees this step.
- **Headline** (600, 30px, 1.35, -0.015em, balanced wrapping): the answer sentence itself, and the loading, empty and error sentences in the same slot. Scales fluidly (2.6vw) between Headline Phone and Headline.
- **Headline Phone** (600, 22px): the floor of the answer sentence, reached below about 850px.
- **Title** (700, 17px, -0.01em): the app name in the top bar. The only title-scale text.
- **Body** (400, 15px, 1.55): controls, tab labels, ladder cells, route line, starter prompts. Evidence and route text hold a 72ch measure.
- **Body Long** (400, 15px, 1.65): coach chat messages, capped at 65ch.
- **Numeric** (400, 15px, tabular): all price cells; the signed profit cell takes weight 600 and the best city's name takes 700.
- **Meta** (400, 13px): server label, filter field labels, data age column.
- **Label** (500, 12px): ladder column heads. The same 12px size, at 400 or 600, carries the phone detail sub-line, the 정제 보너스 tag and the source note.

### Named Rules
**The Fixed-Digit Rule.** Every number in the ladder and the answer uses tabular numerals, and numeric columns hold minimum widths (7ch for profit, 8ch for prices), so a refresh changes digits and never shifts a column.

**The One Family Rule.** Pretendard Variable is the only face. No monospace for numbers, no display face for headings; weight and size do the work.

**The Whole-Word Rule.** Korean text breaks between words, never inside one (keep-all), with anywhere-wrapping only as an overflow escape for long English item names.

## Layout

The page is a fixed shell: a 56px top bar with a 1px rule under it, and below it panes that fill the remaining dynamic viewport height and scroll independently.

- **1100px and wider:** two columns. The flexible left column (28px side gutters, 20px top) holds one board at a time (무역 루트, 정제 계산기 or 자료), chosen by the top-bar tabs; the coach takes a fixed 440px right column, and a 1px rule divides them. The 코치 tab is hidden because the coach is always visible.
- **Below 1100px:** one pane at a time. Four tabs (무역 루트 / 정제 계산기 / 자료 / 코치) in the top bar, the selected one underlined. At 760px and narrower the labels shorten to 무역 / 정제 / 자료 / 코치, and under 480px the "Asia 서버" label leaves the bar.
- **760px and narrower (phone):** 16px side gutters (respecting safe-area insets), profit bars shrink to 40px per side, the three price columns and the age column leave the ladder and fold into a 12px detail sub-line under each city name (raw × count + lower → product · age). Every select, input, ghost button, checkbox row and starter prompt is at least 44px tall, segment options grow to 11px vertical padding, and checkboxes grow to 20px.

Calculator rhythm, top to bottom: a wrapping filter row (10px row gap, 14px column gap, refresh pushed to the far right on wide screens), 24px space, the answer block (reserved minimum height of 112px so loading does not shift the ladder, 20px bottom padding, 1px rule), the ladder, then route and source notes 18px below. Coach rhythm: messages in a column with 16px gaps and 24px side gutters, starter prompts stacked left with 8px gaps, and the ask form pinned to the bottom of the pane with a 1px top rule.

## Elevation & Depth

The board is flat. There are no elevation shadows anywhere. Depth comes from two things only: 1px Hairline Graphite rules that divide bar, panes, answer, and every ladder row; and a single tonal step (Raised Graphite) under things the player types into or wrote. The sticky ask form repaints Night Graphite so scrolled messages pass cleanly beneath it, without a shadow.

The one box-shadow in the build is structural: the selected tab's 2px inset underline in Numeral White. It marks state, not height.

### Named Rules
**The Flat Board Rule.** Surfaces never lift. If something needs separating, give it a 1px rule or move it onto Raised Graphite; never a shadow, glow or blur.

## Shapes

Square by default. The segmented control, city chips, profit bars and their track have no rounding. Controls the player types into or presses (inputs, selects, the question textarea, ghost and primary buttons, starter prompts) and the player's own chat bubble take a barely-there 2px corner. Nothing is rounder.

Lines are 1px everywhere: dividers (Hairline Graphite), control borders and the segmented control's internal separators (Control Edge), the centre axis (Ledger Grey), the focus outline (offset 2px; inset 3px on segment options). The direction brief asked for one device pixel; the build draws 1 CSS pixel. The selected tab underline at 2px is the only heavier stroke.

The centre axis is a 1px Ledger Grey vertical line that runs 7px above and below the 12px bar track, so the axes of adjacent rows read as one broken spine down the ladder.

### Named Rules
**The Two-Pixel Ceiling Rule.** Corners are 0px or 2px. Chips, bars and the segmented control stay square; a radius above 2px does not belong in this world.

## Components

### Buttons
Quiet and rectangular; the press colour is Numeral White, never a hue.
- **Shape:** gently squared corners (2px).
- **Primary (보내기):** Numeral White fill, Night Graphite text, weight 600, 8px × 18px padding. Disabled at 50% opacity (also when the server has no API key).
- **Ghost (시세 새로고침, 대화 지우기):** transparent with a 1px Control Edge border, Ledger Grey text, 6px × 12px padding. Hover lifts both text and border to Numeral White. Disabled at 45% opacity (refresh is disabled while a load is in flight). 대화 지우기 asks for a second press within 3 seconds, changing its label to 한 번 더 누르면 지워집니다.
- **Focus:** 1px Numeral White outline, 2px offset, on every interactive element.

### Segmented Control (resource picker)
- **Style:** a square strip with a 1px Control Edge border and 1px Control Edge internal dividers; options in Ledger Grey with 7px × 13px padding.
- **State:** the selected option inverts to a Numeral White fill with Night Graphite text at weight 600. The native radio stays in place, invisible, as the hit area; keyboard focus draws a 1px inset outline.

### Inputs / Fields
- **Style:** Raised Graphite fill, 1px Control Edge border, 2px corners, Numeral White value text; 6px × 8px padding for selects and the fee input (fixed 7ch wide), 10px × 12px for the question textarea. Field labels sit inline to the left in 13px Ledger Grey; the textarea placeholder is Ledger Grey.
- **Checkboxes:** native, tinted Numeral White, 15px (20px on phone), label to the right.
- **Disabled:** the enchant select disables for tiers below 4 and for stone; the textarea fades to 60% when the coach is unavailable, with a placeholder that says why.
- **Keyboard:** Enter sends, Shift+Enter adds a line, and IME composition is never interrupted.

### Navigation
- **Top bar:** app name (Title) at left, tabs next to it, the "Asia 서버" label in 13px Ledger Grey pushed to the right. A 1px rule closes the bar.
- **Tabs:** full bar height (56px), 12px side padding, Ledger Grey when idle, Numeral White with the 2px inset underline when selected. Shown only below 1100px; the coach tab is addressable as #coach.

### Answer Sentence (signature)
The decision, in one Korean sentence, at Headline scale: "Thetford에서 정제하면 개당 ▲80 은화". The figure sits inline at Display scale in Gain Red, Lifted (or Loss Blue, Lifted with ▼ when every city loses), never wrapped onto its own line. A Ledger Grey evidence line follows (recipe · return rate · sales tax · station fee · margin), and when the best price is older than six hours it ends by telling the player to check in game. The same slot holds the loading sentence (Ledger Grey), the empty sentence and the error sentence (Numeral White, with the cause and the recovery step underneath).

### City Ladder (signature)
A quote ladder, one city per row, ranked by profit with unpriced cities last.
- **Fixed furniture per row, same slot every refresh:** tincture chip (9px square) and city name, an optional 12px 정제 보너스 tag, the three prices, the signed profit with its triangle, the bar, the data age.
- **Rows:** 13px × 10px cell padding, right-aligned numbers, first column flush left, a 1px rule under every row. The best city's name is weight 700. A city without prices goes fully Ledger Grey and shows 데이터 없음 instead of a number; missing prices show an em dash, never zero.
- **Profit bars:** 12px tall, square-ended, growing right of the axis in Gain Red or left of it in Loss Blue, with a 1px gap at the axis. Length is relative to the largest absolute profit in view, up to 120px per side (40px on phone). When the recipe changes, bars grow from the axis over 0.7s on an exponential ease-out (cubic-bezier(.16, 1, .3, 1)).
- **Refresh tick:** a price cell whose value changed since the last load gets a 22% Refresh Amber wash, fading over 2.4s on the same curve.
- **Reduced motion:** both the bar growth and the tick are removed.

### Trade Route Board
The same quote-board grammar as the calculator, applied to hauling from a start city.
- **Answer sentence:** "Martlock에서 T7 보강 처리한 피혁을 사서 Caerleon에 팔면 한 번에 ▲1,164,210 은화", the trip profit at Display scale; "한 번에" never breaks from itself. The evidence line gives the load (kg and units, or the budget cap), profit per unit, silver needed and the sales fee, then the same stale/average caution as the calculator.
- **Filters:** start city, mount select (each transport ox labelled with its estimated load), the editable 실을 무게 field with a Ledger Grey "kg" suffix (editing it switches the mount to 직접 입력), an optional 자금 field (12ch, placeholder 제한 없음), enchant, premium, auto refresh, refresh.
- **Destination ladder:** one row per other royal city, ranked by trip profit: tincture chip and city name (a text button that picks the city for the tier table, underlined in Ledger Grey on hover; the picked row carries a 12px 티어별 보기 tag), the best item, buy price, sell price, signed trip profit, centre-axis bar, data age. Phone folds item, prices and age into the detail sub-line, each segment kept whole.
- **Tier table:** a 15px/600 heading ("Martlock → Caerleon, 티어별로 옮기면"), then T4–T8 rows with the best item of that tier, profit per unit, units that fit (자금 한도 tag when the budget caps it), silver needed, signed trip profit, bar, age. A tier without prices reads 데이터 없음.

### Notes List
Facts the coach saved, newest first, in a 72ch column on the ground; no cards.
- **Answer slot:** "코치가 모은 자료 N건" with the newest topic and its age underneath; the empty state explains how notes appear; errors speak in Numeral White.
- **Note row:** topic (15px/600) with its age at the right in 13px Ledger Grey ("· 다시 확인 예정" after 30 days), the body rendered like a coach reply, then a 13px Ledger Grey source line of links beside a ghost 지우기 button. A 1px Hairline rule closes each row.
- **Delete:** two presses within 3 seconds (한 번 더 누르면 지워집니다), like 대화 지우기; a failure adds a Numeral White line under the row.

### Coach Thread
- **Player message:** right-aligned Raised Graphite bubble, 2px corners, 10px × 14px padding, preserving line breaks.
- **Coach reply:** unboxed Body Long text on the ground, with paragraphs, lists and bold rendered from its markdown.
- **Waiting:** a single Ledger Grey line (시세와 가이드를 확인하는 중…).
- **Error:** Numeral White text under a 1px Numeral White top rule, naming the cause and telling the player the question is still in the field.
- **Starter prompts:** shown only while the thread is empty; a 40ch Ledger Grey lead line, then 2px-corner buttons outlined in Control Edge with Numeral White text, stacked and left-aligned, whose border lifts to Numeral White on hover. At least 44px tall on phone.

### Named Rules
**The Fixed Furniture Rule.** Every city row carries the same parts in the same slots, and data age is always on the row (its own column on wide screens, the end of the detail sub-line on phone). A price is never shown without its age.

**The Axis Rule.** Profit is drawn from one shared 1px centre axis: gain right in red, loss left in blue, never stacked, never from the left edge.

## Do's and Don'ts

### Do:
- **Do** lead every result with one Korean sentence that names the answer, the figure inline at Display scale with its drawn triangle.
- **Do** pair every signed number with the authored SVG triangle (▲ gain, ▼ loss) and its Lifted Ink colour; leave zero in Numeral White with no triangle.
- **Do** set every number in tabular numerals and hold numeric column widths so a refresh changes digits, never layout.
- **Do** show data age next to every price, in the same slot on every row.
- **Do** draw every divider, border, axis and focus outline at 1px; keep the selected tab's 2px underline as the only heavier line.
- **Do** outline controls in Control Edge and divide content with Hairline Graphite; hover lifts a control's border to Numeral White.
- **Do** show a missing value as an em dash or 데이터 없음 in Ledger Grey, never as zero.
- **Do** theme the browser surfaces from the palette: selection inverts to Night Graphite on Numeral White, caret and focus ring in Numeral White, thin scrollbars in Hairline Graphite, link underlines in Ledger Grey at a 3px offset.
- **Do** give every phone control at least 44px of height at 760px and below.
- **Do** remove the bar growth and the amber tick under prefers-reduced-motion.

### Don't:
- **Don't** use the market red or blue for anything except gain and loss; errors speak in Numeral White.
- **Don't** leave amber on screen at rest; it exists only as the fading wash on a value that just changed.
- **Don't** wrap content in cards, panels or drop shadows; panes are separated by 1px rules on one graphite ground.
- **Don't** round anything beyond 2px, and keep chips, bars and the segmented control square.
- **Don't** add a second typeface or a monospace; hierarchy comes from Pretendard Variable's size and weight steps.
- **Don't** label anything twice: a city is named once per row, and the bar column's header exists only for screen readers.
- **Don't** use a city tincture as a text colour, bar colour or row background.
- **Don't** use Albion Online's logo or official art.
