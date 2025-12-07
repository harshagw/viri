// Basic language features for Viri: hover docs and signature help.
const vscode = require("vscode");

// Built-in symbols and docs.
const BUILTIN_DOCS = {
  print: {
    label: "print(value)",
    detail: "Prints the given value.",
    markdown: new vscode.MarkdownString("**print(value)** — prints the given value to output."),
  },
  len: {
    label: "len(value)",
    detail: "Returns length of string, array, or hash.",
    markdown: new vscode.MarkdownString("**len(value)** — length of string, array, or hash."),
  },
  clock: {
    label: "clock()",
    detail: "Returns current Unix time (seconds).",
    markdown: new vscode.MarkdownString("**clock()** — current Unix time (seconds)."),
  },
};

function activate(context) {
  // Hover provider
  context.subscriptions.push(
    vscode.languages.registerHoverProvider("viri", {
      provideHover(document, position) {
        const wordRange = document.getWordRangeAtPosition(position, /[A-Za-z_][A-Za-z0-9_]*/);
        if (!wordRange) return;
        const word = document.getText(wordRange);
        const entry = BUILTIN_DOCS[word];
        if (!entry) return;
        entry.markdown.isTrusted = true;
        return new vscode.Hover(entry.markdown, wordRange);
      },
    })
  );

  // Signature help provider for builtins.
  context.subscriptions.push(
    vscode.languages.registerSignatureHelpProvider(
      "viri",
      {
        provideSignatureHelp(document, position) {
          const text = document.getText(new vscode.Range(new vscode.Position(0, 0), position));
          const triggerIndex = text.lastIndexOf("(");
          if (triggerIndex === -1) return null;

          const before = text.slice(0, triggerIndex);
          const match = /([A-Za-z_][A-Za-z0-9_]*)\s*$/.exec(before);
          if (!match) return null;
          const name = match[1];
          const entry = BUILTIN_DOCS[name];
          if (!entry) return null;

          const sigInfo = new vscode.SignatureInformation(entry.label, entry.detail);
          const params = entry.label
            .slice(entry.label.indexOf("(") + 1, entry.label.indexOf(")"))
            .split(",")
            .map((p) => p.trim())
            .filter(Boolean)
            .map((p) => new vscode.ParameterInformation(p));
          sigInfo.parameters = params;

          const sigHelp = new vscode.SignatureHelp();
          sigHelp.signatures = [sigInfo];
          sigHelp.activeSignature = 0;
          sigHelp.activeParameter = params.length ? 0 : undefined;
          return sigHelp;
        },
      },
      "(",
      "," // trigger characters
    )
  );
}

function deactivate() {}

module.exports = {
  activate,
  deactivate,
};
