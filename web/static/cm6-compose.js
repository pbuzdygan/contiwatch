import { EditorState, Compartment, EditorSelection } from "@codemirror/state";
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from "@codemirror/view";
import { indentWithTab, indentLess, defaultKeymap } from "@codemirror/commands";
import * as language from "@codemirror/language";
import { yaml } from "@codemirror/lang-yaml";
import { linter, lintGutter, forceLinting } from "@codemirror/lint";
import { autocompletion, completeFromList } from "@codemirror/autocomplete";

function parseComposeErrorLocation(message) {
  const text = String(message || "");
  let match = text.match(/line\s+(\d+)\s*[:,]?\s*column\s+(\d+)/i);
  if (match) {
    return { line: Number.parseInt(match[1], 10), column: Number.parseInt(match[2], 10) };
  }
  match = text.match(/docker-compose\.ya?ml:(\d+):(\d+)\b/i);
  if (match) {
    return { line: Number.parseInt(match[1], 10), column: Number.parseInt(match[2], 10) };
  }
  match = text.match(/line\s+(\d+)/i);
  if (match) {
    return { line: Number.parseInt(match[1], 10), column: 1 };
  }
  return { line: 1, column: 1 };
}

function findFirstYamlKeyLocation(docText, key) {
  const target = String(key || "").trim();
  if (!target) return null;
  const escaped = target.replace(/[.*+?^${}()|[\\]\\\\]/g, "\\\\$&");
  const re = new RegExp(`^\\s*${escaped}\\s*:`);
  const lines = String(docText || "").split(/\r?\n/);
  for (let i = 0; i < lines.length; i += 1) {
    if (re.test(lines[i])) {
      const col = lines[i].indexOf(target);
      return { line: i + 1, column: Math.max(1, col + 1) };
    }
  }
  return null;
}

function simplifyErrorMessage(message) {
  const raw = String(message || "").replace(/\s+/g, " ").trim();
  if (!raw) return "Validation failed.";
  let text = raw;
  text = text.replace(/^compose config failed:\s*/i, "");
  text = text.replace(/^validating\s+.*?docker-compose\.ya?ml:\s*/i, "");
  const additional = text.match(/Additional property\s+([A-Za-z0-9_-]+)\s+is not allowed/i);
  if (additional) {
    return `Validation failed: ${additional[1]} is not allowed.`;
  }
  const afterCompose = text.match(/docker-compose\.ya?ml:\s*(.*)$/i);
  if (afterCompose && afterCompose[1]) {
    return `Validation failed: ${afterCompose[1].trim()}`;
  }
  if (/^Validation failed:/i.test(text)) return text;
  return `Validation failed: ${text}`;
}

function isEnvOnlyError(message) {
  const raw = String(message || "").toLowerCase();
  return raw.includes("invalid .env entries:") || raw.includes("invalid .env entry");
}

function errorToDiagnostic(state, message) {
  const shortMessage = simplifyErrorMessage(message);
  let loc = parseComposeErrorLocation(message);
  if (loc.line === 1 && loc.column === 1) {
    const raw = String(message || "");
    const m = raw.match(/Additional property\s+([A-Za-z0-9_-]+)\s+is not allowed/i);
    if (m) {
      const found = findFirstYamlKeyLocation(state.doc.toString(), m[1]);
      if (found) loc = found;
    }
  }
  const line = Math.max(1, loc.line || 1);
  const column = Math.max(1, loc.column || 1);
  const lineInfo = state.doc.line(Math.min(line, state.doc.lines));
  const from = Math.min(lineInfo.to, lineInfo.from + column - 1);
  const to = Math.min(lineInfo.to, from + 1);
  const tooltipMessage = `Line ${line}: ${shortMessage}`;
  return [{
    from,
    to,
    severity: "error",
    message: tooltipMessage,
  }];
}

function envDiagnostics(state) {
  const diagnostics = [];
  for (let lineNo = 1; lineNo <= state.doc.lines; lineNo += 1) {
    const line = state.doc.line(lineNo);
    const raw = line.text;
    const trimmed = raw.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    if (raw.includes("\0")) {
      diagnostics.push({
        from: line.from,
        to: Math.min(line.to, line.from + 1),
        severity: "error",
        message: `Line ${lineNo}: invalid .env entry (NUL byte).`,
      });
      continue;
    }
    const eq = raw.indexOf("=");
    if (eq <= 0) {
      diagnostics.push({
        from: line.from,
        to: Math.min(line.to, line.from + 1),
        severity: "error",
        message: `Line ${lineNo}: invalid .env entry (expected KEY=value).`,
      });
      continue;
    }
    const key = raw.slice(0, eq).trim();
    if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
      const keyStart = raw.indexOf(key);
      diagnostics.push({
        from: line.from + Math.max(0, keyStart),
        to: Math.min(line.to, line.from + Math.max(1, keyStart + key.length)),
        severity: "error",
        message: `Line ${lineNo}: invalid key "${key}" (allowed: A-Z, a-z, 0-9, _; must not start with digit).`,
      });
    }
  }
  return diagnostics;
}

function newlineAndYamlIndent(view) {
  const { state } = view;
  const sel = state.selection.main;
  if (!sel || !sel.empty || state.selection.ranges.length !== 1) return false;
  const pos = sel.from;
  const line = state.doc.lineAt(pos);
  const col = pos - line.from;
  const text = line.text;
  const leading = (text.match(/^\s*/) || [""])[0];
  const before = text.slice(0, col);
  const after = text.slice(col);

  const stripInlineComment = (value) => String(value || "").replace(/\s+#.*$/, "");
  const beforeNoComment = stripInlineComment(before).trimEnd();
  const fullNoComment = stripInlineComment(text).trimEnd();

  let indent = leading;
  const indentStep = "  ";

  // If line ends with ":" (no value) -> indent one level under the key.
  if (pos === line.to && /:\s*$/.test(fullNoComment)) {
    indent = leading + indentStep;
  }

  // If cursor is right after ":" but there is a value after the cursor, treat it as moving the value
  // to the next line under the key (valid YAML) and indent one level (not under the ":" column).
  if (/:$/.test(beforeNoComment) && /\S/.test(after)) {
    indent = leading + indentStep;
  }

  const insert = `\n${indent}`;
  view.dispatch({
    changes: { from: pos, to: pos, insert },
    selection: EditorSelection.cursor(pos + insert.length),
    userEvent: "input.enter",
  });
  return true;
}

const composeTopLevelCompletions = completeFromList([
  { label: "services", type: "keyword", apply: "services:\n  " },
  { label: "volumes", type: "keyword", apply: "volumes:\n  " },
  { label: "networks", type: "keyword", apply: "networks:\n  " },
  { label: "secrets", type: "keyword", apply: "secrets:\n  " },
  { label: "configs", type: "keyword", apply: "configs:\n  " },
  { label: "name", type: "property", apply: "name: " },
  { label: "version", type: "property", apply: "version: \"3\"" },
]);

const composeServiceKeyCompletions = completeFromList([
  { label: "image", type: "property", apply: "image: " },
  { label: "container_name", type: "property", apply: "container_name: " },
  { label: "restart", type: "property", apply: "restart: " },
  { label: "ports", type: "property", apply: "ports:\n  - \"" },
  { label: "volumes", type: "property", apply: "volumes:\n  - " },
  { label: "environment", type: "property", apply: "environment:\n  " },
  { label: "env_file", type: "property", apply: "env_file:\n  - .env" },
  { label: "networks", type: "property", apply: "networks:\n  - " },
  { label: "depends_on", type: "property", apply: "depends_on:\n  - " },
  { label: "command", type: "property", apply: "command: " },
  { label: "build", type: "property", apply: "build: " },
  { label: "labels", type: "property", apply: "labels:\n  - " },
]);

const composeValueCompletionsByKey = {
  version: completeFromList([
    { label: "\"3\"", type: "constant", apply: "\"3\"" },
    { label: "\"3.8\"", type: "constant", apply: "\"3.8\"" },
    { label: "\"3.9\"", type: "constant", apply: "\"3.9\"" },
  ]),
  restart: completeFromList([
    { label: "no", type: "constant", apply: "no" },
    { label: "always", type: "constant", apply: "always" },
    { label: "unless-stopped", type: "constant", apply: "unless-stopped" },
    { label: "on-failure", type: "constant", apply: "on-failure" },
  ]),
  network_mode: completeFromList([
    { label: "bridge", type: "constant", apply: "bridge" },
    { label: "host", type: "constant", apply: "host" },
    { label: "none", type: "constant", apply: "none" },
  ]),
  pull_policy: completeFromList([
    { label: "always", type: "constant", apply: "always" },
    { label: "missing", type: "constant", apply: "missing" },
    { label: "if_not_present", type: "constant", apply: "if_not_present" },
    { label: "never", type: "constant", apply: "never" },
  ]),
};

function composeCompletionSource(context) {
  const line = context.state.doc.lineAt(context.pos);
  const indent = (line.text.match(/^\s*/) || [""])[0].length;
  const beforeText = line.text.slice(0, context.pos - line.from);
  const lineNoComment = beforeText.replace(/\s+#.*$/, "");
  const kv = lineNoComment.match(/^\s*([A-Za-z0-9_-]+)\s*:\s*(.*)$/);
  const word = context.matchBefore(/[A-Za-z0-9_-]+/);
  if (!word && !context.explicit) return null;

  // Value completion when cursor is in the value area of a known key.
  if (kv) {
    const key = kv[1];
    const valuePart = kv[2] || "";
    const isInValue = valuePart.length > 0 || /:\s*$/.test(lineNoComment);
    if (isInValue && composeValueCompletionsByKey[key]) {
      return composeValueCompletionsByKey[key](context);
    }
  }

  if (indent === 0) return composeTopLevelCompletions(context);
  if (indent >= 2) return composeServiceKeyCompletions(context);
  return null;
}

export function init({ container, textarea, getEnv, getUseEnv, onChange, mode = "yaml", lint = false }) {
  if (!container || !textarea) return null;
  const startDoc = textarea.value || "";
  let pending = null;
  const editable = new Compartment();
  let aborter = null;
  const debounceMs = lint ? 700 : 0;
  let stableDiagnostics = null;
  let candidateDiagnostics = null;
  let lastDiagnosticsAt = 0;
  let view = null;

  async function validate(docText) {
    if (aborter) aborter.abort();
    aborter = new AbortController();
    const payload = {
      compose_yaml: docText,
      env: getEnv ? getEnv() : "",
      use_env: getUseEnv ? Boolean(getUseEnv()) : false,
    };
    const res = await fetch("/api/stacks/validate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: aborter.signal,
    });
    const data = await res.json();
    if (data && data.valid) return [];
    const msg = (data && data.error) ? data.error : "Compose validation failed.";
    if (mode === "yaml" && isEnvOnlyError(msg)) {
      // Env errors are handled by the env editor linter; avoid showing them on compose.
      return [];
    }
    return errorToDiagnostic(view.state, msg);
  }

  function diagnosticsEqual(a, b) {
    if (!Array.isArray(a) || !Array.isArray(b)) return false;
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i += 1) {
      const x = a[i];
      const y = b[i];
      if (!x || !y) return false;
      if (x.from !== y.from || x.to !== y.to || x.severity !== y.severity || x.message !== y.message) return false;
    }
    return true;
  }

  function stabilizeDiagnostics(next) {
    if (!next || next.length === 0) {
      stableDiagnostics = null;
      candidateDiagnostics = null;
      return [];
    }
    // Avoid flicker: require same diagnostics twice before switching.
    if (!stableDiagnostics) {
      stableDiagnostics = next;
      candidateDiagnostics = null;
      lastDiagnosticsAt = Date.now();
      return next;
    }
    if (diagnosticsEqual(next, stableDiagnostics)) {
      candidateDiagnostics = null;
      lastDiagnosticsAt = Date.now();
      return next;
    }
    if (candidateDiagnostics && diagnosticsEqual(next, candidateDiagnostics)) {
      stableDiagnostics = next;
      candidateDiagnostics = null;
      lastDiagnosticsAt = Date.now();
      return next;
    }
    // If the stable diagnostic is very old, allow switching faster.
    const ageMs = Date.now() - lastDiagnosticsAt;
    if (ageMs > 5000) {
      stableDiagnostics = next;
      candidateDiagnostics = null;
      lastDiagnosticsAt = Date.now();
      return next;
    }
    candidateDiagnostics = next;
    return stableDiagnostics;
  }

  const lintExtension = lint
    ? linter(async (view) => {
      const docText = view.state.doc.toString();
      if (!docText.trim()) return [];
      if (pending) {
        clearTimeout(pending);
      }
      return new Promise((resolve) => {
        pending = setTimeout(async () => {
          try {
            let result = [];
            if (mode === "env") {
              result = envDiagnostics(view.state);
            } else if (mode === "yaml") {
              result = await validate(docText);
            }
            resolve(stabilizeDiagnostics(result));
          } catch (err) {
            resolve(stabilizeDiagnostics(errorToDiagnostic(view.state, err.message || "Compose validation failed.")));
          }
        }, debounceMs);
      });
    })
    : [];

  const baseExtensions = [
    lineNumbers(),
    highlightActiveLineGutter(),
    highlightActiveLine(),
    language.indentUnit.of("  "),
    EditorState.tabSize.of(2),
    EditorView.lineWrapping,
    keymap.of([
      { key: "Enter", run: newlineAndYamlIndent },
      indentWithTab,
      { key: "Shift-Tab", run: indentLess },
      ...defaultKeymap,
    ]),
    editable.of(EditorView.editable.of(true)),
  ];

  if (mode === "yaml") {
    baseExtensions.push(yaml());
    if (typeof language.indentOnInput === "function") {
      baseExtensions.push(language.indentOnInput());
    }
    if (typeof language.bracketMatching === "function") {
      baseExtensions.push(language.bracketMatching());
    }
    if (typeof language.foldGutter === "function") {
      baseExtensions.push(language.foldGutter());
    }
    if (typeof language.codeFolding === "function") {
      baseExtensions.push(language.codeFolding());
    }
    if (language.foldKeymap) {
      baseExtensions.push(keymap.of(language.foldKeymap));
    }
    if (typeof language.indentGuides === "function") {
      baseExtensions.push(language.indentGuides());
    }
    baseExtensions.push(autocompletion({ override: [composeCompletionSource], activateOnTyping: true }));
  }
  if (lint) {
    baseExtensions.push(lintGutter(), lintExtension);
  }

  const state = EditorState.create({
    doc: startDoc,
    extensions: [
      ...baseExtensions,
      EditorView.updateListener.of((update) => {
        if (!update.docChanged) return;
        const value = update.state.doc.toString();
        textarea.value = value;
        if (onChange) onChange(value);
      }),
    ],
  });

  view = new EditorView({
    state,
    parent: container,
  });

  textarea.value = startDoc;

  return {
    getValue() {
      return view.state.doc.toString();
    },
    setValue(value) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: value || "" },
      });
      textarea.value = value || "";
    },
    setReadOnly(disabled) {
      view.dispatch({
        effects: editable.reconfigure(EditorView.editable.of(!disabled)),
      });
    },
    focus() {
      view.focus();
    },
    forceLint() {
      try {
        forceLinting(view);
      } catch {
        // ignore
      }
    },
  };
}

export default { init };
