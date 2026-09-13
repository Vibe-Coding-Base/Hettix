import DownloadIcon from "@mui/icons-material/Download";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  FormControlLabel,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { useState } from "react";

import { downloadText, toCSV } from "lib/download";
import { useSitemapQuery } from "lib/graphql/generated";

type Entry = {
  host: string;
  path: string;
  methods: string[];
  statusCodes: number[];
  count: number;
  tags: string[];
};

// PathCell renders a path with any plugin-discovery tags.
function PathCell({ entry }: { entry: Entry }): JSX.Element {
  return (
    <Box sx={{ display: "flex", alignItems: "center", gap: 0.75, flexWrap: "wrap" }}>
      <Box component="span" sx={{ fontFamily: "monospace", fontSize: 13, wordBreak: "break-all" }}>
        {entry.path || "/"}
      </Box>
      {entry.tags.map((tag) => (
        <Chip key={tag} size="small" color="info" variant="outlined" label={tag} sx={{ height: 18, fontSize: 10 }} />
      ))}
    </Box>
  );
}

type Kind = "endpoint" | "js" | "static";

const JS_EXT = /\.(m?js|jsx)$/i;
const STATIC_EXT = /\.(css|png|jpe?g|gif|svg|ico|webp|avif|bmp|woff2?|ttf|eot|otf|map|mp4|webm|mp3|wav|pdf)$/i;

// classify decides a path's role for the pentester: application endpoints, the
// JavaScript that drives them, or ignorable static assets.
function classify(path: string): Kind {
  const p = path.split("?")[0];
  if (JS_EXT.test(p)) {
    return "js";
  }
  if (STATIC_EXT.test(p)) {
    return "static";
  }
  return "endpoint";
}

export default function Sitemap(): JSX.Element {
  const { data, loading, error } = useSitemapQuery({ pollInterval: 5000 });
  const [showStatic, setShowStatic] = useState(false);

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  const entries = data?.sitemap ?? [];

  if (loading && entries.length === 0) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (entries.length === 0) {
    return (
      <Typography color="text.secondary">
        No traffic captured yet. Browse through the proxy to build the sitemap.
      </Typography>
    );
  }

  const byHost = new Map<string, Entry[]>();
  for (const entry of entries) {
    if (!showStatic && classify(entry.path) === "static") {
      continue;
    }
    const list = byHost.get(entry.host) ?? [];
    list.push(entry);
    byHost.set(entry.host, list);
  }

  const hosts = [...byHost.keys()].sort();

  const exportCSV = () => {
    const rows = entries.map((e) => [
      e.host,
      e.path,
      e.methods.join(" "),
      e.statusCodes.join(" "),
      e.count,
      e.tags.join(" "),
    ]);
    downloadText("sitemap.csv", toCSV(["Host", "Path", "Methods", "Status codes", "Count", "Tags"], rows), "text/csv");
  };

  return (
    <Box sx={{ maxWidth: 960 }}>
      <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", mb: 1 }}>
        <FormControlLabel
          control={<Switch size="small" checked={showStatic} onChange={(e) => setShowStatic(e.target.checked)} />}
          label="Show static assets"
        />
        <Button size="small" startIcon={<DownloadIcon />} onClick={exportCSV}>
          Export CSV
        </Button>
      </Box>
      {hosts.map((host) => {
        const hostEntries = byHost.get(host) ?? [];
        const endpoints = hostEntries.filter((e) => classify(e.path) === "endpoint");
        const scripts = hostEntries.filter((e) => classify(e.path) === "js");
        const statics = hostEntries.filter((e) => classify(e.path) === "static");

        return (
          <Accordion key={host} defaultExpanded={hosts.length <= 3}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography sx={{ fontFamily: "monospace" }}>{host}</Typography>
              <Chip size="small" label={`${endpoints.length} endpoints`} sx={{ ml: 1 }} />
              {scripts.length > 0 && (
                <Chip size="small" variant="outlined" label={`${scripts.length} JS`} sx={{ ml: 1 }} />
              )}
            </AccordionSummary>
            <AccordionDetails sx={{ p: 0 }}>
              {endpoints.length > 0 && <EndpointTable entries={endpoints} />}
              {scripts.length > 0 && <PathList title="JavaScript" entries={scripts} />}
              {showStatic && statics.length > 0 && <PathList title="Static assets" entries={statics} />}
            </AccordionDetails>
          </Accordion>
        );
      })}
    </Box>
  );
}

function EndpointTable({ entries }: { entries: Entry[] }): JSX.Element {
  return (
    <Table size="small">
      <TableHead>
        <TableRow>
          <TableCell>Path</TableCell>
          <TableCell>Methods</TableCell>
          <TableCell>Status</TableCell>
          <TableCell align="right">Count</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {entries.map((entry) => (
          <TableRow key={entry.path} hover>
            <TableCell>
              <PathCell entry={entry} />
            </TableCell>
            <TableCell>
              <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                {entry.methods.map((m) => (
                  <Chip key={m} size="small" variant="outlined" label={m} />
                ))}
              </Box>
            </TableCell>
            <TableCell>
              <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                {entry.statusCodes.map((code) => (
                  <StatusChip key={code} code={code} />
                ))}
              </Box>
            </TableCell>
            <TableCell align="right">{entry.count}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function PathList({ title, entries }: { title: string; entries: Entry[] }): JSX.Element {
  return (
    <Box>
      <Typography variant="overline" color="text.secondary" sx={{ px: 2, pt: 1, display: "block" }}>
        {title} ({entries.length})
      </Typography>
      <Table size="small">
        <TableBody>
          {entries.map((entry) => (
            <TableRow key={entry.path} hover>
              <TableCell>
                <PathCell entry={entry} />
              </TableCell>
              <TableCell align="right" sx={{ width: 64 }}>
                {entry.count}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Box>
  );
}

function StatusChip({ code }: { code: number }): JSX.Element {
  const color = code >= 500 ? "error" : code >= 400 ? "warning" : code >= 300 ? "info" : "success";
  return <Chip size="small" color={color} label={code} />;
}
