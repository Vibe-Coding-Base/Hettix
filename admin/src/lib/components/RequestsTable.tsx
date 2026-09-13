import {
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  TableSortLabel,
  styled,
  TableCellProps,
  TableRowProps,
} from "@mui/material";
import { useMemo, useState } from "react";

import HttpStatusIcon from "./HttpStatusIcon";

import { HttpMethod } from "lib/graphql/generated";

const baseCellStyle = {
  whiteSpace: "nowrap",
  overflow: "hidden",
  textOverflow: "ellipsis",
} as const;

const SeqTableCell = styled(TableCell)<TableCellProps>(() => ({
  ...baseCellStyle,
  width: "60px",
}));

const MethodTableCell = styled(TableCell)<TableCellProps>(() => ({
  ...baseCellStyle,
  width: "100px",
}));

const OriginTableCell = styled(TableCell)<TableCellProps>(() => ({
  ...baseCellStyle,
  maxWidth: "100px",
}));

const PathTableCell = styled(TableCell)<TableCellProps>(() => ({
  ...baseCellStyle,
  maxWidth: "200px",
}));

const StatusTableCell = styled(TableCell)<TableCellProps>(() => ({
  ...baseCellStyle,
  width: "100px",
}));

const RequestTableRow = styled(TableRow)<TableRowProps>(() => ({
  "&:hover": {
    cursor: "pointer",
  },
}));

interface HttpRequest {
  id: string;
  url: string;
  method: HttpMethod;
  response?: HttpResponse | null;
}

interface HttpResponse {
  statusCode: number;
  statusReason: string;
  body?: string;
}

interface Props {
  requests: HttpRequest[];
  activeRowId?: string;
  actionsCell?: (id: string) => JSX.Element;
  onRowClick?: (id: string) => void;
  onContextMenu?: (e: React.MouseEvent, id: string) => void;
}

function decodeURLPart(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

function parseURLForDisplay(url: string): { origin: string; path: string } {
  try {
    const { origin, pathname, search, hash } = new URL(url);
    return {
      origin,
      path: decodeURLPart(pathname + search + hash),
    };
  } catch {
    return {
      origin: "",
      path: url,
    };
  }
}

type SortOrder = "asc" | "desc";

export default function RequestsTable(props: Props): JSX.Element {
  const { requests, activeRowId, actionsCell, onRowClick, onContextMenu } = props;

  // ULIDs are lexicographically time-ordered, so ascending id order matches the
  // order requests arrived. The "#" column exposes that order and lets the user
  // flip between oldest- and newest-first.
  const [order, setOrder] = useState<SortOrder>("asc");

  const seqById = useMemo(() => {
    const byId = [...requests].sort((a, b) => (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));
    return new Map(byId.map((req, index) => [req.id, index + 1]));
  }, [requests]);

  const sorted = useMemo(() => {
    const byId = [...requests].sort((a, b) => (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));
    return order === "asc" ? byId : byId.reverse();
  }, [requests, order]);

  return (
    <TableContainer sx={{ overflowX: "initial" }}>
      <Table size="small" stickyHeader>
        <TableHead>
          <TableRow>
            <SeqTableCell sortDirection={order}>
              <TableSortLabel active direction={order} onClick={() => setOrder((o) => (o === "asc" ? "desc" : "asc"))}>
                #
              </TableSortLabel>
            </SeqTableCell>
            <TableCell>Method</TableCell>
            <TableCell>Origin</TableCell>
            <TableCell>Path</TableCell>
            <TableCell>Status</TableCell>
            {actionsCell && <TableCell padding="checkbox"></TableCell>}
          </TableRow>
        </TableHead>
        <TableBody>
          {sorted.map(({ id, method, url, response }) => {
            const { origin, path } = parseURLForDisplay(url);

            return (
              <RequestTableRow
                key={id}
                hover
                selected={id === activeRowId}
                onClick={() => {
                  onRowClick && onRowClick(id);
                }}
                onContextMenu={(e) => {
                  onContextMenu && onContextMenu(e, id);
                }}
              >
                <SeqTableCell>{seqById.get(id)}</SeqTableCell>
                <MethodTableCell>
                  <code>{method}</code>
                </MethodTableCell>
                <OriginTableCell>{origin}</OriginTableCell>
                <PathTableCell>{path}</PathTableCell>
                <StatusTableCell>
                  {response && <Status code={response.statusCode} reason={response.statusReason} />}
                </StatusTableCell>
                {actionsCell && actionsCell(id)}
              </RequestTableRow>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

function Status({ code, reason }: { code: number; reason: string }): JSX.Element {
  return (
    <div>
      <HttpStatusIcon status={code} />{" "}
      <code>
        {code} {reason}
      </code>
    </div>
  );
}
