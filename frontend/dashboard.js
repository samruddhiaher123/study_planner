import { formatStudyTime, request, requireLogin, showMessage } from "./common.js";

async function loadDashboard() {
  requireLogin();
  const message = document.querySelector("#dashboard-message");
  try {
    const dashboard = await request("/dashboard");
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

if (document.body.dataset.page === "dashboard") loadDashboard();
