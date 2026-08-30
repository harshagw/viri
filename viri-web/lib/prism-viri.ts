import Prism from "prismjs";

export const registerViri = () => {
  Prism.languages.viri = {
    comment: {
      pattern: /\/\/.*$/,
      greedy: true,
    },
    string: {
      pattern: /"[^"]*"/,
      greedy: true,
    },
    "class-name": {
      pattern: /(\bclass\s+)\w+/,
      lookbehind: true,
    },
    function: {
      pattern: /\b[a-zA-Z_]\w*(?=\()/,
    },
    keyword:
      /\b(?:and|or|if|else|for|while|return|break|continue|var|const|fun|class|print|init|this|super|import|export|as|map)\b/,
    // number, string and bool are predeclared type names rather than
    // keywords, but a reader expects them to look like types.
    builtin: /\b(?:number|string|bool)\b/,
    boolean: /\b(?:true|false)\b/,
    number: /\b\d+(?:\.\d+)?\b/,
    operator: /==|!=|<=|>=|[=!<>\+\-\*\/]/,
    punctuation: /[{}[\];(),.:]/,
  };
};
