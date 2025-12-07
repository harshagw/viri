# Viri syntax highlighting for VS Code/Cursor

## Install locally

1. Install `vsce` if you do not already have it: `npm install -g @vscode/vsce`.
2. From this folder run `vsce package` to produce `viri-syntax-0.0.1.vsix`.
3. In VS Code/Cursor, run **Extensions: Install from VSIX** and select the generated file.
4. Open a `.viri` file and verify highlighting.

## Development

- Use **Run Extension** (Extension Development Host) for quick iteration.
- Update `syntaxes/viri.tmLanguage.json` to tweak scopes; themes pick colors based on those scopes.
- `language-configuration.json` controls comments, brackets, auto-closing pairs, indentation, and on-enter rules.
- Snippets live in `snippets/viri.json` (print, var, fun, class, if/else, for, while).
- Extras: TODO/FIXME/NOTE are highlighted inside `//` comments; hash/object keys before `:` get a property scope for better theming.
- Runtime features: hover docs and signature help for built-ins (`print`, `len`, `clock`) via `extension.js` (activated on Viri files).
