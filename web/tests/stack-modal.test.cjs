const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const source = fs.readFileSync(path.join(__dirname, "../static/app.js"), "utf8");
const functions = [
  "showStackModalError", "clearStackModalError", "getComposeValue", "setComposeValue",
  "getEnvValue", "setEnvValue", "shouldSaveStackEnv", "setStackEditorView",
  "setStackModalLoading", "updateStackModalBusyState", "setStackModalOperation",
  "updateStackEnvDeleteButton", "openStackModal", "getStackModalValues",
  "getStackModalChangedFields", "closeStackModal", "saveStackFromModal", "runStackActionFromModal",
].map((name) => {
  const start = source.search(new RegExp(`^(?:async )?function ${name}\\(`, "m"));
  assert.notEqual(start, -1, `Missing application function: ${name}`);
  const rest = source.slice(start + 1);
  const next = rest.search(/^(?:async )?function \w+\(/m);
  return source.slice(start, start + 1 + next);
}).join("\n");

function element() {
  const classes = new Set();
  return {
    value: "", textContent: "", dataset: {}, disabled: false,
    classList: {
      add: (name) => classes.add(name), remove: (name) => classes.delete(name),
      contains: (name) => classes.has(name),
      toggle(name, enabled) { if (enabled) classes.add(name); else classes.delete(name); },
    },
    setAttribute() {}, contains: () => false,
    focus() { this.focused = true; },
  };
}

function fixture() {
  const requests = [];
  let persisted = { compose_yaml: "services:\n  app:\n    image: example:v1", env: "KEY=old", has_env: true };
  const context = {
    document: { activeElement: null },
    window: { matchMedia: () => ({ matches: false }), setTimeout: (fn) => fn() },
    containersSelectedScope: "local:test", editingStackName: "", stackEditorView: "split",
    stackEnvExists: false, stackEnvDeletePending: false, stackEnvDeleteConfirming: false,
    stackModalLoading: false, stackModalSaving: false, stackModalLoadFailed: false,
    stackModalActionInProgress: "", stackModalSavedValues: null,
    composeEditor: null, envEditor: null,
    isValidStackName: (name) => /^[A-Za-z0-9_-]+$/.test(name),
    showToast() {}, notify() {}, refreshStacks: async () => {},
    fetchJSON: async (url, options) => {
      requests.push(url);
      if (url.startsWith("/api/stacks/get")) return { ...persisted };
      if (url === "/api/stacks/validate") return { valid: true };
      if (url === "/api/stacks/save") {
        const body = JSON.parse(options.body);
        persisted = { compose_yaml: body.compose_yaml, env: body.env, has_env: body.use_env };
        return {};
      }
      throw new Error(`Unexpected request: ${url}`);
    },
    runStackAction: async (name, action) => { requests.push(action); return true; },
  };
  for (const name of [
    "stackModal", "stackModalTitle", "stackModalClose", "stackModalCancel", "stackModalSave",
    "stackModalSaveIcon", "stackModalComposeUp", "stackModalComposeDown", "stackModalRedeploy",
    "stackModalRestart", "stackNameRow", "stackNameInput", "stackComposeInput", "stackEnvInput",
    "stackEnvDeleteBtn", "stackModalErrorEl", "stackModalOperationEl", "stackViewSplitBtn",
    "stackViewComposeBtn", "stackViewEnvBtn",
  ]) context[name] = element();
  vm.createContext(context);
  vm.runInContext(functions, context);
  return { context, requests, persisted: () => persisted };
}

test("New stack warns about every changed field and Cancel performs no save", async () => {
  const { context: c, requests } = fixture();
  await c.openStackModal();
  c.stackNameInput.value = "new_stack";
  c.stackComposeInput.value = "services: {}";
  c.stackEnvInput.value = "KEY=new";
  c.closeStackModal();
  assert.equal(c.stackModal.classList.contains("hidden"), false);
  assert.match(c.stackModalErrorEl.textContent, /Name, Compose \(docker-compose.yml\), \.env/);
  assert.match(c.stackModalErrorEl.textContent, /Save.*Cancel/);
  assert.equal(c.stackModalSave.focused, true);
  c.closeStackModal({ discardChanges: true });
  assert.equal(c.stackModal.classList.contains("hidden"), true);
  assert.deepEqual(requests, []);
  await c.openStackModal();
  assert.equal(c.stackNameInput.value, "");
  assert.equal(c.stackComposeInput.value, "");
  assert.equal(c.stackEnvInput.value, "");
});

for (const [field, label] of [["stackComposeInput", "Compose"], ["stackEnvInput", ".env"]]) {
  test(`Edit stack detects ${label}, allows reverting, and Cancel preserves saved data`, async () => {
    const { context: c, persisted } = fixture();
    await c.openStackModal("existing");
    const previous = { ...persisted() };
    const original = c[field].value;
    c[field].value += "\n# edited";
    c.closeStackModal();
    assert.ok(c.stackModalErrorEl.textContent.includes(label));
    assert.equal(c.stackModal.classList.contains("hidden"), false);
    c[field].value = original;
    assert.equal(c.getStackModalChangedFields().length, 0);
    c[field].value += "\n# edited";
    c.closeStackModal({ discardChanges: true });
    assert.deepEqual(persisted(), previous);
    await c.openStackModal("existing");
    assert.equal(c[field].value, original);
    c.closeStackModal();
    assert.equal(c.stackModal.classList.contains("hidden"), true);
  });
}

test("CodeMirror changes are detected through the live editor value", async () => {
  const { context: c } = fixture();
  await c.openStackModal("existing");
  c.composeEditor = { getValue: () => "services: {}" };
  c.closeStackModal();
  assert.match(c.stackModalErrorEl.textContent, /Compose/);
  assert.equal(c.stackModal.classList.contains("hidden"), false);
});

test("Pending deletion of an empty .env is dirty; undo restores clean state", async () => {
  const { context: c } = fixture();
  await c.openStackModal("existing");
  c.stackEnvInput.value = "";
  c.stackModalSavedValues = c.getStackModalValues();
  c.stackEnvDeletePending = true;
  assert.equal(c.getStackModalChangedFields().join(), ".env");
  c.stackEnvDeletePending = false;
  assert.equal(c.getStackModalChangedFields().length, 0);
});

test("Save persists edited contents and the toolbar save resets the dirty baseline", async () => {
  const { context: c, persisted } = fixture();
  await c.openStackModal("existing");
  c.stackComposeInput.value = "services:\n  app:\n    image: example:v2";
  c.stackEnvInput.value = "KEY=new";
  assert.equal(await c.saveStackFromModal({ closeOnSuccess: false }), true);
  assert.equal(persisted().env, "KEY=new");
  assert.match(persisted().compose_yaml, /example:v2/);
  assert.equal(c.stackModal.classList.contains("hidden"), false);
  assert.equal(c.getStackModalChangedFields().length, 0);
  c.stackEnvInput.value = "KEY=unsaved";
  c.closeStackModal({ discardChanges: true });
  await c.openStackModal("existing");
  assert.equal(c.stackEnvInput.value, "KEY=new");
  c.stackEnvInput.value = "KEY=final";
  assert.equal(await c.saveStackFromModal(), true);
  assert.equal(persisted().env, "KEY=final");
  assert.equal(c.stackModal.classList.contains("hidden"), true);
});

test("Saving .env deletion leaves a clean editor and removes the persisted file", async () => {
  const { context: c, persisted } = fixture();
  await c.openStackModal("existing");
  c.stackEnvDeletePending = true;
  assert.equal(await c.saveStackFromModal({ closeOnSuccess: false }), true);
  assert.equal(persisted().has_env, false);
  assert.equal(c.stackEnvInput.value, "");
  assert.equal(c.shouldSaveStackEnv(), false);
  assert.equal(c.getStackModalChangedFields().length, 0);
});

test("While saving, closing and duplicate saves are blocked; failed saves retain edits", async () => {
  const { context: c, persisted } = fixture();
  await c.openStackModal("existing");
  const previous = { ...persisted() };
  c.stackEnvInput.value = "KEY=new";
  let rejectSave;
  c.fetchJSON = async (url) => {
    if (url === "/api/stacks/validate") return { valid: true };
    return new Promise((resolve, reject) => { rejectSave = reject; });
  };
  const saving = c.saveStackFromModal();
  await Promise.resolve();
  assert.equal(c.stackModalCancel.disabled, true);
  assert.equal(c.stackEnvInput.disabled, true);
  c.closeStackModal({ discardChanges: true });
  assert.equal(c.stackModal.classList.contains("hidden"), false);
  assert.equal(await c.saveStackFromModal(), false);
  rejectSave(new Error("Save unavailable"));
  assert.equal(await saving, false);
  assert.equal(c.stackModalCancel.disabled, false);
  assert.equal(c.stackEnvInput.disabled, false);
  assert.equal(c.stackEnvInput.value, "KEY=new");
  assert.equal(c.getStackModalChangedFields().join(), ".env");
  assert.deepEqual(persisted(), previous);
  assert.equal(c.stackModalErrorEl.textContent, "Save unavailable");
});

for (const action of ["up", "down", "redeploy", "restart"]) {
  test(`${action} validates and saves before executing`, async () => {
    const { context: c, requests } = fixture();
    await c.openStackModal("existing");
    requests.length = 0;
    await c.runStackActionFromModal(action);
    assert.deepEqual(requests, ["/api/stacks/validate", "/api/stacks/save", action]);
    assert.equal(c.getStackModalChangedFields().length, 0);
  });
}

for (const failure of ["validation", "save"]) {
  test(`${failure} failure prevents the stack operation and keeps dirty data`, async () => {
    const { context: c } = fixture();
    await c.openStackModal("existing");
    c.stackEnvInput.value = "KEY=new";
    c.fetchJSON = async (url) => {
      if (failure === "validation") return { valid: false, error: "Invalid Compose" };
      if (url === "/api/stacks/validate") return { valid: true };
      throw new Error("Save unavailable");
    };
    c.runStackAction = async () => assert.fail("Action must not start");
    await c.runStackActionFromModal("up");
    assert.equal(c.stackModal.classList.contains("hidden"), false);
    assert.equal(c.getStackModalChangedFields().join(), ".env");
    assert.equal(c.stackModalSaving, false);
    assert.equal(c.stackModalActionInProgress, "");
  });
}

test("An operation failure does not undo its successful save", async () => {
  const { context: c, persisted } = fixture();
  await c.openStackModal("existing");
  c.stackEnvInput.value = "KEY=new";
  c.runStackAction = async () => false;
  await c.runStackActionFromModal("up");
  assert.equal(persisted().env, "KEY=new");
  assert.equal(c.getStackModalChangedFields().length, 0);
  assert.match(c.stackModalOperationEl.textContent, /Stack action failed/);
  c.closeStackModal({ discardChanges: true });
  assert.equal(persisted().env, "KEY=new");
});

test("Loading blocks dismissal, and a load failure allows closing without saving", async () => {
  const { context: c } = fixture();
  let rejectLoad;
  c.fetchJSON = () => new Promise((resolve, reject) => { rejectLoad = reject; });
  const loading = c.openStackModal("existing");
  assert.equal(c.stackModalClose.disabled, true);
  c.closeStackModal();
  assert.equal(c.stackModal.classList.contains("hidden"), false);
  rejectLoad(new Error("Load unavailable"));
  await loading;
  assert.equal(c.stackModalSave.disabled, true);
  assert.equal(c.stackModalClose.disabled, false);
  assert.equal(c.stackModalErrorEl.textContent, "Load unavailable");
  c.closeStackModal();
  assert.equal(c.stackModal.classList.contains("hidden"), true);
});
