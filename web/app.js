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

/* ---------- tabs (below 1100px one pane at a time) ---------- */
const tabs = [...document.querySelectorAll('[role="tab"]')];
function showTab(tab) {
  for (const t of tabs) {
    const on = t === tab;
    t.setAttribute('aria-selected', String(on));
    document.getElementById(t.getAttribute('aria-controls')).classList.toggle('is-off', !on);
  }
}
tabs.forEach((t) => t.addEventListener('click', () => {
  showTab(t);
  window.history.replaceState(null, '', t.id === 'tab-coach' ? '#coach' : location.pathname);
}));
showTab(location.hash === '#coach' ? tabs[1] : tabs[0]);

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

const AUTO_MS = 5 * 60 * 1000; // matches the server's price cache
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
    const tint = CITY[r.city] || '#888';
    const a = ago(r.updated);
    const p = pick(r);
    const vals = { raw: r.rawPrice, lower: r.lowerPrice, product: r.productPrice };
    for (const k in vals) cells[r.city + k] = vals[k];
    const avg = { raw: r.avgRaw, lower: r.avgLower, product: r.avgProduct };
    const flash = (k) => (!recipeChanged && last.cells[r.city + k] !== undefined && last.cells[r.city + k] !== vals[k] ? ' fresh' : '') + (avg[k] ? ' avg' : '');
    const price = (v) => (v ? fmt(v) : '—');
    const w = r.missing ? 0 : Math.abs(p) / max;
    return `<tr class="${r.missing ? 'missing' : ''}${r === best ? ' best' : ''}">
      <td><div class="city"><i class="chip" style="background:${tint}"></i>${name}${r.bonus ? '<span class="tag">정제 보너스</span>' : ''}${r.avgRaw || r.avgLower || r.avgProduct ? '<span class="tag avg-tag">평균가</span>' : ''}</div>
        <div class="detail">${r.rawPrice ? `${fmt(r.rawPrice)}×${d.raw.count}` : '—'}${d.lower ? ' + ' + price(r.lowerPrice) : ''} → ${price(r.productPrice)}<span class="nw"> · ${a.text}</span></div></td>
      <td class="px${flash('raw')}">${price(r.rawPrice)}</td>
      <td class="px${flash('lower')}">${d.lower ? price(r.lowerPrice) : '—'}</td>
      <td class="px${flash('product')}">${price(r.productPrice)}</td>
      <td class="pl">${r.missing ? '<span class="nodata">데이터 없음</span>' : signed(p)}</td>
      <td class="axis"><div class="track"><i class="fill ${p >= 0 ? 'gain' : 'loss'}" style="--w:calc(var(--bar-max) * ${w.toFixed(3)})"></i></div></td>
      <td class="age">${a.text}</td>
    </tr>`;
  }).join('');
  last = { key: key.replace('&fresh=1', ''), cells };

  if (recipeChanged) {
    const tb = $('#rows');
    tb.classList.add('grow');
    requestAnimationFrame(() => requestAnimationFrame(() => tb.classList.remove('grow')));
  }

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
const FKEY = 'albion-coach.filters.v1';
try {
  const saved = JSON.parse(localStorage.getItem(FKEY)) || {};
  for (const [k, v] of Object.entries(saved)) {
    const el = form.elements[k];
    if (!el) continue;
    if (el instanceof RadioNodeList) el.value = v;
    else if (el.type === 'checkbox') el.checked = v;
    else el.value = v;
  }
} catch { /* storage blocked: start from defaults */ }
form.addEventListener('change', () => {
  const f = form.elements;
  const v = { resource: f.resource.value, tier: f.tier.value, enchant: f.enchant.value, fee: f.fee.value, premium: f.premium.checked, focus: f.focus.checked, auto: f.auto.checked };
  try { localStorage.setItem(FKEY, JSON.stringify(v)); } catch { /* not persisted */ }
});
let autoTimer = null;
function scheduleAuto() {
  clearInterval(autoTimer);
  autoTimer = form.auto.checked
    ? setInterval(() => { if (document.visibilityState === 'visible') loadRefine(true, true); }, AUTO_MS)
    : null;
}
// A phone that slept past a refresh catches up as soon as it is looked at again.
document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible' && form.auto.checked && Date.now() - lastLoadAt >= AUTO_MS) loadRefine(true, true);
});
scheduleAuto();

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
loadRefine(false);

/* ---------- coach ---------- */
const KEY = 'albion-coach.chat.v1';
const store = {
  get() { try { return JSON.parse(localStorage.getItem(KEY)) || []; } catch { return []; } },
  set(v) { try { localStorage.setItem(KEY, JSON.stringify(v)); } catch { /* private mode: keep in memory only */ } },
};
let chat = store.get();

// Minimal markdown: escape first, then only add our own tags (bold, lists, paragraphs).
function md(src) {
  const lines = esc(src).replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>').split('\n');
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
    const r = await fetch('/api/chat', {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ messages: msgs }),
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
    $('.starters-lead').textContent = '코치는 Claude API 키가 있어야 열려요. 맥의 albion-coach 폴더에 .env 파일을 만들어 ANTHROPIC_API_KEY=sk-ant-... 한 줄을 넣고 scripts/rolling-update.sh를 실행하면 됩니다. 정제 계산기는 키 없이 동작합니다.';
    document.querySelectorAll('.starter').forEach((b) => { b.disabled = true; });
    $('#send').disabled = true;
    $('#q').disabled = true;
    $('#q').placeholder = '서버에 API 키가 없어 지금은 질문할 수 없습니다';
  }
}).catch(() => {});
drawLog();
