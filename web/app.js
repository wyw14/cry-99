async function loadJSON(path, options) {
  const response = await fetch(path, options);
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || `Request failed: ${response.status}`);
  return data;
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>'"]/g, ch => ({"&":"&amp;", "<":"&lt;", ">":"&gt;", "'":"&#39;", '"':"&quot;"}[ch]));
}

function phaseTag(phase) {
  const good = ["isolated", "grounded", "ready_to_energize", "energized"].includes(phase);
  const bad = ["failed", "cancelled"].includes(phase);
  return `<span class="tag ${good ? "good" : bad ? "bad" : "warn"}">${escapeHTML(phase)}</span>`;
}

function setError(target, error) {
  document.querySelector(target).innerHTML = `<div class="empty">${escapeHTML(error.message)}</div>`;
}

async function refreshPermits() {
  try {
    const data = await loadJSON("/api/permits");
    const rows = data.permits.map(item => `<tr><td>${escapeHTML(item.id)}</td><td>${escapeHTML(item.worksite)}</td><td>${escapeHTML(item.owner)}</td><td>${phaseTag(item.phase)}</td><td>${escapeHTML(item.topology_revision)}</td></tr>`).join("");
    document.querySelector("#permitRows").innerHTML = rows || `<tr><td colspan="5" class="empty">No permits</td></tr>`;
    document.querySelector("#permitCount").textContent = data.permits.length;
  } catch (error) { setError("#permitTable", error); }
}

async function refreshIsolation() {
  try {
    const data = await loadJSON("/api/isolation");
    const rows = data.plans.map(item => `<tr><td>${escapeHTML(item.id)}</td><td>${escapeHTML(item.permit_id)}</td><td>${item.current} / ${item.steps.length}</td><td>${item.cancelled ? '<span class="tag bad">cancelled</span>' : '<span class="tag good">active</span>'}</td></tr>`).join("");
    document.querySelector("#isolationRows").innerHTML = rows || `<tr><td colspan="4" class="empty">No isolation plans</td></tr>`;
    document.querySelector("#isolationCount").textContent = data.plans.length;
  } catch (error) { setError("#isolationTable", error); }
}

async function refreshEnergize() {
  try {
    const data = await loadJSON("/api/energize");
    const rows = data.readiness.map(item => `<tr><td>${escapeHTML(item.area_id)}</td><td>${escapeHTML(item.revision)}</td><td>${item.ready ? '<span class="tag good">ready</span>' : '<span class="tag warn">blocked</span>'}</td><td>${escapeHTML(item.reason || "All interlocks satisfied")}</td></tr>`).join("");
    document.querySelector("#readinessRows").innerHTML = rows;
    document.querySelector("#commandCount").textContent = data.commands.length;
  } catch (error) { setError("#readinessTable", error); }
}

async function refreshEvents() {
  try {
    const data = await loadJSON("/api/events");
    document.querySelector("#eventRevision").textContent = data.revision;
    document.querySelector("#eventLog").textContent = JSON.stringify(data.events, null, 2);
  } catch (error) { setError("#eventLog", error); }
}

document.addEventListener("DOMContentLoaded", () => {
  const page = document.body.dataset.page;
  const actions = {permits: refreshPermits, isolation: refreshIsolation, energize: refreshEnergize, events: refreshEvents};
  const refresh = actions[page];
  if (refresh) { refresh(); document.querySelector("#refresh")?.addEventListener("click", refresh); }
});
