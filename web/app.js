'use strict';
const $ = (s) => document.querySelector(s);

// The Korean game client keeps English city names, so the tool does too.
const CITY = {
  Thetford: '#6b3f8e', Martlock: '#2c5aa0', Lymhurst: '#3e7d3a', Bridgewatch: '#d98a1f',
  'Fort Sterling': '#e6e6e1', Caerleon: '#9e2a2b', Brecilien: '#3a8c8c',
};
const RES = {
  ore: ['광석', '금속괴'], wood: ['목재', '판자'], fiber: ['섬유', '천'], hide: ['생가죽', '가죽'], rock: ['돌', '석재'],
};
const TRI_UP = '<svg class="tri" viewBox="0 0 10 9" aria-hidden="true"><path d="M5 0 10 9H0z"/></svg>';
const TRI_DN = '<svg class="tri" viewBox="0 0 10 9" aria-hidden="true"><path d="M0 0h10L5 9z"/></svg>';

const fmt = (n) => Math.round(n).toLocaleString('ko-KR');
const cityName = (c) => c;
const esc = (s) => String(s).replace(/[&<>"']/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch]));
const signed = (p) => (p > 0 ? `<span class="up">${TRI_UP}${fmt(p)}</span>` : p < 0 ? `<span class="down">${TRI_DN}${fmt(-p)}</span>` : '0');

function ago(iso) {
  const t = Date.parse(iso);
  if (!(t > Date.UTC(2000, 0, 1))) return { text: '없음', stale: true };
  const m = Math.max(0, Math.round((Date.now() - t) / 60000));
  const text = m < 1 ? '방금' : m < 60 ? `${m}분 전` : m < 1440 ? `${Math.round(m / 60)}시간 전` : `${Math.round(m / 1440)}일 전`;
  return { text, stale: m > 360 };
}

// Korean object particle: 을 after a final consonant, 를 after a vowel.
const obj = (w) => {
  const c = w.charCodeAt(w.length - 1) - 0xac00;
  return w + (c < 0 || c > 11171 ? '을(를)' : c % 28 ? '을' : '를');
};

/* ---------- tabs: boards share the left column next to the coach; below 1100px one pane at a time ---------- */
const tabs = [...document.querySelectorAll('[role="tab"]')];
const wide = matchMedia('(min-width: 1100px)');
const HASH = { 'tab-calc': '#calc', 'tab-notes': '#notes', 'tab-coach': '#coach' };
const onShow = {}; // tab id -> runs whenever that pane comes into view
let current = 'tab-trade';
let board = 'tab-trade'; // the board beside the coach on wide screens
function showTab(id) {
  current = id;
  if (id !== 'tab-coach') board = id;
  const selected = wide.matches && id === 'tab-coach' ? board : id;
  for (const t of tabs) {
    t.setAttribute('aria-selected', String(t.id === selected));
    const pane = document.getElementById(t.getAttribute('aria-controls'));
    pane.classList.toggle('is-off', t.id !== id);
    pane.classList.toggle('is-left-off', t.id !== board);
  }
  onShow[wide.matches ? board : id]?.();
}
tabs.forEach((t) => t.addEventListener('click', () => {
  showTab(t.id);
  window.history.replaceState(null, '', HASH[t.id] || location.pathname);
}));
wide.addEventListener('change', () => showTab(current));

const AUTO_MS = 5 * 60 * 1000; // matches the server's price cache

// Remembers a board's filters in this browser.
function persist(formEl, key) {
  try {
    for (const [k, v] of Object.entries(JSON.parse(localStorage.getItem(key)) || {})) {
      const el = formEl.elements[k];
      if (!el) continue;
      if (el.type === 'checkbox') el.checked = v;
      else el.value = v;
    }
  } catch { /* storage blocked: start from defaults */ }
  formEl.addEventListener('change', () => {
    const v = {};
    for (const el of formEl.elements) {
      if (!el.name || (el.type === 'radio' && !el.checked)) continue;
      v[el.name] = el.type === 'checkbox' ? el.checked : el.value;
    }
    try { localStorage.setItem(key, JSON.stringify(v)); } catch { /* not persisted */ }
  });
}

// Five-minute refresh for a board with an `auto` checkbox. A phone that slept past a refresh
// catches up as soon as it is looked at again. Returns the function that re-reads the checkbox.
function autoRefresh(formEl, load, lastAt) {
  let timer = null;
  const schedule = () => {
    clearInterval(timer);
    timer = formEl.auto.checked ? setInterval(() => { if (document.visibilityState === 'visible') load(); }, AUTO_MS) : null;
  };
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible' && formEl.auto.checked && Date.now() - lastAt() >= AUTO_MS) load();
  });
  schedule();
  return schedule;
}

const priceText = (v) => (v ? fmt(v) : '—');
const bar = (p, w) => `<td class="axis"><div class="track"><i class="fill ${p >= 0 ? 'gain' : 'loss'}" style="--w:calc(var(--bar-max) * ${w.toFixed(3)})"></i></div></td>`;
function grow(tbody) {
  tbody.classList.add('grow');
  requestAnimationFrame(() => requestAnimationFrame(() => tbody.classList.remove('grow')));
}

/* ---------- calculator ---------- */
const form = $('#filters');
let last = { key: '', cells: {} };

function query() {
  const f = new FormData(form);
  return new URLSearchParams({
    resource: f.get('resource'), tier: f.get('tier'), enchant: f.get('enchant'),
    premium: f.get('premium') ? '1' : '0', fee: String(Math.max(0, Math.min(5000, Number(f.get('fee')) || 0))),
  });
}

function label(item) {
  if (!item) return '';
  const res = RES[item.resource];
  const name = item.name || (item.refined ? res[1] : res[0]);
  return `T${item.tier}${item.enchant ? '.' + item.enchant : ''} ${name}`;
}

let lastLoadAt = 0;

// quiet = background refresh: keep what is on screen, report failures in one line.
async function loadRefine(force, quiet) {
  const q = query();
  const noEnchant = Number(q.get('tier')) < 4 || q.get('resource') === 'rock';
  form.enchant.disabled = noEnchant;
  if (noEnchant) q.set('enchant', '0');
  if (force) q.set('fresh', '1');
  const answer = $('#answer');
  if (!quiet) {
    answer.className = 'answer loading';
    answer.innerHTML = '<h1>시세를 불러오는 중…</h1>';
  }
  $('#refresh').disabled = true;
  try {
    const r = await fetch('/api/refine?' + q);
    const data = await r.json();
    if (!r.ok) throw new Error(data.error || `HTTP ${r.status}`);
    lastData = data;
    lastLoadAt = Date.now();
    $('#autonote').textContent = '';
    render(data, q.toString());
  } catch (e) {
    if (quiet && lastData) {
      const next = new Date(Date.now() + AUTO_MS).toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' });
      $('#autonote').textContent = `자동 새로고침 실패: ${e.message}. ${next}에 다시 시도합니다.`;
      return;
    }
    answer.className = 'answer error';
    answer.innerHTML = `<h1>시세를 가져오지 못했습니다</h1><p>${esc(e.message)}. 잠시 뒤 '시세 새로고침'을 눌러 다시 시도하세요.</p>`;
    $('#rows').innerHTML = '';
    $('#route').textContent = '';
  } finally {
    $('#refresh').disabled = false;
  }
}

function render(d, key) {
  const focus = form.focus.checked;
  const pick = (row) => (focus ? row.profitFocus : row.profit);
  const recipeChanged = key.replace('&fresh=1', '') !== last.key;

  $('#h-raw').textContent = `${label(d.raw)} ×${d.raw.count}`;
  $('#h-lower').textContent = d.lower ? label(d.lower) : '하위 정제품';
  $('#h-product').textContent = label(d.product);

  const rows = [...d.rows].sort((a, b) => (a.missing - b.missing) || pick(b) - pick(a));
  const priced = rows.filter((r) => !r.missing);
  const best = priced[0];
  const max = Math.max(1, ...priced.map((r) => Math.abs(pick(r))));
  const fee = (d.sellFee * 100).toFixed(1).replace(/\.0$/, '');
  const station = d.feePer100 ? `제작소 사용료 ${fmt(d.feePer100)}/영양100 반영` : '제작소 사용료 빠짐';
  const recipe = `${label(d.raw)} ${d.raw.count}개${d.lower ? ` + ${label(d.lower)} 1개` : ''} → ${label(d.product)}`;

  const answer = $('#answer');
  answer.className = 'answer';
  if (!best) {
    answer.innerHTML = `<h1>이 조합은 Asia 서버에서 최근 7일 거래 기록이 부족합니다</h1><p>${recipe}. 표에 있는 가격만으로는 계산이 안 돼요. 다른 티어를 보거나, 게임에서 Albion Data Client를 켜고 시장을 열면 데이터가 채워집니다.</p>`;
  } else if (pick(best) > 0) {
    const rrr = ((focus ? best.rrrFocus : best.rrr) * 100).toFixed(1);
    const age = ago(best.updated);
    const usesAvg = best.avgRaw || best.avgLower || best.avgProduct;
    const caution = usesAvg
      ? ' · 회색 가격은 최근 하루 평균 거래가라 실제 주문가와 다를 수 있어요'
      : age.stale ? ` · 시세가 ${age.text} 값이라 게임에서 한 번 확인하세요` : '';
    answer.innerHTML = `<h1>${cityName(best.city)}에서 정제하면 개당 <span class="fig up">${TRI_UP}${fmt(pick(best))}</span> 은화</h1>
      <p>${recipe} · 리턴율 ${rrr}%${focus ? ' (포커스)' : ''} · 판매 수수료 ${fee}% · ${station} · 마진 ${Math.round(best[focus ? 'marginFocus' : 'margin'] * 100)}%${caution}</p>`;
  } else {
    answer.innerHTML = `<h1>지금은 어디서 정제해도 개당 <span class="fig down">${TRI_DN}${fmt(-pick(best))}</span> 손해입니다</h1>
      <p>${recipe}. 원자재를 그대로 파는 편이 낫습니다${focus ? '' : '. 포커스를 쓰면 결과가 달라질 수 있습니다'}.</p>`;
  }

  const cells = {};
  $('#rows').innerHTML = rows.map((r) => {
    const name = r.city;
    const tint = CITY[r.city] || 'var(--mute)';
    const a = ago(r.updated);
    const p = pick(r);
    const vals = { raw: r.rawPrice, lower: r.lowerPrice, product: r.productPrice };
    for (const k in vals) cells[r.city + k] = vals[k];
    const avg = { raw: r.avgRaw, lower: r.avgLower, product: r.avgProduct };
    const flash = (k) => (!recipeChanged && last.cells[r.city + k] !== undefined && last.cells[r.city + k] !== vals[k] ? ' fresh' : '') + (avg[k] ? ' avg' : '');
    const price = priceText;
    const w = r.missing ? 0 : Math.abs(p) / max;
    return `<tr class="${r.missing ? 'missing' : ''}${r === best ? ' best' : ''}">
      <td><div class="city"><i class="chip" style="background:${tint}"></i>${name}${r.bonus ? '<span class="tag">정제 보너스</span>' : ''}${r.avgRaw || r.avgLower || r.avgProduct ? '<span class="tag avg-tag">평균가</span>' : ''}</div>
        <div class="detail">${r.rawPrice ? `${fmt(r.rawPrice)}×${d.raw.count}` : '—'}${d.lower ? ' + ' + price(r.lowerPrice) : ''} → ${price(r.productPrice)}<span class="nw"> · ${a.text}</span></div></td>
      <td class="px${flash('raw')}">${price(r.rawPrice)}</td>
      <td class="px${flash('lower')}">${d.lower ? price(r.lowerPrice) : '—'}</td>
      <td class="px${flash('product')}">${price(r.productPrice)}</td>
      <td class="pl">${r.missing ? '<span class="nodata">데이터 없음</span>' : signed(p)}</td>
      ${bar(p, w)}
      <td class="age">${a.text}</td>
    </tr>`;
  }).join('');
  last = { key: key.replace('&fresh=1', ''), cells };

  if (recipeChanged) grow($('#rows'));

  const route = d.route && (focus ? d.route.profitFocus : d.route.profit);
  const rt = d.route;
  const where = rt && (rt.buy === rt.refine
    ? `${rt.buy}에서 원자재를 사서 정제하고`
    : `${rt.buy}에서 원자재를 사서 ${rt.refine}에서 정제하고`);
  $('#route').innerHTML = route && best && route > pick(best)
    ? `운반까지 하면: ${where} ${rt.sell}에서 팔면 개당 ${signed(route)} 은화. 시세 ${ago(rt.updated).text} 기준이고 이동 위험은 빠져 있습니다.`
    : '';
}

let lastData = null;
persist(form, 'albion-coach.filters.v1');
const scheduleAuto = autoRefresh(form, () => loadRefine(true, true), () => lastLoadAt);

form.addEventListener('change', (e) => {
  if (e.target.name === 'auto') {
    scheduleAuto();
    return;
  }
  if (e.target.name === 'focus' && lastData) {
    render(lastData, last.key);
    return;
  }
  loadRefine(false);
});
form.addEventListener('submit', (e) => e.preventDefault());
$('#refresh').addEventListener('click', () => loadRefine(true));
onShow['tab-calc'] = () => { if (!lastLoadAt && !$('#refresh').disabled) loadRefine(false); };

/* ---------- trade route ---------- */
// Transport ox load per tier, estimated from game data as mount weight × HALFMOUNTED maxload
// (items.xml + spells.xml, ao-bin-dumps 2026-11-26). Not checked against the in-game tooltip,
// so the weight field stays editable and the source note says so.
const OX_KG = { 3: 413, 4: 517, 5: 689, 6: 954, 7: 1349, 8: 1935 };
const tform = $('#tfilters');
for (const o of tform.mount.options) if (OX_KG[o.value]) o.textContent += ` · 약 ${fmt(OX_KG[o.value])}kg`;
persist(tform, 'albion-coach.trade.v1');
let trade = { key: '', data: null, at: 0, dest: '', cells: {} };

function tquery() {
  const f = tform.elements;
  return new URLSearchParams({
    from: f.from.value,
    enchant: f.enchant.value,
    kg: String(Math.min(20000, Math.max(1, Math.round(Number(f.kg.value) || OX_KG[4])))),
    budget: String(Math.max(0, Math.floor(Number(f.budget.value) || 0))),
    premium: f.premium.checked ? '1' : '0',
  });
}

async function loadTrade(force, quiet) {
  const q = tquery();
  if (force) q.set('fresh', '1');
  const answer = $('#tanswer');
  if (!quiet) {
    answer.className = 'answer loading';
    answer.innerHTML = '<h1>시세를 불러오는 중…</h1>';
  }
  $('#trefresh').disabled = true;
  try {
    const r = await fetch('/api/trade?' + q);
    const data = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(data.error || `HTTP ${r.status}`);
    trade.at = Date.now();
    $('#tautonote').textContent = '';
    renderTrade(data, q.toString().replace('&fresh=1', ''));
  } catch (e) {
    if (quiet && trade.data) {
      const next = new Date(Date.now() + AUTO_MS).toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' });
      $('#tautonote').textContent = `자동 새로고침 실패: ${e.message}. ${next}에 다시 시도합니다.`;
      return;
    }
    answer.className = 'answer error';
    answer.innerHTML = `<h1>시세를 가져오지 못했습니다</h1><p>${esc(e.message)}. 잠시 뒤 '시세 새로고침'을 눌러 다시 시도하세요.</p>`;
    $('#trows').innerHTML = '';
    $('#tiers').innerHTML = '';
    $('#ttitle').textContent = '';
  } finally {
    $('#trefresh').disabled = false;
  }
}

function caution(o) {
  if (o.avgBuy || o.avgSell) return ' · 회색 가격은 최근 하루 평균 거래가라 실제 주문가와 다를 수 있어요';
  const a = ago(o.updated);
  return a.stale ? ` · 시세가 ${a.text} 값이라 게임에서 한 번 확인하세요` : '';
}

function renderTrade(d, key) {
  const changed = key !== trade.key;
  trade = { ...trade, key, data: d };
  const dests = d.destinations;
  const priced = dests.filter((x) => x.best);
  const top = priced[0];
  const fee = (d.sellFee * 100).toFixed(1).replace(/\.0$/, '');

  const answer = $('#tanswer');
  answer.className = 'answer';
  if (!top) {
    answer.innerHTML = `<h1>${d.from} 시세가 부족해서 계산할 수 없습니다</h1><p>Asia 서버에 최근 7일 동안 ${d.from}에서 이 인챈트의 자원·정제품 거래 기록이 거의 없어요. 다른 출발 도시나 인챈트를 보거나, 게임에서 Albion Data Client를 켜고 시장을 열면 데이터가 채워집니다.</p>`;
  } else if (top.best.trip > 0) {
    const o = top.best;
    const load = o.byBudget ? `자금 ${fmt(d.budget)} 한도로 ${fmt(o.units)}개` : `${fmt(d.kg)}kg에 ${fmt(o.units)}개`;
    answer.innerHTML = `<h1>${d.from}에서 ${obj(label(o.item))} 사서 ${top.city}에 팔면 <span class="nw">한 번에</span> <span class="fig up">${TRI_UP}${fmt(o.trip)}</span> 은화</h1>
      <p>${load} · 개당 ${signed(o.unit)} · 필요 자금 ${fmt(o.need)} · 판매 수수료 ${fee}%${caution(o)}</p>`;
  } else {
    const o = top.best;
    answer.innerHTML = `<h1>지금은 ${d.from}에서 어디로 옮겨도 남는 물건이 없습니다</h1>
      <p>그나마 나은 건 ${top.city}로 가는 ${label(o.item)}(개당 ${signed(o.unit)})입니다. 다른 출발 도시나 인챈트를 보세요.</p>`;
  }

  if (!dests.some((x) => x.city === trade.dest && x.best)) trade.dest = top ? top.city : '';
  const max = Math.max(1, ...priced.map((x) => Math.abs(x.best.trip)));
  const cells = {};
  $('#trows').innerHTML = dests.map((x) => {
    const o = x.best;
    const a = o ? ago(o.updated) : { text: '없음' };
    const vals = { buy: o?.buy, sell: o?.sell };
    for (const k in vals) cells[x.city + k] = vals[k];
    const cls = (k, avg) => (!changed && trade.cells[x.city + k] !== undefined && trade.cells[x.city + k] !== vals[k] ? ' fresh' : '') + (avg ? ' avg' : '');
    const sel = x.city === trade.dest;
    return `<tr class="${o ? '' : 'missing'}${x === top ? ' best' : ''}">
      <td><div class="city"><button type="button" class="pick" data-city="${esc(x.city)}" aria-pressed="${sel}"${o ? '' : ' disabled'}><i class="chip" style="background:${CITY[x.city] || 'var(--mute)'}"></i><span class="cname">${x.city}</span></button>${sel ? '<span class="tag">티어별 보기</span>' : ''}</div>
        ${o ? `<div class="detail">${label(o.item)} · <span class="nw">${fmt(o.buy)} → ${fmt(o.sell)}</span><span class="nw"> · ${a.text}</span></div>` : ''}</td>
      <td class="px item">${o ? label(o.item) : '—'}</td>
      <td class="px${cls('buy', o?.avgBuy)}">${priceText(o?.buy)}</td>
      <td class="px${cls('sell', o?.avgSell)}">${priceText(o?.sell)}</td>
      <td class="pl">${o ? signed(o.trip) : '<span class="nodata">데이터 없음</span>'}</td>
      ${bar(o ? o.trip : 0, o ? Math.abs(o.trip) / max : 0)}
      <td class="age">${a.text}</td>
    </tr>`;
  }).join('');
  trade.cells = cells;
  if (changed) grow($('#trows'));
  renderTiers(changed);
}

function renderTiers(changed) {
  const d = trade.data;
  const x = d && d.destinations.find((y) => y.city === trade.dest);
  if (!x) {
    $('#ttitle').textContent = '';
    $('#tiers').innerHTML = '';
    return;
  }
  $('#ttitle').textContent = `${d.from} → ${x.city}, 티어별로 옮기면`;
  const opts = x.tiers.filter(Boolean);
  const max = Math.max(1, ...opts.map((o) => Math.abs(o.trip)));
  $('#tiers').innerHTML = x.tiers.map((o, i) => {
    const tier = `T${i + 4}`;
    if (!o) {
      return `<tr class="missing"><td><div class="city">${tier}</div></td><td class="px">—</td><td class="px">—</td><td class="px">—</td>
        <td class="pl"><span class="nodata">데이터 없음</span></td>${bar(0, 0)}<td class="age">—</td></tr>`;
    }
    const a = ago(o.updated);
    return `<tr class="${o === x.best ? 'best' : ''}">
      <td><div class="city">${label(o.item)}</div>
        <div class="detail"><span class="nw">개당 ${signed(o.unit)}</span> · <span class="nw">${fmt(o.units)}개</span> · <span class="nw">자금 ${fmt(o.need)}</span><span class="nw"> · ${a.text}</span></div></td>
      <td class="px${o.avgBuy || o.avgSell ? ' avg' : ''}">${signed(o.unit)}</td>
      <td class="px">${fmt(o.units)}개${o.byBudget ? '<span class="tag"> 자금 한도</span>' : ''}</td>
      <td class="px">${fmt(o.need)}</td>
      <td class="pl">${signed(o.trip)}</td>
      ${bar(o.trip, Math.abs(o.trip) / max)}
      <td class="age">${a.text}</td>
    </tr>`;
  }).join('');
  if (changed) grow($('#tiers'));
}

$('#trows').addEventListener('click', (e) => {
  const b = e.target.closest('.pick');
  if (!b || b.dataset.city === trade.dest) return;
  trade.dest = b.dataset.city;
  for (const p of document.querySelectorAll('#trows .pick')) p.setAttribute('aria-pressed', String(p === b));
  for (const t of document.querySelectorAll('#trows .city .tag')) t.remove();
  b.insertAdjacentHTML('afterend', '<span class="tag">티어별 보기</span>');
  renderTiers(true);
});

const scheduleTradeAuto = autoRefresh(tform, () => loadTrade(true, true), () => trade.at);
tform.addEventListener('change', (e) => {
  const f = tform.elements;
  if (e.target.name === 'auto') {
    scheduleTradeAuto();
    return;
  }
  if (e.target.name === 'mount' && OX_KG[f.mount.value]) f.kg.value = OX_KG[f.mount.value];
  if (e.target.name === 'kg' && Number(f.kg.value) !== OX_KG[f.mount.value]) f.mount.value = '0';
  loadTrade(false);
});
tform.addEventListener('submit', (e) => e.preventDefault());
$('#trefresh').addEventListener('click', () => loadTrade(true));
onShow['tab-trade'] = () => { if (!trade.at && !$('#trefresh').disabled) loadTrade(false); };

/* ---------- notes ---------- */
let notes = [];

async function loadNotes() {
  const answer = $('#nanswer');
  answer.className = 'answer loading';
  answer.innerHTML = '<h1>자료를 불러오는 중…</h1>';
  try {
    const r = await fetch('/api/notes');
    const d = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(d.error || `HTTP ${r.status}`);
    if (!d.enabled) {
      answer.className = 'answer';
      answer.innerHTML = '<h1>자료 저장소가 꺼져 있습니다</h1><p>서버에 DATABASE_URL이 없어서 코치가 찾은 내용을 저장하지 않습니다. docker compose로 실행하면 켜집니다.</p>';
      $('#nlist').innerHTML = '';
      return;
    }
    notes = d.notes;
    renderNotes();
  } catch (e) {
    answer.className = 'answer error';
    answer.innerHTML = `<h1>자료를 불러오지 못했습니다</h1><p>${esc(e.message)}. 잠시 뒤 자료 탭을 다시 열어 보세요.</p>`;
    $('#nlist').innerHTML = '';
  }
}

function renderNotes() {
  const answer = $('#nanswer');
  answer.className = 'answer';
  answer.innerHTML = notes.length
    ? `<h1>코치가 모은 자료 ${notes.length}건</h1><p>가장 최근: ${esc(notes[0].topic)} · ${ago(notes[0].updated).text}</p>`
    : '<h1>아직 쌓인 자료가 없습니다</h1><p>코치에게 가이드에 없는 것(섬, 길드, 최신 패치 같은 것)을 물어보면, 웹에서 찾은 사실을 출처와 함께 여기에 모읍니다.</p>';
  $('#nlist').innerHTML = notes.map((n) => {
    const days = (Date.now() - Date.parse(n.updated)) / 864e5;
    const links = n.sources.filter((src) => /^https?:\/\//.test(src.url))
      .map((src) => `<a href="${esc(src.url)}" target="_blank" rel="noopener noreferrer">${esc(src.title || src.url)}</a>`).join(' · ');
    return `<li class="note" data-id="${n.id}">
      <div class="note-head"><h3>${esc(n.topic)}</h3><span class="age">${ago(n.updated).text}${days > 30 ? ' · 다시 확인 예정' : ''}</span></div>
      <div class="note-body">${md(n.body)}</div>
      <div class="note-foot"><p class="note-src">출처: ${links || '없음'}</p><button type="button" class="ghost del">지우기</button></div>
    </li>`;
  }).join('');
}

$('#nlist').addEventListener('click', async (e) => {
  const b = e.target.closest('.del');
  if (!b) return;
  const li = b.closest('.note');
  if (!b.dataset.armed) { // two presses, like 대화 지우기
    b.dataset.armed = '1';
    b.textContent = '한 번 더 누르면 지워집니다';
    setTimeout(() => { delete b.dataset.armed; b.textContent = '지우기'; }, 3000);
    return;
  }
  b.disabled = true;
  try {
    const r = await fetch(`/api/notes/${li.dataset.id}`, { method: 'DELETE' });
    if (!r.ok) throw new Error((await r.json().catch(() => ({}))).error || `HTTP ${r.status}`);
    notes = notes.filter((n) => String(n.id) !== li.dataset.id);
    renderNotes();
  } catch (err) {
    b.disabled = false;
    li.querySelector('.err')?.remove();
    li.insertAdjacentHTML('beforeend', `<p class="err">지우지 못했습니다: ${esc(err.message)}. 다시 눌러 보세요.</p>`);
  }
});
onShow['tab-notes'] = loadNotes;

/* ---------- coach ---------- */
const KEY = 'albion-coach.chat.v1';
const store = {
  get() { try { return JSON.parse(localStorage.getItem(KEY)) || []; } catch { return []; } },
  set(v) { try { localStorage.setItem(KEY, JSON.stringify(v)); } catch { /* private mode: keep in memory only */ } },
};
let chat = store.get();

// Minimal markdown: escape first, then only add our own tags (bold, http(s) links, lists, paragraphs).
function md(src) {
  const lines = esc(src)
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\[([^\]\n]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
    .split('\n');
  let html = '';
  let list = null;
  let para = [];
  const flushP = () => { if (para.length) { html += `<p>${para.join('<br>')}</p>`; para = []; } };
  const flushL = () => { if (list) { html += `</${list}>`; list = null; } };
  for (const raw of lines) {
    const l = raw.replace(/^#+ /, '');
    const ul = /^\s*[-*] (.*)/.exec(l);
    const ol = /^\s*(\d+)\. (.*)/.exec(l);
    if (ul || ol) {
      flushP();
      const kind = ul ? 'ul' : 'ol';
      if (list !== kind) { flushL(); html += kind === 'ol' ? `<ol start="${ol[1]}">` : '<ul>'; list = kind; }
      html += `<li>${ul ? ul[1] : ol[2]}</li>`;
    } else if (!l.trim()) {
      flushP(); flushL();
    } else {
      flushL(); para.push(l);
    }
  }
  flushP(); flushL();
  return html;
}

function drawLog(extra) {
  const items = chat.map((m) => (m.role === 'user'
    ? `<li class="me">${esc(m.content)}</li>`
    : `<li class="bot">${md(m.content)}</li>`));
  if (extra) items.push(extra);
  $('#log').innerHTML = items.join('');
  $('#starters').hidden = chat.length > 0 || Boolean(extra);
  const pane = $('#pane-coach');
  pane.scrollTop = pane.scrollHeight;
}

async function ask(text) {
  const send = $('#send');
  send.disabled = true;
  const msgs = [...chat, { role: 'user', content: text }];
  chat = msgs;
  drawLog('<li class="wait">시세와 가이드를 확인하는 중…</li>');
  try {
    // The server takes at most 40 messages; an odd-sized tail still starts with a question.
    const r = await fetch('/api/chat', {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ messages: msgs.slice(-39) }),
    });
    const data = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(data.error || `HTTP ${r.status}`);
    chat = [...msgs, { role: 'assistant', content: data.reply }];
    store.set(chat);
    drawLog();
  } catch (e) {
    chat = msgs.slice(0, -1);
    $('#q').value = text;
    drawLog(`<li class="err">답을 받지 못했습니다: ${esc(e.message)}. 질문은 입력칸에 그대로 두었으니 다시 보내 보세요.</li>`);
  } finally {
    send.disabled = !coachOn;
  }
}

let coachOn = true;
$('#ask').addEventListener('submit', (e) => {
  e.preventDefault();
  const q = $('#q');
  const text = q.value.trim();
  if (!text || $('#send').disabled) return;
  q.value = '';
  ask(text);
});
$('#q').addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) { e.preventDefault(); $('#ask').requestSubmit(); }
});
document.querySelectorAll('.starter').forEach((b) => b.addEventListener('click', () => {
  $('#q').value = b.textContent;
  $('#ask').requestSubmit();
}));
let armed = 0;
$('#clear').addEventListener('click', (e) => {
  if (Date.now() - armed > 3000) {
    armed = Date.now();
    e.target.textContent = '한 번 더 누르면 지워집니다';
    setTimeout(() => { e.target.textContent = '대화 지우기'; }, 3000);
    return;
  }
  chat = [];
  store.set(chat);
  e.target.textContent = '대화 지우기';
  drawLog();
});

fetch('/api/status').then((r) => r.json()).then((s) => {
  coachOn = Boolean(s.coach);
  if (!coachOn) {
    $('.starters-lead').textContent = '코치는 Claude API 키가 있어야 열려요. 맥의 albion-coach 폴더에 .env 파일을 만들어 ANTHROPIC_API_KEY=sk-ant-... 한 줄을 넣고 scripts/rolling-update.sh를 실행하면 됩니다. 무역 루트와 정제 계산기는 키 없이 동작합니다.';
    document.querySelectorAll('.starter').forEach((b) => { b.disabled = true; });
    $('#send').disabled = true;
    $('#q').disabled = true;
    $('#q').placeholder = '서버에 API 키가 없어 지금은 질문할 수 없습니다';
  }
}).catch(() => {});
drawLog();

showTab(Object.keys(HASH).find((k) => HASH[k] === location.hash) || 'tab-trade');
