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
    keyword: /\b(?:and|or|if|else|for|while|return|break|var|fun|class|print|init|this|super|import|as)\b/,
    boolean: /\b(?:true|false)\b/,
    nil: /\bnil\b/,
    number: /\b\d+(?:\.\d+)?\b/,
    operator: /==|!=|<=|>=|[=!<>\+\-\*\/]/,
    punctuation: /[{}[\];(),.:]/,
  };
};
