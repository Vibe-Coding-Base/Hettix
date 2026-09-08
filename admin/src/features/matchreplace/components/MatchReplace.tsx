import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Divider,
  FormControlLabel,
  IconButton,
  MenuItem,
  Paper,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import { MatchReplacePhase, useMatchReplaceRulesQuery, useSetMatchReplaceRulesMutation } from "lib/graphql/generated";

interface EditableRule {
  name: string;
  enabled: boolean;
  phase: MatchReplacePhase;
  condition: string;
  headerName: string;
  headerValue: string;
  removeHeader: boolean;
  bodyMatcher: string;
  bodyReplacement: string;
}

function blankRule(): EditableRule {
  return {
    name: "",
    enabled: true,
    phase: MatchReplacePhase.Request,
    condition: "",
    headerName: "",
    headerValue: "",
    removeHeader: false,
    bodyMatcher: "",
    bodyReplacement: "",
  };
}

export default function MatchReplace(): JSX.Element {
  const { data } = useMatchReplaceRulesQuery();
  const [rules, setRules] = useState<EditableRule[]>([]);
  const [save, { loading: saving, error }] = useSetMatchReplaceRulesMutation();

  useEffect(() => {
    if (data?.matchReplaceRules) {
      setRules(
        data.matchReplaceRules.map((r) => ({
          name: r.name,
          enabled: r.enabled,
          phase: r.phase,
          condition: r.condition ?? "",
          headerName: r.headerName ?? "",
          headerValue: r.headerValue ?? "",
          removeHeader: r.removeHeader,
          bodyMatcher: r.bodyMatcher ?? "",
          bodyReplacement: r.bodyReplacement ?? "",
        }))
      );
    }
  }, [data]);

  const update = (i: number, patch: Partial<EditableRule>) =>
    setRules((rs) => rs.map((r, j) => (j === i ? { ...r, ...patch } : r)));
  const remove = (i: number) => setRules((rs) => rs.filter((_, j) => j !== i));
  const add = () => setRules((rs) => [...rs, blankRule()]);

  const onSave = () => {
    save({
      variables: {
        rules: rules.map((r) => ({
          name: r.name,
          enabled: r.enabled,
          phase: r.phase,
          condition: r.condition,
          headerName: r.headerName,
          headerValue: r.headerValue,
          removeHeader: r.removeHeader,
          bodyMatcher: r.bodyMatcher,
          bodyReplacement: r.bodyReplacement,
        })),
      },
    });
  };

  return (
    <Box sx={{ maxWidth: 720 }}>
      <Typography color="text.secondary" sx={{ mb: 2 }}>
        Rules rewrite proxied traffic. Each runs on requests or responses, applies when its optional HTTPQL condition
        matches, and can set or remove a header and/or regexp-replace the body.
      </Typography>

      {rules.map((rule, i) => (
        <RuleCard key={i} rule={rule} onChange={(patch) => update(i, patch)} onDelete={() => remove(i)} />
      ))}

      <Box sx={{ display: "flex", gap: 1, mt: 1 }}>
        <Button startIcon={<AddIcon />} onClick={add}>
          Add rule
        </Button>
        <Button variant="contained" onClick={onSave} disabled={saving}>
          Save
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error.message}
        </Alert>
      )}
    </Box>
  );
}

function RuleCard({
  rule,
  onChange,
  onDelete,
}: {
  rule: EditableRule;
  onChange: (patch: Partial<EditableRule>) => void;
  onDelete: () => void;
}): JSX.Element {
  return (
    <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: "flex", gap: 1, alignItems: "center", mb: 1 }}>
        <TextField
          size="small"
          label="Name"
          value={rule.name}
          onChange={(e) => onChange({ name: e.target.value })}
          sx={{ flex: 1 }}
        />
        <TextField
          size="small"
          select
          label="Phase"
          value={rule.phase}
          onChange={(e) => onChange({ phase: e.target.value as MatchReplacePhase })}
          sx={{ width: 140 }}
        >
          <MenuItem value={MatchReplacePhase.Request}>Request</MenuItem>
          <MenuItem value={MatchReplacePhase.Response}>Response</MenuItem>
        </TextField>
        <FormControlLabel
          control={<Switch checked={rule.enabled} onChange={(e) => onChange({ enabled: e.target.checked })} />}
          label="On"
        />
        <IconButton onClick={onDelete} aria-label="Delete rule">
          <DeleteIcon />
        </IconButton>
      </Box>

      <TextField
        size="small"
        fullWidth
        label="Condition (HTTPQL, optional)"
        placeholder='e.g. req.host cont "target.com"'
        value={rule.condition}
        onChange={(e) => onChange({ condition: e.target.value })}
        sx={{ mb: 1.5 }}
      />

      <Typography variant="caption" color="text.secondary">
        Header
      </Typography>
      <Box sx={{ display: "flex", gap: 1, alignItems: "center", mb: 1.5 }}>
        <TextField
          size="small"
          label="Header name"
          value={rule.headerName}
          onChange={(e) => onChange({ headerName: e.target.value })}
        />
        <TextField
          size="small"
          label="Header value"
          value={rule.headerValue}
          onChange={(e) => onChange({ headerValue: e.target.value })}
          disabled={rule.removeHeader}
          sx={{ flex: 1 }}
        />
        <FormControlLabel
          control={
            <Checkbox checked={rule.removeHeader} onChange={(e) => onChange({ removeHeader: e.target.checked })} />
          }
          label="Remove"
        />
      </Box>

      <Divider sx={{ mb: 1.5 }} />

      <Typography variant="caption" color="text.secondary">
        Body (regexp replace)
      </Typography>
      <Box sx={{ display: "flex", gap: 1 }}>
        <TextField
          size="small"
          label="Match (regexp)"
          value={rule.bodyMatcher}
          onChange={(e) => onChange({ bodyMatcher: e.target.value })}
          sx={{ flex: 1 }}
        />
        <TextField
          size="small"
          label="Replace with"
          value={rule.bodyReplacement}
          onChange={(e) => onChange({ bodyReplacement: e.target.value })}
          sx={{ flex: 1 }}
        />
      </Box>
    </Paper>
  );
}
