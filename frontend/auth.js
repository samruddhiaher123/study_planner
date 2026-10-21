import { request, showMessage } from "./common.js";

document.querySelectorAll("[data-toggle-password]").forEach((button) => {
  button.addEventListener("click", () => {
    const input = document.getElementById(button.dataset.togglePassword);
    if (!input) return;

    const showPassword = input.type === "password";
    input.type = showPassword ? "text" : "password";
    button.setAttribute("aria-pressed", String(showPassword));
    button.setAttribute("aria-label", showPassword ? "Hide password" : "Show password");
    button.querySelector(".eye-icon").classList.toggle("hidden", showPassword);
    button.querySelector(".eye-off-icon").classList.toggle("hidden", !showPassword);
    input.focus();
  });
});

function validateEmail(input) {
  input.setCustomValidity(input.validity.typeMismatch ? "Enter a valid email address." : "");
}

const loginForm = document.querySelector("#login-form");
if (loginForm) {
  const email = loginForm.elements.namedItem("email");
  const submitButton = loginForm.querySelector('[type="submit"]');
  const message = document.querySelector("#form-message");
  let submitting = false;

  email.addEventListener("input", () => validateEmail(email));
  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (submitting) return;
    validateEmail(email);
    if (!loginForm.reportValidity()) return;

    const form = new FormData(loginForm);
    submitting = true;
    submitButton.disabled = true;
    submitButton.textContent = "Signing in…";
    showMessage(message, "");
    try {
      const result = await request("/login", {
        method: "POST",
        body: JSON.stringify({ email: form.get("email"), password: form.get("password") }),
      });
      localStorage.setItem("studyPlannerToken", result.token);
      window.location.href = "/dashboard.html";
    } catch (error) {
      showMessage(message, error.message);
    } finally {
      submitting = false;
      submitButton.disabled = false;
      submitButton.textContent = "Sign in";
    }
  });
}

const registerForm = document.querySelector("#register-form");
if (registerForm) {
  const email = registerForm.elements.namedItem("email");
  const password = registerForm.elements.namedItem("password");
  const confirmPassword = registerForm.elements.namedItem("confirmPassword");
  const matchMessage = document.querySelector("#password-match");
  const message = document.querySelector("#form-message");
  const submitButton = registerForm.querySelector('[type="submit"]');
  let submitting = false;

  function validatePasswords() {
    const bothFilled = password.value !== "" && confirmPassword.value !== "";
    const matches = bothFilled && password.value === confirmPassword.value;
    confirmPassword.setCustomValidity(
      confirmPassword.value && !matches ? "Passwords do not match." : "",
    );
    matchMessage.textContent = !bothFilled
      ? ""
      : matches
        ? "Passwords match."
        : "Passwords do not match.";
    matchMessage.classList.toggle("match-success", matches);
    matchMessage.classList.toggle("match-error", bothFilled && !matches);
  }

  email.addEventListener("input", () => validateEmail(email));
  password.addEventListener("input", validatePasswords);
  confirmPassword.addEventListener("input", validatePasswords);

  registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (submitting) return;
    validateEmail(email);
    validatePasswords();
    if (!registerForm.reportValidity()) return;

    const form = new FormData(registerForm);
    submitting = true;
    submitButton.disabled = true;
    submitButton.textContent = "Creating account…";
    showMessage(message, "");
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
      validatePasswords();
    } catch (error) {
      showMessage(message, error.message);
    } finally {
      submitting = false;
      submitButton.disabled = false;
      submitButton.textContent = "Create account";
    }
  });
}
