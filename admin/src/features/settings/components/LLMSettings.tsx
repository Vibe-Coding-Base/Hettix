import {
  Alert,
  Box,
  Button,
  CircularProgress,
  FormControlLabel,
  MenuItem,
  Snackbar,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import { LlmSettingsDocument, useLlmSettingsQuery, useUpdateLlmSettingsMutation } from "lib/graphql/generated";

// Presets pre-fill the base URL for common OpenAI-compatible providers. "custom"
// leaves the fields untouched.
const PRESETS: Record<string, { baseURL: string; model: string; needsKey: boolean }> = {
  deepseek: { baseURL: "https://api.deepseek.com/v1", model: "deepseek-chat", needsKey: true },
  openai: { baseURL: "https://api.openai.com/v1", model: "gpt-4o", needsKey: true },
  ollama: { baseURL: "http://localhost:11434/v1", model: "llama3.1", needsKey: false },
  custom: { baseURL: "", model: "", needsKey: false },
};

export default function LLMSettings(): JSX.Element {
  const { data } = useLlmSettingsQuery();
  const [provider, setProvider] = useState("deepseek");
  const [baseURL, setBaseURL] = useState("");
  const [model, setModel] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [enabled, setEnabled] = useState(false);
  const [hasApiKey, setHasApiKey] = useState(false);
  const [saved, setSaved] = useState(false);

  const [update, { loading, error }] = useUpdateLlmSettingsMutation({
    refetchQueries: [{ query: LlmSettingsDocument }],
    onCompleted() {
      setApiKey("");
      setSaved(true);
    },
  });

  useEffect(() => {
    const s = data?.llmSettings;
    if (s) {
      setProvider(s.provider || "custom");
      setBaseURL(s.baseURL);
      setModel(s.model);
      setEnabled(s.enabled);
      setHasApiKey(s.hasApiKey);
    }
  }, [data]);

  const onPreset = (value: string) => {
    setProvider(value);
    const preset = PRESETS[value];
    if (preset && value !== "custom") {
      setBaseURL(preset.baseURL);
      if (!model) {
        setModel(preset.model);
      }
    }
  };

  const onSave = () => {
    update({
      variables: {
        input: {
          provider,
          baseURL,
          model,
          // Send the key only when the operator typed a new one; otherwise keep
          // the stored key (null = keep).
          apiKey: apiKey === "" ? null : apiKey,
          enabled,
        },
      },
    });
  };

  return (
    <Box sx={{ maxWidth: 640 }}>
      <Snackbar open={saved} autoHideDuration={3000} onClose={() => setSaved(false)}>
        <Alert onClose={() => setSaved(false)} severity="info">
          AI assistant settings saved.
        </Alert>
      </Snackbar>

      <Typography variant="h5" sx={{ mb: 1 }}>
        AI Assistant
      </Typography>
      <Typography paragraph color="text.secondary">
        Connect an OpenAI-compatible LLM (DeepSeek, OpenAI, a local Ollama, or any compatible endpoint) to power the
        assistant and its agent tools.
      </Typography>

      <FormControlLabel
        control={<Switch checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />}
        label="Enable AI assistant"
        sx={{ mb: 1 }}
      />

      <TextField
        select
        fullWidth
        size="small"
        label="Provider"
        value={provider}
        onChange={(e) => onPreset(e.target.value)}
        sx={{ mb: 2 }}
      >
        <MenuItem value="deepseek">DeepSeek</MenuItem>
        <MenuItem value="openai">OpenAI</MenuItem>
        <MenuItem value="ollama">Ollama (local)</MenuItem>
        <MenuItem value="custom">Custom</MenuItem>
      </TextField>

      <TextField
        fullWidth
        size="small"
        label="Base URL"
        placeholder="https://api.deepseek.com/v1"
        value={baseURL}
        onChange={(e) => setBaseURL(e.target.value)}
        sx={{ mb: 2 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />

      <TextField
        fullWidth
        size="small"
        label="Model"
        placeholder="deepseek-chat"
        value={model}
        onChange={(e) => setModel(e.target.value)}
        sx={{ mb: 2 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />

      <TextField
        fullWidth
        size="small"
        type="password"
        label="API key"
        placeholder={hasApiKey ? "•••••••• (stored — leave blank to keep)" : "Required for hosted providers"}
        value={apiKey}
        onChange={(e) => setApiKey(e.target.value)}
        sx={{ mb: 2 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />

      <Button
        variant="contained"
        onClick={onSave}
        disabled={loading}
        startIcon={loading ? <CircularProgress size={20} /> : undefined}
      >
        Save
      </Button>

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error.message}
        </Alert>
      )}
    </Box>
  );
}
