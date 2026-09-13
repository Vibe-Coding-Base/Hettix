import { useApolloClient } from "@apollo/client";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import UploadIcon from "@mui/icons-material/Upload";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Snackbar,
  Stack,
  Switch,
  Typography,
} from "@mui/material";
import { useRef, useState } from "react";

import Editor from "lib/components/Editor";
import {
  PluginFieldsFragment,
  PluginSourceDocument,
  PluginSourceQuery,
  PluginsDocument,
  useDeletePluginMutation,
  useInstallPluginMutation,
  usePluginsQuery,
  useSetPluginEnabledMutation,
  useUpdatePluginMutation,
} from "lib/graphql/generated";

export default function Plugins(): JSX.Element {
  const { data, loading, error } = usePluginsQuery();
  const [setEnabled] = useSetPluginEnabledMutation();
  const [installPlugin] = useInstallPluginMutation({ refetchQueries: [{ query: PluginsDocument }] });
  const [deletePlugin] = useDeletePluginMutation({ refetchQueries: [{ query: PluginsDocument }] });

  const fileInput = useRef<HTMLInputElement>(null);
  const [editing, setEditing] = useState<PluginFieldsFragment | null>(null);
  const [message, setMessage] = useState("");

  const install = async (file: File) => {
    try {
      await installPlugin({ variables: { content: await file.text() } });
      setMessage(`Installed ${file.name}.`);
    } catch (e) {
      setMessage(e instanceof Error ? e.message : String(e));
    }
  };

  const remove = async (plugin: PluginFieldsFragment) => {
    if (!window.confirm(`Delete plugin "${plugin.name}"? This removes ${plugin.filename} from disk.`)) {
      return;
    }
    try {
      await deletePlugin({ variables: { id: plugin.id } });
      setMessage(`Deleted ${plugin.name}.`);
    } catch (e) {
      setMessage(e instanceof Error ? e.message : String(e));
    }
  };

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  if (loading && !data) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  const plugins = data?.plugins ?? [];

  return (
    <Box sx={{ maxWidth: 760 }}>
      <Box sx={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 2, mb: 2 }}>
        <Typography color="text.secondary">
          Plugins are JavaScript files that observe proxied responses and can add discovered endpoints to the sitemap
          and record findings. They live as editable files in your plugins directory.
        </Typography>
        <Button
          variant="contained"
          startIcon={<UploadIcon />}
          sx={{ flexShrink: 0 }}
          onClick={() => fileInput.current?.click()}
        >
          Install plugin
        </Button>
        <input
          ref={fileInput}
          type="file"
          accept=".js,text/javascript"
          hidden
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) {
              install(file);
            }
            e.target.value = "";
          }}
        />
      </Box>

      {plugins.map((plugin) => (
        <Card key={plugin.id} variant="outlined" sx={{ mb: 1.5 }}>
          <CardContent sx={{ "&:last-child": { pb: 2 } }}>
            <Box sx={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 2 }}>
              <Box>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5, flexWrap: "wrap" }}>
                  <Typography variant="subtitle1">{plugin.name}</Typography>
                  {plugin.version && <Chip size="small" label={`v${plugin.version}`} variant="outlined" />}
                  {plugin.builtin && <Chip size="small" label="built-in" color="info" variant="outlined" />}
                  {plugin.capabilities.map((cap) => (
                    <Chip key={cap} size="small" label={cap} />
                  ))}
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {plugin.description}
                </Typography>
              </Box>
              <FormControlLabel
                sx={{ flexShrink: 0, m: 0 }}
                control={
                  <Switch
                    checked={plugin.enabled}
                    onChange={(e) =>
                      setEnabled({
                        variables: { id: plugin.id, enabled: e.target.checked },
                        optimisticResponse: {
                          setPluginEnabled: { __typename: "Plugin", ...plugin, enabled: e.target.checked },
                        },
                      })
                    }
                  />
                }
                label={plugin.enabled ? "Enabled" : "Disabled"}
              />
            </Box>
            <Stack direction="row" spacing={1} sx={{ mt: 1.5 }}>
              <Button size="small" startIcon={<EditIcon />} onClick={() => setEditing(plugin)}>
                Edit source
              </Button>
              <Button size="small" color="error" startIcon={<DeleteIcon />} onClick={() => remove(plugin)}>
                Delete
              </Button>
            </Stack>
          </CardContent>
        </Card>
      ))}

      {editing && <EditPluginDialog plugin={editing} onClose={() => setEditing(null)} onSaved={setMessage} />}

      <Snackbar
        open={message !== ""}
        autoHideDuration={4000}
        onClose={() => setMessage("")}
        anchorOrigin={{ horizontal: "center", vertical: "bottom" }}
      >
        <Alert onClose={() => setMessage("")} severity="info">
          {message}
        </Alert>
      </Snackbar>
    </Box>
  );
}

interface EditPluginDialogProps {
  plugin: PluginFieldsFragment;
  onClose: () => void;
  onSaved: (message: string) => void;
}

function EditPluginDialog({ plugin, onClose, onSaved }: EditPluginDialogProps): JSX.Element {
  const client = useApolloClient();
  const [update, { loading: saving }] = useUpdatePluginMutation();
  const [source, setSource] = useState<string | null>(null);
  const [error, setError] = useState("");

  if (source === null) {
    client
      .query<PluginSourceQuery>({ query: PluginSourceDocument, variables: { id: plugin.id }, fetchPolicy: "no-cache" })
      .then(({ data }) => setSource(data.pluginSource))
      .catch((e) => {
        setError(e instanceof Error ? e.message : String(e));
        setSource("");
      });
  }

  const save = async () => {
    setError("");
    try {
      await update({ variables: { id: plugin.id, content: source ?? "" } });
      onSaved(`Saved ${plugin.name}.`);
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <Dialog open fullWidth maxWidth="md" onClose={onClose}>
      <DialogTitle>Edit {plugin.name}</DialogTitle>
      <DialogContent dividers sx={{ p: 0, height: "60vh" }}>
        {error && <Alert severity="error">{error}</Alert>}
        {source === null ? (
          <Box sx={{ display: "flex", justifyContent: "center", p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Editor
            content={source}
            language="javascript"
            monacoOptions={{ readOnly: false }}
            onChange={(value) => setSource(value ?? "")}
          />
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={save} disabled={saving || source === null}>
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
}
