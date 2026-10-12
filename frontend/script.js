const API = "/api";
const token = () => localStorage.getItem("studyPlannerToken");
let serverDate = "";

async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (token()) headers.Authorization = `Bearer ${token()}`;
  const response = await fetch(`${API}${path}`, { ...options, headers });
  serverDate = response.headers.get("X-Server-Date") || serverDate;
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
  const message = document.querySelector("#dashboard-message");
  try {
    const dashboard = await request("/dashboard");
    serverDate = dashboard.todayDate || serverDate;
    document.querySelector("#today-progress").textContent = `${dashboard.todayProgress}%`;
    document.querySelector("#today-progress-bar").style.width = `${dashboard.todayProgress}%`;
    document.querySelector("#today-breakdown").textContent =
      `${dashboard.todayCompleted} completed · ${dashboard.todayRemaining} remaining of ${dashboard.todayPlanned} planned`;
    document.querySelector("#total-tasks").textContent = dashboard.totalTasks;
    document.querySelector("#completed-tasks").textContent = dashboard.completed;
    document.querySelector("#pending-tasks").textContent = dashboard.pending;
    document.querySelector("#not-started-tasks").textContent = dashboard.notStarted;
    document.querySelector("#in-progress-tasks").textContent = dashboard.inProgress;
    document.querySelector("#on-hold-tasks").textContent = dashboard.onHold;
    document.querySelector("#study-time").textContent = formatStudyTime(dashboard.studyMinutes);
    document.querySelector("#current-streak").textContent = dashboard.currentStreak;
    document.querySelector("#longest-streak").textContent = dashboard.longestStreak;
    document.querySelector("#last-study-date").textContent = dashboard.lastStudyDate || "—";
    document.querySelector("#weekly-comparison").textContent = dashboard.previousWeekAvailable
      ? `This week ${dashboard.weeklyProgress}% · last week ${dashboard.previousWeekProgress}% (${dashboard.weekChange >= 0 ? "+" : ""}${dashboard.weekChange} pts)`
      : `This week ${dashboard.weeklyProgress}% · not enough data to compare last week`;

    const chart = document.querySelector("#weekly-chart");
    chart.replaceChildren();
    dashboard.week.forEach((day) => {
      const item = document.createElement("div");
      item.className = "chart-day";
      const track = document.createElement("div");
      track.className = "chart-track";
      track.title = `${day.completed} of ${day.total} tasks completed · ${formatStudyTime(day.studyMinutes)} recorded`;
      const bar = document.createElement("div");
      bar.className = "chart-bar";
      bar.style.height = `${day.progress}%`;
      track.append(bar);
      const details = document.createElement("span");
      details.className = "chart-details";
      details.textContent = formatStudyTime(day.studyMinutes);
      const label = document.createElement("span");
      label.className = "chart-label";
      label.textContent = day.day.slice(0, 3);
      item.append(track, details, label);
      chart.append(item);
    });
  } catch (error) {
    if (error.message === "Missing or invalid token") {
      localStorage.removeItem("studyPlannerToken");
      window.location.href = "/login.html";
    } else {
      showMessage(message, error.message);
    }
  }
}

const taskForm = document.querySelector("#task-form");
if (taskForm) {
  requireLogin();
  const taskList = document.querySelector("#task-list");
  const taskMessage = document.querySelector("#task-message");
  const cancelEdit = document.querySelector("#cancel-edit");
  const filters = document.querySelector("#task-filters");
  const subjectFilter = filters.elements.namedItem("subject");
  const groupToggle = filters.elements.namedItem("groupBySubject");
  const calendar = document.querySelector("#calendar-grid");
  const calendarMessage = document.querySelector("#calendar-message");
  let editingID = null;
  let subjectsLoaded = false;
  let calendarMonth = new Date(new Date().getFullYear(), new Date().getMonth(), 1);
  let searchDelay;

  function resetTaskForm() {
    editingID = null;
    taskForm.reset();
    taskForm.elements.namedItem("actualStudyMinutes").value = "0";
    document.querySelector("#task-form-title").textContent = "Add a task";
    document.querySelector("#save-task").textContent = "Add task";
    cancelEdit.classList.add("hidden");
  }

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (character) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    })[character]);
  }

  function todayString() {
    if (serverDate) return serverDate;
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  }

  function statusClass(status) {
    return `status status-${status.toLowerCase().replaceAll(" ", "-")}`;
  }

  function isOverdue(task) {
    return Boolean(task.date) && task.status !== "Completed" && task.date < todayString();
  }

  function makeTaskCard(task) {
    const card = document.createElement("article");
    card.className = `task-card${isOverdue(task) ? " task-overdue" : ""}`;
    card.innerHTML = `
      <div>
        <h3>${escapeHTML(task.title)}
          <span class="${statusClass(task.status)}">${escapeHTML(task.status)}</span>
          <span class="${statusClass(`priority-${task.priority.toLowerCase()}`)}">${escapeHTML(task.priority)} priority</span>
          ${isOverdue(task) ? '<span class="overdue-label">Overdue</span>' : ""}
        </h3>
        <p>${escapeHTML(task.subject)} · ${escapeHTML(task.date)} · ${escapeHTML(task.startTime)}–${escapeHTML(task.endTime)}</p>
        <p class="task-description">${escapeHTML(task.description || "")}</p>
        ${task.actualStudyMinutes ? `<p class="task-description">Recorded study: ${task.actualStudyMinutes} min</p>` : ""}
      </div>
      <div class="task-actions">
        <label class="inline-status">Status<select data-action="status" data-id="${task.id}" aria-label="Update status for ${escapeHTML(task.title)}">
          ${["Not Started", "In Progress", "Completed", "On Hold"].map((status) => `<option${task.status === status ? " selected" : ""}>${status}</option>`).join("")}
        </select></label>
        <button class="small-button secondary-button" data-action="edit" data-id="${task.id}">Edit</button>
        <button class="small-button delete-button" data-action="delete" data-id="${task.id}">Delete</button>
      </div>`;
    return card;
  }

  function renderTasks(tasks) {
    taskList.replaceChildren();
    if (tasks.length === 0) {
      taskList.innerHTML = '<div class="empty-state">No tasks match these filters.</div>';
      return;
    }
    if (!groupToggle.checked) {
      tasks.forEach((task) => taskList.append(makeTaskCard(task)));
      return;
    }
    const groups = new Map();
    tasks.forEach((task) => {
      if (!groups.has(task.subject)) groups.set(task.subject, []);
      groups.get(task.subject).push(task);
    });
    groups.forEach((subjectTasks, subject) => {
      const group = document.createElement("section");
      group.className = "subject-group";
      const heading = document.createElement("h3");
      heading.textContent = subject;
      group.append(heading);
      subjectTasks.forEach((task) => group.append(makeTaskCard(task)));
      taskList.append(group);
    });
  }

  function filterQuery() {
    const data = new FormData(filters);
    const params = new URLSearchParams();
    ["q", "subject", "status", "priority", "date", "from", "to", "view", "sort"].forEach((key) => {
      const value = String(data.get(key) || "").trim();
      if (value && !(key === "sort" && value === "default")) params.set(key, value);
    });
    return params.toString();
  }

  async function populateSubjects(tasks) {
    if (subjectsLoaded) return;
    const selected = subjectFilter.value;
    const subjects = [...new Set(tasks.map((task) => task.subject))].sort((a, b) => a.localeCompare(b));
    subjectFilter.replaceChildren(new Option("All subjects", ""));
    subjects.forEach((subject) => subjectFilter.add(new Option(subject, subject)));
    subjectFilter.value = selected;
    subjectsLoaded = true;
  }

  async function loadTasks() {
    taskList.innerHTML = '<div class="empty-state">Loading tasks…</div>';
    try {
      const query = filterQuery();
      const tasks = await request(`/tasks${query ? `?${query}` : ""}`);
      renderTasks(tasks);
      if (!subjectsLoaded) {
        const allTasks = query ? await request("/tasks") : tasks;
        await populateSubjects(allTasks);
      }
    } catch (error) {
      showMessage(taskMessage, error.message);
      taskList.innerHTML = '<div class="empty-state">Tasks could not be loaded.</div>';
    }
  }

  async function loadCalendar() {
    const year = calendarMonth.getFullYear();
    const month = calendarMonth.getMonth();
    document.querySelector("#calendar-month").textContent =
      calendarMonth.toLocaleDateString(undefined, { month: "long", year: "numeric" });
    calendar.innerHTML = '<div class="empty-state">Loading calendar…</div>';
    showMessage(calendarMessage, "");
    const firstDay = new Date(year, month, 1);
    const mondayOffset = (firstDay.getDay() + 6) % 7;
    const start = new Date(year, month, 1 - mondayOffset);
    const end = new Date(start.getFullYear(), start.getMonth(), start.getDate() + 42);
    const isoDate = (date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
    try {
      const tasks = await request(`/tasks?from=${isoDate(start)}&to=${isoDate(new Date(end.getFullYear(), end.getMonth(), end.getDate() - 1))}`);
      const tasksByDate = new Map();
      tasks.forEach((task) => {
        if (!tasksByDate.has(task.date)) tasksByDate.set(task.date, []);
        tasksByDate.get(task.date).push(task);
      });
      calendar.replaceChildren();
      ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].forEach((day) => {
        const heading = document.createElement("div");
        heading.className = "calendar-weekday";
        heading.textContent = day;
        calendar.append(heading);
      });
      for (let index = 0; index < 42; index++) {
        const date = new Date(start.getFullYear(), start.getMonth(), start.getDate() + index);
        const dateKey = isoDate(date);
        const cell = document.createElement("div");
        cell.className = `calendar-cell${date.getMonth() !== month ? " other-month" : ""}${dateKey === todayString() ? " calendar-today" : ""}`;
        cell.dataset.date = dateKey;
        cell.innerHTML = `<span class="calendar-date">${date.getDate()}</span>`;
        (tasksByDate.get(dateKey) || []).forEach((task) => {
          const button = document.createElement("button");
          button.type = "button";
          button.className = `calendar-task ${statusClass(task.status)}${isOverdue(task) ? " calendar-overdue" : ""}`;
          button.dataset.action = "edit";
          button.dataset.id = task.id;
          button.textContent = `${task.startTime} ${task.title}`;
          button.title = `${task.subject} · ${task.status}${isOverdue(task) ? " · Overdue" : ""}`;
          cell.append(button);
        });
        calendar.append(cell);
      }
      const monthPrefix = `${year}-${String(month + 1).padStart(2, "0")}-`;
      if (!tasks.some((task) => task.date.startsWith(monthPrefix))) {
        showMessage(calendarMessage, "No tasks are scheduled this month.");
      }
    } catch (error) {
      calendar.innerHTML = '<div class="empty-state">Calendar could not be loaded.</div>';
      showMessage(calendarMessage, error.message);
    }
  }

  async function editTask(id) {
    const task = await request(`/tasks/${id}`);
    for (const [key, value] of Object.entries(task)) {
      const field = taskForm.elements.namedItem(key);
      if (field) field.value = value ?? "";
    }
    editingID = id;
    document.querySelector("#task-form-title").textContent = "Edit task";
    document.querySelector("#save-task").textContent = "Save changes";
    cancelEdit.classList.remove("hidden");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function updateTaskStatus(id, status) {
    const task = await request(`/tasks/${id}`);
    task.status = status;
    await request(`/tasks/${id}`, {
      method: "PUT",
      body: JSON.stringify({
        title: task.title,
        subject: task.subject,
        description: task.description,
        date: task.date,
        startTime: task.startTime,
        endTime: task.endTime,
        status: task.status,
        priority: task.priority,
        actualStudyMinutes: task.actualStudyMinutes,
      }),
    });
    subjectsLoaded = false;
    await Promise.all([loadTasks(), loadCalendar()]);
  }

  taskForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = new FormData(taskForm);
    const body = Object.fromEntries(form.entries());
    body.actualStudyMinutes = Number(body.actualStudyMinutes || 0);
    try {
      await request(editingID ? `/tasks/${editingID}` : "/tasks", {
        method: editingID ? "PUT" : "POST",
        body: JSON.stringify(body),
      });
      showMessage(taskMessage, editingID ? "Task updated." : "Task added.", true);
      resetTaskForm();
      subjectsLoaded = false;
      await Promise.all([loadTasks(), loadCalendar()]);
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
        subjectsLoaded = false;
        await Promise.all([loadTasks(), loadCalendar()]);
      } else if (button.dataset.action === "edit") {
        await editTask(id);
      }
    } catch (error) {
      showMessage(taskMessage, error.message);
    }
  });

  taskList.addEventListener("change", async (event) => {
    const select = event.target.closest('select[data-action="status"]');
    if (!select) return;
    select.disabled = true;
    try {
      await updateTaskStatus(select.dataset.id, select.value);
      showMessage(taskMessage, "Task status updated.", true);
    } catch (error) {
      showMessage(taskMessage, error.message);
      await loadTasks();
    }
  });

  calendar.addEventListener("click", async (event) => {
    const button = event.target.closest('button[data-action="edit"]');
    if (!button) {
      const cell = event.target.closest(".calendar-cell");
      if (cell && !cell.querySelector(".calendar-task")) {
        showMessage(calendarMessage, `No tasks scheduled for ${cell.dataset.date}.`);
      }
      return;
    }
    try {
      await editTask(button.dataset.id);
    } catch (error) {
      showMessage(calendarMessage, error.message);
    }
  });

  filters.addEventListener("input", (event) => {
    if (event.target.name === "q") {
      window.clearTimeout(searchDelay);
      searchDelay = window.setTimeout(loadTasks, 250);
    }
  });
  filters.addEventListener("change", (event) => {
    if (event.target.name !== "groupBySubject") loadTasks();
    else loadTasks();
  });
  filters.addEventListener("reset", () => window.setTimeout(loadTasks, 0));
  document.querySelector("#calendar-previous").addEventListener("click", () => {
    calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() - 1, 1);
    loadCalendar();
  });
  document.querySelector("#calendar-next").addEventListener("click", () => {
    calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() + 1, 1);
    loadCalendar();
  });

  loadTasks();
  loadCalendar();
}

if (document.body.dataset.page === "dashboard") loadDashboard();
