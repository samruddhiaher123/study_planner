const API = "/api";
const token = () => localStorage.getItem("studyPlannerToken");

async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (token()) headers.Authorization = `Bearer ${token()}`;
  const response = await fetch(`${API}${path}`, { ...options, headers });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || "Something went wrong");
  return data;
}

function showMessage(element, message, success = false) {
  if (!element) return;
  element.textContent = message;
  element.classList.toggle("success-message", success);
}

function requireLogin() {
  if (!token()) window.location.href = "/login.html";
}

document.querySelectorAll("[data-logout]").forEach((button) => {
  button.addEventListener("click", () => {
    localStorage.removeItem("studyPlannerToken");
    window.location.href = "/login.html";
  });
});

const loginForm = document.querySelector("#login-form");
if (loginForm) {
  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(loginForm);
    const message = document.querySelector("#form-message");
    try {
      const result = await request("/login", {
        method: "POST",
        body: JSON.stringify({ email: form.get("email"), password: form.get("password") }),
      });
      localStorage.setItem("studyPlannerToken", result.token);
      window.location.href = "/dashboard.html";
    } catch (error) {
      showMessage(message, error.message);
    }
  });
}

const registerForm = document.querySelector("#register-form");
if (registerForm) {
  registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(registerForm);
    const message = document.querySelector("#form-message");
    try {
      const result = await request("/register", {
        method: "POST",
        body: JSON.stringify({
          name: form.get("name"),
          email: form.get("email"),
          password: form.get("password"),
        }),
      });
      showMessage(message, `${result.message}. You can now sign in.`, true);
      registerForm.reset();
    } catch (error) {
      showMessage(message, error.message);
    }
  });
}

function formatStudyTime(minutes) {
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (hours === 0) return `${remainingMinutes} min`;
  if (remainingMinutes === 0) return `${hours} hr`;
  return `${hours} hr ${remainingMinutes} min`;
}

async function loadDashboard() {
  requireLogin();
  try {
    const dashboard = await request("/dashboard");
    document.querySelector("#today-progress").textContent = `${dashboard.todayProgress}%`;
    document.querySelector("#today-progress-bar").style.width = `${dashboard.todayProgress}%`;
    document.querySelector("#total-tasks").textContent = dashboard.totalTasks;
    document.querySelector("#completed-tasks").textContent = dashboard.completed;
    document.querySelector("#pending-tasks").textContent = dashboard.pending;
    document.querySelector("#study-time").textContent = formatStudyTime(dashboard.studyMinutes);

    const chart = document.querySelector("#weekly-chart");
    chart.replaceChildren();
    dashboard.week.forEach((day) => {
      const item = document.createElement("div");
      item.className = "chart-day";
      const track = document.createElement("div");
      track.className = "chart-track";
      track.title = `${day.completed} of ${day.total} tasks completed`;
      const bar = document.createElement("div");
      bar.className = "chart-bar";
      bar.style.height = `${day.progress}%`;
      track.append(bar);
      const label = document.createElement("span");
      label.className = "chart-label";
      label.textContent = day.day.slice(0, 3);
      item.append(track, label);
      chart.append(item);
    });
  } catch (error) {
    if (error.message === "Missing or invalid token") {
      localStorage.removeItem("studyPlannerToken");
      window.location.href = "/login.html";
    } else {
      window.alert(error.message);
    }
  }
}

const taskForm = document.querySelector("#task-form");
if (taskForm) {
  requireLogin();
  const taskList = document.querySelector("#task-list");
  const taskMessage = document.querySelector("#task-message");
  const cancelEdit = document.querySelector("#cancel-edit");
  let editingID = null;

  function resetTaskForm() {
    editingID = null;
    taskForm.reset();
    document.querySelector("#task-form-title").textContent = "Add a task";
    document.querySelector("#save-task").textContent = "Add task";
    cancelEdit.classList.add("hidden");
  }

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (character) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    })[character]);
  }

  function renderTasks(tasks) {
    taskList.replaceChildren();
    if (tasks.length === 0) {
      taskList.innerHTML = '<div class="empty-state">No tasks yet. Add your first study session above.</div>';
      return;
    }

    tasks.forEach((task) => {
      const card = document.createElement("article");
      card.className = "task-card";
      const statusClass = task.status === "Completed" ? "status completed" : "status";
      card.innerHTML = `
        <div>
          <h3>${escapeHTML(task.title)} <span class="${statusClass}">${escapeHTML(task.status)}</span></h3>
          <p>${escapeHTML(task.subject)} · ${escapeHTML(task.date)} · ${escapeHTML(task.startTime)}–${escapeHTML(task.endTime)}</p>
          <p class="task-description">${escapeHTML(task.description || "")}</p>
        </div>
        <div class="task-actions">
          ${task.status === "Pending" ? `<button class="small-button" data-action="complete" data-id="${task.id}">Complete</button>` : ""}
          <button class="small-button secondary-button" data-action="edit" data-id="${task.id}">Edit</button>
          <button class="small-button delete-button" data-action="delete" data-id="${task.id}">Delete</button>
        </div>`;
      taskList.append(card);
    });
  }

  async function loadTasks() {
    try {
      renderTasks(await request("/tasks"));
    } catch (error) {
      showMessage(taskMessage, error.message);
    }
  }

  taskForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(taskForm);
    const body = Object.fromEntries(form.entries());
    try {
      await request(editingID ? `/tasks/${editingID}` : "/tasks", {
        method: editingID ? "PUT" : "POST",
        body: JSON.stringify(body),
      });
      showMessage(taskMessage, editingID ? "Task updated." : "Task added.", true);
      resetTaskForm();
      await loadTasks();
    } catch (error) {
      showMessage(taskMessage, error.message);
    }
  });

  cancelEdit.addEventListener("click", resetTaskForm);

  taskList.addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action]");
    if (!button) return;
    const id = button.dataset.id;
    try {
      if (button.dataset.action === "delete") {
        if (!window.confirm("Delete this task?")) return;
        await request(`/tasks/${id}`, { method: "DELETE" });
        await loadTasks();
      } else if (button.dataset.action === "complete") {
        await request(`/tasks/${id}/complete`, { method: "PUT" });
        await loadTasks();
      } else if (button.dataset.action === "edit") {
        const task = await request(`/tasks/${id}`);
        for (const [key, value] of Object.entries(task)) {
          const field = taskForm.elements.namedItem(key);
          if (field) field.value = value;
        }
        editingID = id;
        document.querySelector("#task-form-title").textContent = "Edit task";
        document.querySelector("#save-task").textContent = "Save changes";
        cancelEdit.classList.remove("hidden");
        window.scrollTo({ top: 0, behavior: "smooth" });
      }
    } catch (error) {
      showMessage(taskMessage, error.message);
    }
  });

  loadTasks();
}

if (document.body.dataset.page === "dashboard") loadDashboard();
