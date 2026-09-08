import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Chip,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";

import { useSitemapQuery } from "lib/graphql/generated";

type Entry = {
  host: string;
  path: string;
  methods: string[];
  statusCodes: number[];
  count: number;
};

export default function Sitemap(): JSX.Element {
  const { data, loading, error } = useSitemapQuery({ pollInterval: 5000 });

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
    const list = byHost.get(entry.host) ?? [];
    list.push(entry);
    byHost.set(entry.host, list);
  }

  const hosts = [...byHost.keys()].sort();

  return (
    <Box sx={{ maxWidth: 960 }}>
      {hosts.map((host) => (
        <Accordion key={host} defaultExpanded={hosts.length <= 3}>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Typography sx={{ fontFamily: "monospace" }}>{host}</Typography>
            <Chip size="small" label={byHost.get(host)?.length ?? 0} sx={{ ml: 1 }} />
          </AccordionSummary>
          <AccordionDetails sx={{ p: 0 }}>
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
                {byHost.get(host)?.map((entry) => (
                  <TableRow key={entry.path} hover>
                    <TableCell sx={{ fontFamily: "monospace", fontSize: 13, wordBreak: "break-all" }}>
                      {entry.path || "/"}
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
          </AccordionDetails>
        </Accordion>
      ))}
    </Box>
  );
}

function StatusChip({ code }: { code: number }): JSX.Element {
  const color = code >= 500 ? "error" : code >= 400 ? "warning" : code >= 300 ? "info" : "success";
  return <Chip size="small" color={color} label={code} />;
}
