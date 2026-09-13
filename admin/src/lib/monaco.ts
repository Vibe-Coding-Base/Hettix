// Bundle the Monaco editor with the app instead of loading it from a CDN at
// runtime. A CDN load makes the editor show "Loading..." (or fail entirely
// offline), which is unacceptable for a locally embedded tool. Importing the
// package and its web workers here, and pointing @monaco-editor/react's loader
// at it, makes the editor render instantly and work offline.
import { loader } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import cssWorker from "monaco-editor/esm/vs/language/css/css.worker?worker";
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker";
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

self.MonacoEnvironment = {
  getWorker(_workerId: string, label: string) {
    switch (label) {
      case "json":
        return new jsonWorker();
      case "css":
      case "scss":
      case "less":
        return new cssWorker();
      case "html":
      case "handlebars":
      case "razor":
        return new htmlWorker();
      case "typescript":
      case "javascript":
        return new tsWorker();
      default:
        return new editorWorker();
    }
  },
};

// A lightweight language for raw HTTP messages: it colours the request/status
// line and header names, then highlights the body with generic JSON/XML tokens.
// It has no validation, so headers never show as errors the way a real JSON/XML
// grammar would.
monaco.languages.register({ id: "http" });
monaco.languages.setMonarchTokensProvider("http", {
  tokenizer: {
    root: [
      [
        /^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE|CONNECT)(\s+)(\S+)(\s+)(HTTP\/[\d.]+)\s*$/,
        ["keyword", "white", "string", "white", "type"],
      ],
      [/^(HTTP\/[\d.]+)(\s+)(\d{3})(\s*)(.*)$/, ["type", "white", "number", "white", "string"]],
      [/^([A-Za-z0-9-]+)(:)(\s*)(.*)$/, ["attribute.name", "delimiter", "white", "string"]],
      [/^\s*$/, { token: "white", next: "@body" }],
    ],
    body: [
      [/"(?:[^"\\]|\\.)*"/, "string"],
      [/\b(?:true|false|null)\b/, "keyword"],
      [/-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/, "number"],
      [/[{}[\],:]/, "delimiter"],
      [/<\/?[A-Za-z][\w-]*/, "tag"],
      [/>/, "tag"],
    ],
  },
});

loader.config({ monaco });
