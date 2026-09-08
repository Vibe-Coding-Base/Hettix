import SendIcon from "@mui/icons-material/Send";
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  IconButton,
  Paper,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
  Typography,
} from "@mui/material";
import { useState } from "react";

import HttpqlHighlight from "lib/components/HttpqlHighlight";
import { AgentMode, useRunAgentMutation } from "lib/graphql/generated";

interface Action {
  tool: string;
  input: string;
  output: string;
  denied: boolean;
}

interface Turn {
  role: "user" | "assistant";
  text: string;
  actions?: Action[];
}

const MODES: { value: AgentMode; label: string; hint: string }[] = [
  { value: AgentMode.Ask, label: "Ask", hint: "Read-only: investigate traffic, no changes." },
  { value: AgentMode.Assist, label: "Assist", hint: "Suggests actions; each change needs approval." },
  {
    value: AgentMode.Auto,
    label: "Auto-pilot",
    hint: "Plans and acts autonomously toward your objective, within scope and policy guardrails.",
  },
];

export default function Assistant(): JSX.Element {
  const [mode, setMode] = useState<AgentMode>(AgentMode.Ask);
  const [input, setInput] = useState("");
  const [turns, setTurns] = useState<Turn[]>([]);
  const [runAgent, { loading, error }] = useRunAgentMutation();

  const send = async () => {
    const message = input.trim();
    if (message === "" || loading) {
      return;
    }

    setTurns((prev) => [...prev, { role: "user", text: message }]);
    setInput("");

    try {
      const { data } = await runAgent({ variables: { input: { message, mode } } });
      if (data?.runAgent) {
        setTurns((prev) => [...prev, { role: "assistant", text: data.runAgent.reply, actions: data.runAgent.actions }]);
      }
    } catch {
      // The error is surfaced from the mutation's `error` below.
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%", maxWidth: 900 }}>
      <ToggleButtonGroup
        exclusive
        size="small"
        value={mode}
        onChange={(_, next) => next && setMode(next)}
        sx={{ mb: 2 }}
      >
        {MODES.map((m) => (
          <ToggleButton key={m.value} value={m.value}>
            <Tooltip title={m.hint}>
              <span>{m.label}</span>
            </Tooltip>
          </ToggleButton>
        ))}
      </ToggleButtonGroup>

      <Box sx={{ flex: 1, overflowY: "auto", mb: 2 }}>
        {turns.length === 0 && (
          <Typography color="text.secondary">
            Ask about the captured traffic, e.g. “Which requests returned a 500?” or “Find requests sending credentials
            over plain HTTP.”
          </Typography>
        )}

        {turns.map((turn, i) => (
          <TurnView key={i} turn={turn} />
        ))}

        {loading && (
          <Box sx={{ display: "flex", alignItems: "center", gap: 1, mt: 1 }}>
            <CircularProgress size={16} />
            <Typography variant="body2" color="text.secondary">
              Hetty is working…
            </Typography>
          </Box>
        )}

        {error && (
          <Alert severity="error" sx={{ mt: 1 }}>
            {error.message}
          </Alert>
        )}
      </Box>

      <Box
        component="form"
        onSubmit={(e) => {
          e.preventDefault();
          send();
        }}
        sx={{ display: "flex", gap: 1 }}
      >
        <TextField
          fullWidth
          size="small"
          placeholder="Ask Hetty about the traffic…"
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <Tooltip title="Send">
          <span>
            <IconButton type="submit" color="primary" disabled={loading || input.trim() === ""}>
              <SendIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>
    </Box>
  );
}

function TurnView({ turn }: { turn: Turn }): JSX.Element {
  const isUser = turn.role === "user";

  return (
    <Paper
      variant="outlined"
      sx={{
        p: 1.5,
        mb: 1.5,
        ml: isUser ? "auto" : 0,
        maxWidth: "85%",
        bgcolor: isUser ? "action.selected" : "background.paper",
      }}
    >
      <Typography variant="caption" color="text.secondary">
        {isUser ? "You" : "Hetty"}
      </Typography>

      {turn.actions?.map((action, i) => (
        <ActionView key={i} action={action} />
      ))}

      {turn.text !== "" && (
        <Typography sx={{ whiteSpace: "pre-wrap", mt: turn.actions?.length ? 1 : 0 }}>{turn.text}</Typography>
      )}
    </Paper>
  );
}

function ActionView({ action }: { action: Action }): JSX.Element {
  return (
    <Box sx={{ my: 1, p: 1, borderRadius: 1, bgcolor: "action.hover" }}>
      <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 0.5 }}>
        <Chip size="small" label={action.tool} color={action.denied ? "warning" : "default"} variant="outlined" />
        {action.denied && (
          <Typography variant="caption" color="warning.main">
            denied
          </Typography>
        )}
      </Box>

      {action.tool === "search_traffic" ? (
        <SearchArgs input={action.input} />
      ) : (
        <Typography component="pre" sx={{ m: 0, fontSize: "0.75rem", whiteSpace: "pre-wrap" }}>
          {action.input}
        </Typography>
      )}

      <Typography
        component="pre"
        sx={{ m: 0, mt: 0.5, fontSize: "0.75rem", color: "text.secondary", whiteSpace: "pre-wrap" }}
      >
        {action.output}
      </Typography>
    </Box>
  );
}

function SearchArgs({ input }: { input: string }): JSX.Element {
  let query = input;
  try {
    query = (JSON.parse(input) as { query?: string }).query ?? input;
  } catch {
    // Fall back to the raw arguments if they aren't valid JSON.
  }

  return <HttpqlHighlight query={query} />;
}
