import MonacoEditor, { EditorProps, OnMount } from "@monaco-editor/react";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import { Box, IconButton, Popover, Typography } from "@mui/material";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { useAppearance } from "lib/AppearanceContext";
import { DecoderDialog } from "lib/components/DecoderDialog";

const defaultMonacoOptions: EditorProps["options"] = {
  readOnly: true,
  wordWrap: "on",
  minimap: {
    enabled: false,
  },
  scrollBeyondLastLine: false,
  // Drop the chrome that reads as ugly "borders": the overview ruler strip on
  // the right, editor shadows, and the current-line highlight.
  overviewRulerLanes: 0,
  overviewRulerBorder: false,
  hideCursorInOverviewRuler: true,
  renderLineHighlight: "none",
  scrollbar: {
    verticalScrollbarSize: 10,
    horizontalScrollbarSize: 10,
    useShadows: false,
    alwaysConsumeMouseWheel: false,
  },
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

interface DecodePopover {
  top: number;
  left: number;
  label: string;
  value: string;
}

interface DecodeDialogState {
  top: number;
  left: number;
  text: string;
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
  const navigate = useNavigate();
  const [decode, setDecode] = useState<DecodePopover | null>(null);
  const [decodeDialog, setDecodeDialog] = useState<DecodeDialogState | null>(null);

  // handleMount adds Burp-style "Decode…" and "Send to Decoder" items to the
  // editor's right-click menu, so encoded strings can be decoded in place
  // without leaving the request/response view.
  const handleMount: OnMount = (editor) => {
    const anchorAtSelection = (): { top: number; left: number } | undefined => {
      const selection = editor.getSelection();
      const dom = editor.getDomNode();
      if (!selection || !dom) {
        return undefined;
      }
      const visible = editor.getScrolledVisiblePosition(selection.getEndPosition());
      if (!visible) {
        return undefined;
      }
      const rect = dom.getBoundingClientRect();
      return { top: rect.top + visible.top + visible.height, left: rect.left + visible.left };
    };

    editor.addAction({
      id: "hettix.decode",
      label: "Decode…",
      contextMenuGroupId: "9_hettix",
      contextMenuOrder: 0,
      precondition: "editorHasSelection",
      run: (ed) => {
        const selection = ed.getSelection();
        const model = ed.getModel();
        if (!selection || !model || selection.isEmpty()) {
          return;
        }
        const anchor = anchorAtSelection();
        if (!anchor) {
          return;
        }
        setDecodeDialog({ ...anchor, text: model.getValueInRange(selection) });
      },
    });

    editor.addAction({
      id: "hettix.sendToDecoder",
      label: "Send to Decoder",
      contextMenuGroupId: "9_hettix",
      contextMenuOrder: 1,
      run: (ed) => {
        const selection = ed.getSelection();
        const model = ed.getModel();
        const text =
          selection && model && !selection.isEmpty() ? model.getValueInRange(selection) : model?.getValue() ?? "";
        navigate(`/decoder?input=${encodeURIComponent(text)}`);
      },
    });
  };

  return (
    <>
      <MonacoEditor
        language={resolved}
        theme="vs-dark"
        options={{ fontFamily, fontSize, ...defaultMonacoOptions, ...monacoOptions }}
        value={content}
        onChange={onChange}
        onMount={handleMount}
      />
      <DecoderDialog
        open={decodeDialog !== null}
        initialText={decodeDialog?.text ?? ""}
        onClose={() => setDecodeDialog(null)}
        onApply={(label, value) => {
          if (decodeDialog) {
            setDecode({ top: decodeDialog.top, left: decodeDialog.left, label, value });
          }
        }}
      />
      <Popover
        open={decode !== null}
        onClose={() => setDecode(null)}
        anchorReference="anchorPosition"
        anchorPosition={decode ? { top: decode.top, left: decode.left } : undefined}
        transformOrigin={{ vertical: "top", horizontal: "left" }}
      >
        {decode && (
          <Box sx={{ p: 1, maxWidth: 460 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 0.5, mb: 0.5 }}>
              <Typography variant="caption" color="text.secondary">
                {decode.label} decoded
              </Typography>
              <IconButton
                size="small"
                aria-label="Copy decoded value"
                onClick={() => navigator.clipboard?.writeText(decode.value)}
              >
                <ContentCopyIcon sx={{ fontSize: 14 }} />
              </IconButton>
            </Box>
            <Box
              sx={{
                fontFamily: "monospace",
                fontSize: 12,
                whiteSpace: "pre-wrap",
                wordBreak: "break-all",
                maxHeight: 240,
                overflow: "auto",
              }}
            >
              {decode.value}
            </Box>
          </Box>
        )}
      </Popover>
    </>
  );
}

export default Editor;
