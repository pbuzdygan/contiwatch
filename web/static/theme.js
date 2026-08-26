(function initializeTheme() {
  try {
    const mode = localStorage.getItem("contiwatch_theme");
    document.documentElement.setAttribute("data-theme", mode === "dark" ? "dark" : "light");
  } catch {
    document.documentElement.setAttribute("data-theme", "light");
  }
}());
