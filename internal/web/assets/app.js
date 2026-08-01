const sessionSelect = document.querySelector("#session-select");
const importButton = document.querySelector("#import-button");
const importFile = document.querySelector("#import-file");
const backupMessage = document.querySelector("#backup-message");
const sessionActions = document.querySelector("#session-actions");
const scorePanel = document.querySelector("#score-panel");
const scorePanelMode = document.querySelector("#score-panel-mode");
const scorePanelTitle = document.querySelector("#score-panel-title");
const historyPanel = document.querySelector("#history-panel");
const scoreForm = document.querySelector("#score-form");
const scoreInputs = document.querySelector("#score-inputs");
const scoreTotal = document.querySelector("#score-total");
const formMessage = document.querySelector("#form-message");
const gamesContainer = document.querySelector("#games");
const historyViewButtons = document.querySelectorAll("[data-history-view]");
const dialog = document.querySelector("#session-dialog");
const sessionForm = document.querySelector("#session-form");
const saveGameButton = document.querySelector("#save-game-button");
const cancelGameEditButton = document.querySelector("#cancel-game-edit");
let currentSessions = [];
let currentGames = [];
let editingGameID = null;
let editingSessionID = null;
let playerNamesBeforeEdit = null;
const savedHistoryView = localStorage.getItem("history-view");
let historyView = ["tiles", "table", "chart"].includes(savedHistoryView) ? savedHistoryView : "tiles";

for (let index = 0; index < 4; index += 1) {
  scoreInputs.insertAdjacentHTML("beforeend", `
    <div class="score-entry">
      <label>
        <span>プレイヤー ${index + 1}</span>
        <input name="player-${index}" autocomplete="off" required placeholder="名前">
      </label>
      <label>
        <span>スコア</span>
        <input name="score-${index}" type="number" inputmode="numeric" required value="0" aria-label="プレイヤー ${index + 1}の1000点単位のスコア">
      </label>
    </div>
  `);
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {"Content-Type": "application/json", ...options.headers},
  });
  const body = response.status === 204 ? null : await response.json();
  if (!response.ok) throw new Error(body.error || "通信に失敗しました");
  return body;
}

async function loadSessions(selectedID = "") {
  currentSessions = await api("/api/sessions");
  sessionSelect.innerHTML = '<option value="">対局日を選択してください</option>';
  for (const session of currentSessions) {
    const option = document.createElement("option");
    option.value = session.id;
    option.textContent = `${session.name}（${session.playedAt.slice(0, 10)}）`;
    sessionSelect.append(option);
  }
  if (selectedID) {
    sessionSelect.value = String(selectedID);
    await selectSession();
  }
}

async function selectSession() {
  const active = Boolean(sessionSelect.value);
  scorePanel.hidden = !active;
  historyPanel.hidden = !active;
  sessionActions.hidden = !active;
  cancelGameEdit();
  if (active) await loadGames();
}

async function loadGames() {
  currentGames = await api(`/api/sessions/${sessionSelect.value}/games`);
  renderGames();
}

function renderGames() {
  updateHistoryViewButtons();
  if (currentGames.length === 0) {
    gamesContainer.className = "games";
    gamesContainer.innerHTML = '<p class="empty-state">まだ半荘が記録されていません。</p>';
    return;
  }

  if (historyView === "table") {
    renderGamesTable();
    return;
  }
  if (historyView === "chart") {
    renderGamesChart();
    return;
  }

  gamesContainer.className = "games";
  gamesContainer.innerHTML = renderScoreSummary() + currentGames.map((game, gameIndex) => `
    <article class="game-card">
      <div class="game-card-heading">
        <h3>${gameIndex + 1}回戦</h3>
        ${renderGameTime(game.createdAt)}
      </div>
      <table>
        <thead><tr><th>プレイヤー</th><th>スコア（千点）</th></tr></thead>
        <tbody>${game.results.map(result => `
          <tr>
            <td>${escapeHTML(result.playerName)}</td>
            <td><span class="score-value score-rank-${getScoreRank(game, result.score)}">${result.score.toLocaleString()}</span></td>
          </tr>`).join("")}
        </tbody>
      </table>
      <div class="game-card-actions">
        <button class="text-button" type="button" data-edit-game="${game.id}">編集</button>
        <button class="danger-button" type="button" data-delete-game="${game.id}">削除</button>
      </div>
    </article>`).join("");
}

function getScoreRank(game, score) {
  const distinctScores = [...new Set(game.results.map(result => result.score))]
    .sort((left, right) => right - left);
  const rank = distinctScores.indexOf(score) + 1;
  return rank <= 2 ? rank : 0;
}
function renderScoreSummary() {
  const totals = getPlayerTotals().sort((left, right) => right.total - left.total);
  return `
    <section class="score-summary" aria-labelledby="score-summary-title">
      <div class="score-summary-heading">
        <div>
          <p class="eyebrow">TOTAL SCORE</p>
          <h3 id="score-summary-title">合計得点</h3>
        </div>
        <span>単位：千点</span>
      </div>
      <div class="score-summary-list">
        ${totals.map((player, index) => `
          <article class="score-summary-player">
            <span class="rank-badge">${index + 1}</span>
            <span class="score-summary-name">${escapeHTML(player.name)}</span>
            <strong class="${player.total < 0 ? "negative-score" : ""}">${formatScore(player.total)}</strong>
          </article>`).join("")}
      </div>
    </section>`;
}

function getPlayerTotals() {
  const totals = new Map();
  for (const game of currentGames) {
    for (const result of game.results) {
      totals.set(result.playerName, (totals.get(result.playerName) ?? 0) + result.score);
    }
  }
  return [...totals].map(([name, total]) => ({name, total}));
}

function formatScore(score) {
  return `${score > 0 ? "+" : ""}${score.toLocaleString()}`;
}
function renderGameTime(createdAt) {
  const date = new Date(createdAt);
  if (Number.isNaN(date.getTime())) return '<span class="game-time">—</span>';

  const time = new Intl.DateTimeFormat("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
  const fullDate = new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
  return `<time class="game-time" datetime="${escapeHTML(createdAt)}" title="${escapeHTML(fullDate)}">${time}</time>`;
}
function renderGamesTable() {
  const playerNames = [];
  for (const game of currentGames) {
    for (const result of game.results) {
      if (!playerNames.includes(result.playerName)) playerNames.push(result.playerName);
    }
  }

  gamesContainer.className = "games-table-wrap";
  gamesContainer.innerHTML = `
    <table class="games-table">
      <thead>
        <tr>
          <th scope="col">回戦</th>
          ${playerNames.map(name => `<th scope="col">${escapeHTML(name)}</th>`).join("")}
          <th scope="col">記録時刻</th>
          <th scope="col"><span class="visually-hidden">操作</span></th>
        </tr>
      </thead>
      <tbody>
        ${currentGames.map((game, gameIndex) => {
          const scores = new Map(game.results.map(result => [result.playerName, result.score]));
          return `
            <tr>
              <th scope="row">${gameIndex + 1}回戦</th>
              ${playerNames.map(name => {
                const score = scores.get(name);
                return `<td>${score === undefined ? "—" : `<span class="score-value score-rank-${getScoreRank(game, score)}">${score.toLocaleString()}</span>`}</td>`;
              }).join("")}
              <td>${renderGameTime(game.createdAt)}</td>
              <td class="actions-cell">
                <button class="text-button" type="button" data-edit-game="${game.id}">編集</button>
                <button class="danger-button" type="button" data-delete-game="${game.id}">削除</button>
              </td>
            </tr>`;
        }).join("")}
      </tbody>
      <tfoot>
        <tr>
          <th scope="row">合計得点</th>
          ${playerNames.map(name => {
            const total = currentGames.reduce((sum, game) => {
              const result = game.results.find(item => item.playerName === name);
              return sum + (result?.score ?? 0);
            }, 0);
            return `<td>${formatScore(total)}</td>`;
          }).join("")}
          <td></td>
          <td></td>
        </tr>
      </tfoot>
    </table>
    <p class="table-unit">単位：千点</p>`;
}

function renderGamesChart() {
  const playerNames = [];
  for (const game of currentGames) {
    for (const result of game.results) {
      if (!playerNames.includes(result.playerName)) playerNames.push(result.playerName);
    }
  }

  const totals = new Map(playerNames.map(name => [name, 0]));
  const series = new Map(playerNames.map(name => [name, [0]]));
  for (const game of currentGames) {
    for (const result of game.results) {
      totals.set(result.playerName, totals.get(result.playerName) + result.score);
    }
    for (const name of playerNames) series.get(name).push(totals.get(name));
  }

  const width = 920;
  const height = 430;
  const padding = {top: 28, right: 32, bottom: 54, left: 66};
  const plotWidth = width - padding.left - padding.right;
  const plotHeight = height - padding.top - padding.bottom;
  const allScores = [...series.values()].flat();
  const scoreMin = Math.min(0, ...allScores);
  const scoreMax = Math.max(0, ...allScores);
  const tickStep = getChartTickStep(scoreMax - scoreMin);
  const yMin = Math.floor(scoreMin / tickStep) * tickStep;
  const yMax = Math.ceil(scoreMax / tickStep) * tickStep || tickStep;
  const x = round => padding.left + (round / currentGames.length) * plotWidth;
  const y = score => padding.top + ((yMax - score) / (yMax - yMin)) * plotHeight;
  const colors = ["#17613f", "#d97706", "#2563a8", "#b33b5c", "#7656a8", "#008b8b", "#8a5a2b", "#59645e"];
  const yTicks = [];
  for (let value = yMin; value <= yMax; value += tickStep) yTicks.push(value);

  gamesContainer.className = "score-chart";
  gamesContainer.innerHTML = `
    <div class="score-chart-heading">
      <div>
        <p class="eyebrow">SCORE TREND</p>
        <h3>累積得点の推移</h3>
      </div>
      <span>単位：千点</span>
    </div>
    <div class="score-chart-legend">
      ${playerNames.map((name, index) => `
        <span class="chart-legend-item">
          <i style="--series-color: ${colors[index % colors.length]}"></i>
          ${escapeHTML(name)}
          <strong class="${totals.get(name) < 0 ? "negative-score" : ""}">${formatScore(totals.get(name))}</strong>
        </span>`).join("")}
    </div>
    <div class="score-chart-scroll">
      <svg class="score-chart-svg" viewBox="0 0 ${width} ${height}" role="img" aria-labelledby="score-chart-title score-chart-description">
        <title id="score-chart-title">各プレイヤーの累積得点推移</title>
        <desc id="score-chart-description">開始時点を0として、各回戦終了後の累積得点を折れ線で表示しています。</desc>
        ${yTicks.map(value => `
          <line class="chart-grid-line ${value === 0 ? "chart-zero-line" : ""}" x1="${padding.left}" y1="${y(value)}" x2="${width - padding.right}" y2="${y(value)}"></line>
          <text class="chart-axis-label" x="${padding.left - 12}" y="${y(value) + 4}" text-anchor="end">${formatScore(value)}</text>`).join("")}
        ${Array.from({length: currentGames.length + 1}, (_, round) => `
          <line class="chart-x-tick" x1="${x(round)}" y1="${height - padding.bottom}" x2="${x(round)}" y2="${height - padding.bottom + 6}"></line>
          <text class="chart-axis-label" x="${x(round)}" y="${height - padding.bottom + 24}" text-anchor="middle">${round === 0 ? "開始" : `${round}回`}</text>`).join("")}
        ${playerNames.map((name, index) => {
          const color = colors[index % colors.length];
          const points = series.get(name);
          const path = points.map((score, round) => `${round === 0 ? "M" : "L"} ${x(round)} ${y(score)}`).join(" ");
          return `
            <path class="chart-series-line" d="${path}" stroke="${color}"></path>
            ${points.map((score, round) => `
              <circle class="chart-series-point" cx="${x(round)}" cy="${y(score)}" r="4" fill="${color}">
                <title>${escapeHTML(name)}：${round === 0 ? "開始" : `${round}回戦`} ${formatScore(score)}</title>
              </circle>`).join("")}`;
        }).join("")}
      </svg>
    </div>`;
}

function getChartTickStep(range) {
  if (range <= 0) return 10;
  const roughStep = range / 5;
  const magnitude = 10 ** Math.floor(Math.log10(roughStep));
  const normalized = roughStep / magnitude;
  const nice = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10;
  return nice * magnitude;
}

function updateHistoryViewButtons() {
  for (const button of historyViewButtons) {
    const active = button.dataset.historyView === historyView;
    button.classList.toggle("is-active", active);
    button.setAttribute("aria-pressed", String(active));
  }
}

for (const button of historyViewButtons) {
  button.addEventListener("click", () => {
    historyView = button.dataset.historyView;
    localStorage.setItem("history-view", historyView);
    renderGames();
  });
}
function updateTotal() {
  const total = [...scoreForm.querySelectorAll('input[name^="score-"]')]
    .reduce((sum, input) => sum + Number(input.value || 0), 0);
  scoreTotal.textContent = `合計 ${total.toLocaleString()}（千点）`;
  scoreTotal.classList.toggle("warning", total !== 0);
}

scoreForm.addEventListener("input", updateTotal);
scoreForm.addEventListener("submit", async event => {
  event.preventDefault();
  formMessage.textContent = "";
  const wasEditing = editingGameID !== null;
  const results = Array.from({length: 4}, (_, index) => ({
    playerName: scoreForm.elements[`player-${index}`].value,
    score: Number(scoreForm.elements[`score-${index}`].value),
  }));
  try {
    const path = editingGameID
      ? `/api/sessions/${sessionSelect.value}/games/${editingGameID}`
      : `/api/sessions/${sessionSelect.value}/games`;
    await api(path, {
      method: editingGameID ? "PUT" : "POST",
      body: JSON.stringify({results}),
    });
    if (wasEditing) cancelGameEdit();
    formMessage.textContent = wasEditing ? "更新しました。" : "保存しました。";
    await loadGames();
  } catch (error) {
    formMessage.textContent = error.message;
  }
});

function startGameEdit(gameID) {
  const game = currentGames.find(item => item.id === gameID);
  if (!game) return;
  if (editingGameID === null) {
    playerNamesBeforeEdit = Array.from({length: 4}, (_, index) =>
      scoreForm.elements[`player-${index}`].value
    );
  }
  editingGameID = gameID;
  scorePanel.classList.add("is-editing");
  scorePanelMode.textContent = "EDITING";
  scorePanelTitle.textContent = "半荘結果を編集中";
  game.results.forEach((result, index) => {
    scoreForm.elements[`player-${index}`].value = result.playerName;
    scoreForm.elements[`score-${index}`].value = result.score;
  });
  saveGameButton.textContent = "変更を保存";
  cancelGameEditButton.hidden = false;
  formMessage.textContent = "";
  updateTotal();
  scorePanel.scrollIntoView({behavior: "smooth", block: "start"});
}

function cancelGameEdit(clearInputs = true) {
  const namesToRestore = editingGameID !== null ? playerNamesBeforeEdit : null;
  editingGameID = null;
  playerNamesBeforeEdit = null;
  scorePanel.classList.remove("is-editing");
  scorePanelMode.textContent = "NEW GAME";
  scorePanelTitle.textContent = "半荘結果を入力";
  saveGameButton.textContent = "半荘を保存";
  cancelGameEditButton.hidden = true;
  if (clearInputs) {
    for (let index = 0; index < 4; index += 1) {
      scoreForm.elements[`player-${index}`].value = namesToRestore?.[index] ?? "";
      scoreForm.elements[`score-${index}`].value = "0";
    }
    formMessage.textContent = "";
    updateTotal();
  }
}

cancelGameEditButton.addEventListener("click", () => cancelGameEdit());
gamesContainer.addEventListener("click", async event => {
  const editButton = event.target.closest("[data-edit-game]");
  if (editButton) {
    startGameEdit(Number(editButton.dataset.editGame));
    return;
  }
  const deleteButton = event.target.closest("[data-delete-game]");
  if (!deleteButton) return;
  const gameID = Number(deleteButton.dataset.deleteGame);
  if (!confirm("この半荘結果を削除しますか？")) return;
  try {
    await api(`/api/sessions/${sessionSelect.value}/games/${gameID}`, {method: "DELETE"});
    if (editingGameID === gameID) cancelGameEdit();
    await loadGames();
  } catch (error) {
    formMessage.textContent = error.message;
  }
});

document.querySelector("#new-session-button").addEventListener("click", () => {
  editingSessionID = null;
  sessionForm.reset();
  document.querySelector("#session-dialog-title").textContent = "対局日を作成";
  document.querySelector("#save-session-button").textContent = "作成する";
  document.querySelector("#session-date").valueAsDate = new Date();
  dialog.showModal();
});

importButton.addEventListener("click", () => {
  importFile.value = "";
  importFile.click();
});

importFile.addEventListener("change", async () => {
  const file = importFile.files[0];
  if (!file) return;
  backupMessage.hidden = false;
  backupMessage.classList.remove("is-error");
  backupMessage.textContent = "インポートしています…";
  importButton.disabled = true;
  try {
    const result = await api("/api/import", {
      method: "POST",
      body: file,
    });
    backupMessage.textContent =
      `対局日 ${result.sessions}件、半荘結果 ${result.games}件をインポートしました。`;
    await loadSessions(sessionSelect.value);
  } catch (error) {
    backupMessage.classList.add("is-error");
    backupMessage.textContent = error.message;
  } finally {
    importButton.disabled = false;
  }
});
document.querySelector("#close-dialog").addEventListener("click", () => dialog.close());
sessionSelect.addEventListener("change", selectSession);

sessionForm.addEventListener("submit", async event => {
  event.preventDefault();
  const message = document.querySelector("#session-message");
  message.textContent = "";
  try {
    const session = await api(editingSessionID ? `/api/sessions/${editingSessionID}` : "/api/sessions", {
      method: editingSessionID ? "PUT" : "POST",
      body: JSON.stringify({
        name: document.querySelector("#session-name").value,
        playedAt: document.querySelector("#session-date").value,
      }),
    });
    sessionForm.reset();
    dialog.close();
    await loadSessions(session.id);
  } catch (error) {
    message.textContent = error.message;
  }
});

document.querySelector("#edit-session-button").addEventListener("click", () => {
  const session = currentSessions.find(item => item.id === Number(sessionSelect.value));
  if (!session) return;
  editingSessionID = session.id;
  document.querySelector("#session-name").value = session.name;
  document.querySelector("#session-date").value = session.playedAt.slice(0, 10);
  document.querySelector("#session-dialog-title").textContent = "対局日を編集";
  document.querySelector("#save-session-button").textContent = "変更を保存";
  document.querySelector("#session-message").textContent = "";
  dialog.showModal();
});

document.querySelector("#delete-session-button").addEventListener("click", async () => {
  const session = currentSessions.find(item => item.id === Number(sessionSelect.value));
  if (!session || !confirm(`「${session.name}」と、その半荘結果をすべて削除しますか？`)) return;
  try {
    await api(`/api/sessions/${session.id}`, {method: "DELETE"});
    sessionSelect.value = "";
    scorePanel.hidden = true;
    historyPanel.hidden = true;
    sessionActions.hidden = true;
    await loadSessions();
  } catch (error) {
    alert(error.message);
  }
});

function escapeHTML(value) {
  const element = document.createElement("span");
  element.textContent = value;
  return element.innerHTML;
}

loadSessions().catch(error => {
  document.body.insertAdjacentHTML("beforeend", `<p class="fatal-error">${escapeHTML(error.message)}</p>`);
});
updateTotal();
