const sessionSelect = document.querySelector("#session-select");
const scorePanel = document.querySelector("#score-panel");
const historyPanel = document.querySelector("#history-panel");
const scoreForm = document.querySelector("#score-form");
const scoreInputs = document.querySelector("#score-inputs");
const scoreTotal = document.querySelector("#score-total");
const formMessage = document.querySelector("#form-message");
const gamesContainer = document.querySelector("#games");
const dialog = document.querySelector("#session-dialog");
const sessionForm = document.querySelector("#session-form");

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
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || "通信に失敗しました");
  return body;
}

async function loadSessions(selectedID = "") {
  const sessions = await api("/api/sessions");
  sessionSelect.innerHTML = '<option value="">対局日を選択してください</option>';
  for (const session of sessions) {
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
  if (active) await loadGames();
}

async function loadGames() {
  const games = await api(`/api/sessions/${sessionSelect.value}/games`);
  if (games.length === 0) {
    gamesContainer.innerHTML = '<p class="empty-state">まだ半荘が記録されていません。</p>';
    return;
  }
  gamesContainer.innerHTML = games.map((game, gameIndex) => `
    <article class="game-card">
      <h3>${games.length - gameIndex}回戦</h3>
      <table>
        <thead><tr><th>プレイヤー</th><th>スコア（千点）</th></tr></thead>
        <tbody>${game.results.map(result => `
          <tr>
            <td>${escapeHTML(result.playerName)}</td>
            <td>${result.score.toLocaleString()}</td>
          </tr>`).join("")}
        </tbody>
      </table>
    </article>`).join("");
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
    await api(`/api/sessions/${sessionSelect.value}/games`, {
      method: "POST",
      body: JSON.stringify({results}),
    });
    formMessage.textContent = "保存しました。";
    await loadGames();
  } catch (error) {
    formMessage.textContent = error.message;
  }
});

document.querySelector("#new-session-button").addEventListener("click", () => {
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
    const session = await api("/api/sessions", {
      method: "POST",
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

function escapeHTML(value) {
  const element = document.createElement("span");
  element.textContent = value;
  return element.innerHTML;
}

loadSessions().catch(error => {
  document.body.insertAdjacentHTML("beforeend", `<p class="fatal-error">${escapeHTML(error.message)}</p>`);
});
updateTotal();
