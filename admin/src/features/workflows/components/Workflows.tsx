import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import { Alert, Box, Button, Chip, Divider, IconButton, MenuItem, Paper, TextField, Typography } from "@mui/material";
import { useEffect, useState } from "react";

import {
  WorkflowStepType,
  WorkflowsDocument,
  useDeleteWorkflowMutation,
  useRunWorkflowMutation,
  useSaveWorkflowMutation,
  useWorkflowsQuery,
} from "lib/graphql/generated";

type Step = {
  type: WorkflowStepType;
  query: string;
  name: string;
  method: string;
  url: string;
  body: string;
  payloads: string;
  title: string;
  description: string;
  severity: string;
};

function blankStep(type: WorkflowStepType): Step {
  return {
    type,
    query: "",
    name: "",
    method: "GET",
    url: "",
    body: "",
    payloads: "",
    title: "",
    description: "",
    severity: "medium",
  };
}

export default function Workflows(): JSX.Element {
  const { data } = useWorkflowsQuery({ pollInterval: 3000 });
  const [selected, setSelected] = useState<string | null>(null);

  const workflows = data?.workflows ?? [];

  return (
    <Box sx={{ display: "flex", gap: 2, height: "calc(100vh - 120px)" }}>
      <Paper variant="outlined" sx={{ width: 260, flexShrink: 0, overflow: "auto" }}>
        <Box sx={{ p: 1 }}>
          <Button fullWidth startIcon={<AddIcon />} onClick={() => setSelected("new")}>
            New workflow
          </Button>
        </Box>
        <Divider />
        {workflows.map((wf) => (
          <Box
            key={wf.id}
            onClick={() => setSelected(wf.id)}
            sx={{
              p: 1.5,
              cursor: "pointer",
              borderBottom: 1,
              borderColor: "divider",
              bgcolor: selected === wf.id ? "action.selected" : undefined,
            }}
          >
            <Typography noWrap>{wf.name}</Typography>
            <Typography variant="caption" color="text.secondary">
              {wf.steps.length} step{wf.steps.length === 1 ? "" : "s"}
            </Typography>
          </Box>
        ))}
      </Paper>

      <Box sx={{ flex: 1, overflow: "auto" }}>
        {selected ? (
          <Editor key={selected} id={selected === "new" ? null : selected} onSaved={setSelected} />
        ) : (
          <Typography color="text.secondary" sx={{ p: 2 }}>
            Select a workflow or create a new one.
          </Typography>
        )}
      </Box>
    </Box>
  );
}

function Editor({ id, onSaved }: { id: string | null; onSaved: (id: string) => void }): JSX.Element {
  const { data } = useWorkflowsQuery();
  const [name, setName] = useState("");
  const [steps, setSteps] = useState<Step[]>([]);

  const [save, { loading: saving, error: saveError }] = useSaveWorkflowMutation({
    refetchQueries: [{ query: WorkflowsDocument }],
    onCompleted({ saveWorkflow }) {
      onSaved(saveWorkflow.id);
    },
  });
  const [remove] = useDeleteWorkflowMutation({ refetchQueries: [{ query: WorkflowsDocument }] });
  const [run, { data: runData, loading: running, error: runError }] = useRunWorkflowMutation();

  useEffect(() => {
    const wf = data?.workflows.find((w) => w.id === id);
    if (wf) {
      setName(wf.name);
      setSteps(
        wf.steps.map((s) => ({
          type: s.type,
          query: s.query ?? "",
          name: s.name ?? "",
          method: s.method ?? "GET",
          url: s.url ?? "",
          body: s.body ?? "",
          payloads: (s.payloads ?? []).join("\n"),
          title: s.title ?? "",
          description: s.description ?? "",
          severity: s.severity ?? "medium",
        }))
      );
    } else if (id === null) {
      setName("");
      setSteps([]);
    }
  }, [data, id]);

  const update = (i: number, patch: Partial<Step>) =>
    setSteps((ss) => ss.map((s, j) => (j === i ? { ...s, ...patch } : s)));
  const removeStep = (i: number) => setSteps((ss) => ss.filter((_, j) => j !== i));

  const onSave = () => {
    save({
      variables: {
        input: {
          id,
          name: name || "Workflow",
          steps: steps.map((s) => ({
            type: s.type,
            query: s.query,
            name: s.name,
            method: s.method,
            url: s.url,
            body: s.body,
            payloads: s.payloads
              .split("\n")
              .map((p) => p.trim())
              .filter(Boolean),
            title: s.title,
            description: s.description,
            severity: s.severity,
          })),
        },
      },
    });
  };

  return (
    <Box sx={{ maxWidth: 760 }}>
      <Box sx={{ display: "flex", gap: 1, alignItems: "center", mb: 2 }}>
        <TextField size="small" label="Name" value={name} onChange={(e) => setName(e.target.value)} sx={{ flex: 1 }} />
        <Button variant="contained" onClick={onSave} disabled={saving}>
          Save
        </Button>
        {id && (
          <>
            <Button startIcon={<PlayArrowIcon />} onClick={() => run({ variables: { id } })} disabled={running}>
              Run
            </Button>
            <IconButton aria-label="Delete workflow" onClick={() => remove({ variables: { id } })}>
              <DeleteIcon />
            </IconButton>
          </>
        )}
      </Box>

      {steps.map((step, i) => (
        <StepCard key={i} step={step} onChange={(patch) => update(i, patch)} onDelete={() => removeStep(i)} />
      ))}

      <Box sx={{ display: "flex", gap: 1, mt: 1 }}>
        <Button size="small" onClick={() => setSteps((ss) => [...ss, blankStep(WorkflowStepType.Search)])}>
          + Search
        </Button>
        <Button size="small" onClick={() => setSteps((ss) => [...ss, blankStep(WorkflowStepType.Fuzz)])}>
          + Fuzz
        </Button>
        <Button size="small" onClick={() => setSteps((ss) => [...ss, blankStep(WorkflowStepType.Finding)])}>
          + Finding
        </Button>
      </Box>

      {saveError && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {saveError.message}
        </Alert>
      )}
      {runError && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {runError.message}
        </Alert>
      )}

      {runData && (
        <Paper variant="outlined" sx={{ mt: 2, p: 2 }}>
          <Typography variant="subtitle2" sx={{ mb: 1 }}>
            Run log
          </Typography>
          {runData.runWorkflow.map((res, i) => (
            <Box key={i} sx={{ mb: 1 }}>
              <Chip size="small" label={res.type.toLowerCase()} sx={{ mr: 1 }} />
              {res.error ? (
                <Typography component="span" color="error" variant="body2">
                  {res.error}
                </Typography>
              ) : (
                <Typography component="span" variant="body2">
                  {res.output}
                </Typography>
              )}
            </Box>
          ))}
        </Paper>
      )}
    </Box>
  );
}

function StepCard({
  step,
  onChange,
  onDelete,
}: {
  step: Step;
  onChange: (patch: Partial<Step>) => void;
  onDelete: () => void;
}): JSX.Element {
  return (
    <Paper variant="outlined" sx={{ p: 2, mb: 1.5 }}>
      <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
        <Chip size="small" color="primary" label={step.type.toLowerCase()} />
        <Box sx={{ flex: 1 }} />
        <IconButton size="small" aria-label="Delete step" onClick={onDelete}>
          <DeleteIcon fontSize="small" />
        </IconButton>
      </Box>

      {step.type === WorkflowStepType.Search && (
        <TextField
          size="small"
          fullWidth
          label="HTTPQL query"
          placeholder='req.path cont "login"'
          value={step.query}
          onChange={(e) => onChange({ query: e.target.value })}
          InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
        />
      )}

      {step.type === WorkflowStepType.Fuzz && (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <Box sx={{ display: "flex", gap: 1 }}>
            <TextField
              size="small"
              label="Name"
              value={step.name}
              onChange={(e) => onChange({ name: e.target.value })}
              sx={{ flex: 1 }}
            />
            <TextField
              size="small"
              select
              label="Method"
              value={step.method}
              onChange={(e) => onChange({ method: e.target.value })}
              sx={{ width: 110 }}
            >
              {["GET", "POST", "PUT", "DELETE", "PATCH"].map((m) => (
                <MenuItem key={m} value={m}>
                  {m}
                </MenuItem>
              ))}
            </TextField>
          </Box>
          <TextField
            size="small"
            label="URL (mark insertion with §)"
            placeholder="https://target.com/search?q=§"
            value={step.url}
            onChange={(e) => onChange({ url: e.target.value })}
            InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
          />
          <TextField
            size="small"
            multiline
            minRows={2}
            label="Payloads (one per line)"
            value={step.payloads}
            onChange={(e) => onChange({ payloads: e.target.value })}
            InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
          />
        </Box>
      )}

      {step.type === WorkflowStepType.Finding && (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <Box sx={{ display: "flex", gap: 1 }}>
            <TextField
              size="small"
              label="Title"
              value={step.title}
              onChange={(e) => onChange({ title: e.target.value })}
              sx={{ flex: 1 }}
            />
            <TextField
              size="small"
              select
              label="Severity"
              value={step.severity}
              onChange={(e) => onChange({ severity: e.target.value })}
              sx={{ width: 130 }}
            >
              {["info", "low", "medium", "high", "critical"].map((s) => (
                <MenuItem key={s} value={s}>
                  {s}
                </MenuItem>
              ))}
            </TextField>
          </Box>
          <TextField
            size="small"
            multiline
            minRows={2}
            label="Description (supports ${count}, ${hosts})"
            value={step.description}
            onChange={(e) => onChange({ description: e.target.value })}
          />
        </Box>
      )}
    </Paper>
  );
}
