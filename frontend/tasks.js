import { createCalendar } from "./calendar.js";
import { request, requestTasks, requireLogin, serverDate, showMessage } from "./common.js";

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
  let searchDelay;
  let latestTaskLoad = 0;

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
    const classes = {
      "Not Started": "status status-not-started",
      "In Progress": "status status-in-progress",
      Completed: "status status-completed",
      "On Hold": "status status-on-hold",
      "priority-high": "status status-priority-high",
      "priority-medium": "status status-priority-medium",
      "priority-low": "status status-priority-low",
    };
    return classes[status] || "status";
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

  function populateSubjects(tasks) {
    const selected = subjectFilter.value;
    const subjects = [...new Set(tasks.map((task) => task.subject))].sort((a, b) => a.localeCompare(b));
    subjectFilter.replaceChildren(new Option("All subjects", ""));
    subjects.forEach((subject) => subjectFilter.add(new Option(subject, subject)));
    subjectFilter.value = selected;
    subjectsLoaded = true;
  }

  async function loadTasks() {
    const loadID = ++latestTaskLoad;
    taskList.innerHTML = '<div class="empty-state">Loading tasks…</div>';
    try {
      const query = filterQuery();
      const tasks = await requestTasks(`/tasks${query ? `?${query}` : ""}`);
      if (loadID !== latestTaskLoad) return;
      renderTasks(tasks);
      if (!subjectsLoaded) {
        const allTasks = query ? await requestTasks("/tasks") : tasks;
        if (loadID === latestTaskLoad) populateSubjects(allTasks);
      }
    } catch (error) {
      if (loadID !== latestTaskLoad) return;
      showMessage(taskMessage, error.message);
      taskList.innerHTML = '<div class="empty-state">Tasks could not be loaded.</div>';
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

  const taskCalendar = createCalendar({
    calendar,
    calendarMessage,
    onTaskSelect: async (id) => {
      try {
        await editTask(id);
      } catch (error) {
        showMessage(calendarMessage, error.message);
      }
    },
    todayString,
    isOverdue,
    statusClass,
  });

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
    await Promise.all([loadTasks(), taskCalendar.load()]);
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
      await Promise.all([loadTasks(), taskCalendar.load()]);
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
        await Promise.all([loadTasks(), taskCalendar.load()]);
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

  filters.addEventListener("input", (event) => {
    if (event.target.name === "q") {
      window.clearTimeout(searchDelay);
      searchDelay = window.setTimeout(loadTasks, 250);
    }
  });
  filters.addEventListener("change", loadTasks);
  filters.addEventListener("reset", () => window.setTimeout(loadTasks, 0));

  loadTasks();
  taskCalendar.load();
}
