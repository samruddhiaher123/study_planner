import { requestTasks, showMessage } from "./common.js";

export function createCalendar({ calendar, calendarMessage, onTaskSelect, todayString, isOverdue, statusClass }) {
  let month = new Date(new Date().getFullYear(), new Date().getMonth(), 1);
  let selectedDate = "";
  let latestLoad = 0;

  const isoDate = (date) =>
    `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;

  async function load() {
    const loadID = ++latestLoad;
    const year = month.getFullYear();
    const monthNumber = month.getMonth();
    const monthLabel = document.querySelector("#calendar-month");
    monthLabel.textContent = month.toLocaleDateString(undefined, { month: "long", year: "numeric" });
    calendar.innerHTML = '<div class="empty-state">Loading calendar…</div>';
    showMessage(calendarMessage, "");

    const firstDay = new Date(year, monthNumber, 1);
    const mondayOffset = (firstDay.getDay() + 6) % 7;
    const start = new Date(year, monthNumber, 1 - mondayOffset);
    const finalDay = new Date(start.getFullYear(), start.getMonth(), start.getDate() + 41);
    const startDate = isoDate(start);
    const endDate = isoDate(finalDay);

    try {
      const tasks = await requestTasks(`/tasks?from=${startDate}&to=${endDate}`);
      if (loadID !== latestLoad) return;
      selectedDate ||= todayString();

      const tasksByDate = new Map();
      for (const task of tasks) {
        if (
          !task ||
          typeof task.date !== "string" ||
          !/^\d{4}-\d{2}-\d{2}$/.test(task.date) ||
          isoDate(new Date(`${task.date}T00:00:00`)) !== task.date ||
          task.date < startDate ||
          task.date > endDate
        ) {
          throw new Error("The task list contains an invalid task date");
        }
        if (!tasksByDate.has(task.date)) tasksByDate.set(task.date, []);
        tasksByDate.get(task.date).push(task);
      }

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
        cell.className = `calendar-cell${date.getMonth() !== monthNumber ? " other-month" : ""}${dateKey === todayString() ? " calendar-today" : ""}${dateKey === selectedDate ? " calendar-selected" : ""}`;
        cell.dataset.date = dateKey;
        cell.innerHTML = `<span class="calendar-date">${date.getDate()}</span>`;
        for (const task of tasksByDate.get(dateKey) || []) {
          const button = document.createElement("button");
          button.type = "button";
          button.className = `calendar-task ${statusClass(task.status)}${isOverdue(task) ? " calendar-overdue" : ""}`;
          button.dataset.action = "edit";
          button.dataset.id = task.id;
          button.textContent = `${task.startTime} ${task.title}`;
          button.title = `${task.subject} · ${task.status}${isOverdue(task) ? " · Overdue" : ""}`;
          cell.append(button);
        }
        calendar.append(cell);
      }

      const selectedTasks = tasksByDate.get(selectedDate) || [];
      if (selectedTasks.length === 0) {
        showMessage(calendarMessage, "No tasks scheduled for this date.");
      }
    } catch (error) {
      if (loadID !== latestLoad) return;
      calendar.replaceChildren();
      calendar.innerHTML = '<div class="empty-state">Calendar could not be loaded.</div>';
      showMessage(calendarMessage, error.message);
    }
  }

  calendar.addEventListener("click", async (event) => {
    const cell = event.target.closest(".calendar-cell");
    if (!cell) return;
    selectedDate = cell.dataset.date;
    calendar.querySelectorAll(".calendar-selected").forEach((selected) => {
      selected.classList.remove("calendar-selected");
    });
    cell.classList.add("calendar-selected");

    const taskButton = event.target.closest('button[data-action="edit"]');
    if (!taskButton) {
      const hasTasks = cell.querySelector(".calendar-task");
      showMessage(
        calendarMessage,
        hasTasks ? "Select a task to edit it." : "No tasks scheduled for this date.",
      );
      return;
    }

    showMessage(calendarMessage, "");
    await onTaskSelect(taskButton.dataset.id);
  });

  document.querySelector("#calendar-previous").addEventListener("click", () => {
    month = new Date(month.getFullYear(), month.getMonth() - 1, 1);
    selectedDate = isoDate(month);
    load();
  });
  document.querySelector("#calendar-next").addEventListener("click", () => {
    month = new Date(month.getFullYear(), month.getMonth() + 1, 1);
    selectedDate = isoDate(month);
    load();
  });

  return { load };
}
