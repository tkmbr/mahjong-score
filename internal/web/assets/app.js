const sessionSelect = document.querySelector("#session-select");
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
let historyView = localStorage.getItem("history-view") === "table" ? "table" : "tiles";

for (let index = 0; index < 4; index += 1) {
  scoreInputs.insertAdjacentHTML("beforeend", `
    <tr>
      <td><input name="player-${index}" autocomplete="off" required placeholder="プレイヤー ${index + 1}"></td>
      <td><input name="score-${index}" type="number" inputmode="numeric" required value="0" aria-label="1000点単位のスコア"></td>
    </tr>
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

  gamesContainer.className = "games";
  gamesContainer.innerHTML = renderScoreSummary() + [...currentGames].reverse().map((game, gameIndex) => `
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
  for (const game of [...currentGames].reverse()) {
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
          <th scope="col">記録時刻</th>
          ${playerNames.map(name => `<th scope="col">${escapeHTML(name)}</th>`).join("")}
          <th scope="col"><span class="visually-hidden">操作</span></th>
        </tr>
      </thead>
      <tbody>
        ${[...currentGames].reverse().map((game, gameIndex) => {
          const scores = new Map(game.results.map(result => [result.playerName, result.score]));
          return `
            <tr>
              <th scope="row">${gameIndex + 1}回戦</th>
              <td>${renderGameTime(game.createdAt)}</td>
              ${playerNames.map(name => {
                const score = scores.get(name);
                return `<td>${score === undefined ? "—" : `<span class="score-value score-rank-${getScoreRank(game, score)}">${score.toLocaleString()}</span>`}</td>`;
              }).join("")}
              <td class="actions-cell">
                <button class="text-button" type="button" data-edit-game="${game.id}">編集</button>
                <button class="danger-button" type="button" data-delete-game="${game.id}">削除</button>
              </td>
            </tr>`;
        }).join("")}
      </tbody>
      <tfoot>
        <tr>
          <th scope="row" colspan="2">合計得点</th>
          ${playerNames.map(name => {
            const total = currentGames.reduce((sum, game) => {
              const result = game.results.find(item => item.playerName === name);
              return sum + (result?.score ?? 0);
            }, 0);
            return `<td>${formatScore(total)}</td>`;
          }).join("")}
          <td></td>
        </tr>
      </tfoot>
    </table>
    <p class="table-unit">単位：千点</p>`;
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
    formMessage.textContent = editingGameID ? "更新しました。" : "保存しました。";
    cancelGameEdit(false);
    await loadGames();
  } catch (error) {
    formMessage.textContent = error.message;
  }
});

function startGameEdit(gameID) {
  const game = currentGames.find(item => item.id === gameID);
  if (!game) return;
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
  editingGameID = null;
  scorePanel.classList.remove("is-editing");
  scorePanelMode.textContent = "NEW GAME";
  scorePanelTitle.textContent = "半荘結果を入力";
  saveGameButton.textContent = "半荘を保存";
  cancelGameEditButton.hidden = true;
  if (clearInputs) {
    for (let index = 0; index < 4; index += 1) {
      scoreForm.elements[`player-${index}`].value = "";
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
