const API = "/api";

const token = () => localStorage.getItem("studyPlannerToken");

export let serverDate = "";

export async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (token()) headers.Authorization = `Bearer ${token()}`;
  const response = await fetch(`${API}${path}`, { ...options, headers });
  serverDate = response.headers.get("X-Server-Date") || serverDate;
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || "Something went wrong");
  return data;
}

export async function requestTasks(path) {
  const tasks = await request(path);
  if (!Array.isArray(tasks)) {
    throw new Error("The task list response was invalid");
  }
  return tasks;
}

export function showMessage(element, message, success = false) {
  if (!element) return;
  element.textContent = message;
  element.classList.toggle("success-message", success);
}

export function requireLogin() {
  if (!token()) window.location.href = "/login.html";
}

export function formatStudyTime(minutes) {
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (hours === 0) return `${remainingMinutes} min`;
  if (remainingMinutes === 0) return `${hours} hr`;
  return `${hours} hr ${remainingMinutes} min`;
}

document.querySelectorAll("[data-logout]").forEach((button) => {
  button.addEventListener("click", () => {
    localStorage.removeItem("studyPlannerToken");
    window.location.href = "/login.html";
  });
});
