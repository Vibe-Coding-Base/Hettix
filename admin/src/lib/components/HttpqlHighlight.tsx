import { Box, useTheme } from "@mui/material";

type TokenKind = "field" | "operator" | "keyword" | "string" | "number" | "text";

interface Token {
  value: string;
  kind: TokenKind;
}

const KEYWORDS = /^(and|or|not)$/i;
const FIELD = /^(req|resp|res)\.[a-z_]+(\[[^\]]*\])?$/i;
const OPERATOR = /^(eq|ne|cont|ncont|like|nlike|regex|nregex|gt|gte|lt|lte|=~|!~|!=|>=|<=|=|>|<)$/i;

// tokenize splits an HTTPQL query into classified tokens for display. It mirrors
// the server grammar closely enough to colorize; it is not a validating parser.
function tokenize(query: string): Token[] {
  const tokens: Token[] = [];
  const pattern = /(\s+)|("(?:[^"\\]|\\.)*"?)|(\(|\))|([^\s()]+)/g;

  let match: RegExpExecArray | null;
  while ((match = pattern.exec(query)) !== null) {
    const [, space, str, paren, word] = match;

    if (space) {
      tokens.push({ value: space, kind: "text" });
    } else if (str) {
      tokens.push({ value: str, kind: "string" });
    } else if (paren) {
      tokens.push({ value: paren, kind: "text" });
    } else if (word) {
      tokens.push({ value: word, kind: classify(word) });
    }
  }

  return tokens;
}

function classify(word: string): TokenKind {
  if (KEYWORDS.test(word)) return "keyword";
  if (FIELD.test(word)) return "field";
  if (OPERATOR.test(word)) return "operator";
  if (/^\d+$/.test(word)) return "number";
  return "text";
}

// HttpqlHighlight renders an HTTPQL query with syntax coloring.
export default function HttpqlHighlight({ query }: { query: string }): JSX.Element {
  const theme = useTheme();

  const colors: Record<TokenKind, string> = {
    field: theme.palette.primary.main,
    operator: theme.palette.info.main,
    keyword: theme.palette.primary.main,
    string: theme.palette.success.main,
    number: theme.palette.warning.main,
    text: theme.palette.text.primary,
  };

  return (
    <Box
      component="pre"
      sx={{
        m: 0,
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: "0.8rem",
        whiteSpace: "pre-wrap",
        wordBreak: "break-word",
      }}
    >
      {tokenize(query).map((token, i) => (
        <span
          key={i}
          style={{
            color: colors[token.kind],
            fontWeight: token.kind === "keyword" ? 700 : 400,
          }}
        >
          {token.value}
        </span>
      ))}
    </Box>
  );
}
