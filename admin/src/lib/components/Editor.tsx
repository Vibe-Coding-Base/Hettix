import MonacoEditor, { EditorProps } from "@monaco-editor/react";

import { useAppearance } from "lib/AppearanceContext";

const defaultMonacoOptions: EditorProps["options"] = {
  readOnly: true,
  wordWrap: "on",
  minimap: {
    enabled: false,
  },
  scrollBeyondLastLine: false,
};

// languageForContentType maps a Content-Type to a Monaco language id.
function languageForContentType(contentType?: string): string | undefined {
  const mime = contentType?.split(";")[0].trim().toLowerCase();

  switch (mime) {
    case "text/html":
    case "application/xhtml+xml":
      return "html";
    case "application/json":
    case "application/manifest+json":
      return "json";
    case "application/javascript":
    case "text/javascript":
      return "javascript";
    case "text/css":
      return "css";
    case "application/xml":
    case "text/xml":
    case "image/svg+xml":
      return "xml";
    default:
      if (mime?.endsWith("+json")) {
        return "json";
      }
      if (mime?.endsWith("+xml")) {
        return "xml";
      }
      return undefined;
  }
}

// sniffLanguage guesses the language from the body when the Content-Type is
// missing or unhelpful, so responses still get highlighted. It also skips a
// leading XSSI guard (Google's )]}' , for(;;); / while(1);) that some APIs add.
function sniffLanguage(content: string): string | undefined {
  const trimmed = content.replace(/^(\)\]\}'[,\s]*|for\s*\(;;\);|while\s*\(1\);|&&&START&&&)\s*/, "").trimStart();
  if (trimmed === "") {
    return undefined;
  }

  if (trimmed[0] === "{" || trimmed[0] === "[") {
    return "json";
  }
  if (trimmed.startsWith("<?xml")) {
    return "xml";
  }
  if (/^<!doctype html|^<html[\s>]/i.test(trimmed)) {
    return "html";
  }
  if (trimmed[0] === "<") {
    return "xml";
  }

  return undefined;
}

interface Props {
  content: string;
  contentType?: string;
  language?: string;
  monacoOptions?: EditorProps["options"];
  onChange?: EditorProps["onChange"];
}

function Editor({ content, contentType, language, monacoOptions, onChange }: Props): JSX.Element {
  const resolved = language ?? languageForContentType(contentType) ?? sniffLanguage(content);
  const { fontFamily, fontSize } = useAppearance();

  return (
    <MonacoEditor
      language={resolved}
      theme="vs-dark"
      options={{ fontFamily, fontSize, ...defaultMonacoOptions, ...monacoOptions }}
      value={content}
      onChange={onChange}
    />
  );
}

export default Editor;
