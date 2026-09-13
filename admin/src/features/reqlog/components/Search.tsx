import FilterListIcon from "@mui/icons-material/FilterList";
import SearchIcon from "@mui/icons-material/Search";
import {
  Alert,
  Box,
  Checkbox,
  CircularProgress,
  ClickAwayListener,
  Divider,
  FormControlLabel,
  FormGroup,
  InputBase,
  Link,
  Paper,
  Popper,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
  Typography,
  useTheme,
} from "@mui/material";
import IconButton from "@mui/material/IconButton";
import React, { useRef, useState } from "react";

import HttpqlHighlight from "lib/components/HttpqlHighlight";
import { openExternal } from "lib/desktop";
import {
  HttpRequestLogFilterDocument,
  useHttpRequestLogFilterQuery,
  useSetHttpRequestLogFilterMutation,
} from "lib/graphql/generated";
import { withoutTypename } from "lib/graphql/omitTypename";
import { emptyLogFilters, hasActiveLogFilters, LogFilters, STATUS_CLASSES, TYPE_CATEGORIES } from "lib/logFilters";

interface SearchProps {
  filters: LogFilters;
  onFiltersChange: (filters: LogFilters) => void;
}

function Search({ filters, onFiltersChange }: SearchProps): JSX.Element {
  const theme = useTheme();

  const toggleType = (id: string) => {
    onFiltersChange({
      ...filters,
      types: filters.types.includes(id) ? filters.types.filter((t) => t !== id) : [...filters.types, id],
    });
  };
  const filtersActive = hasActiveLogFilters(filters);

  const [searchExpr, setSearchExpr] = useState("");
  const filterResult = useHttpRequestLogFilterQuery({
    onCompleted: (data) => {
      setSearchExpr(data.httpRequestLogFilter?.searchExpression || "");
    },
  });
  const filter = filterResult.data?.httpRequestLogFilter;

  const [setFilterMutate, setFilterResult] = useSetHttpRequestLogFilterMutation({
    update(cache, { data }) {
      cache.writeQuery({
        query: HttpRequestLogFilterDocument,
        data: {
          httpRequestLogFilter: data?.setHttpRequestLogFilter,
        },
      });
    },
  });

  const filterRef = useRef<HTMLFormElement>(null);
  const [filterOpen, setFilterOpen] = useState(false);

  const handleSubmit = (e: React.SyntheticEvent) => {
    setFilterMutate({
      variables: {
        filter: {
          ...withoutTypename(filter),
          searchExpression: searchExpr,
        },
      },
    });
    setFilterOpen(false);
    e.preventDefault();
  };

  const handleClickAway = (event: MouseEvent | TouchEvent) => {
    if (filterRef?.current && filterRef.current.contains(event.target as HTMLElement)) {
      return;
    }
    setFilterOpen(false);
  };

  return (
    <Box>
      <Error prefix="Error fetching filter" error={filterResult.error} />
      <Error prefix="Error setting filter" error={setFilterResult.error} />
      <Box style={{ display: "flex", flex: 1 }}>
        <ClickAwayListener onClickAway={handleClickAway}>
          <Paper
            component="form"
            autoComplete="off"
            onSubmit={handleSubmit}
            ref={filterRef}
            sx={{
              padding: "2px 4px",
              display: "flex",
              alignItems: "center",
              width: "100%",
            }}
          >
            <Tooltip title="Toggle filter options">
              <IconButton
                onClick={() => setFilterOpen(!filterOpen)}
                sx={{
                  p: 1,
                  color: filter?.onlyInScope || filtersActive ? "primary.main" : "inherit",
                }}
              >
                {filterResult.loading || setFilterResult.loading ? (
                  <CircularProgress sx={{ color: theme.palette.text.primary }} size={23} />
                ) : (
                  <FilterListIcon />
                )}
              </IconButton>
            </Tooltip>
            <InputBase
              sx={{
                ml: 1,
                flex: 1,
              }}
              placeholder='Search, e.g. req.method eq "POST" and resp.code gte 400'
              value={searchExpr}
              onChange={(e) => setSearchExpr(e.target.value)}
              onFocus={() => setFilterOpen(true)}
              autoCorrect="false"
              spellCheck="false"
            />
            <Tooltip title="Search">
              <IconButton type="submit" sx={{ padding: 1.25 }}>
                <SearchIcon />
              </IconButton>
            </Tooltip>
            <Popper
              open={filterOpen}
              anchorEl={filterRef.current}
              placement="bottom"
              style={{ zIndex: theme.zIndex.appBar }}
            >
              <Paper
                sx={{
                  width: 400,
                  marginTop: 0.5,
                  p: 1.5,
                }}
              >
                {searchExpr.trim() !== "" && (
                  <Box sx={{ mb: 1, p: 1, borderRadius: 1, bgcolor: "action.hover", overflowX: "auto" }}>
                    <HttpqlHighlight query={searchExpr} />
                  </Box>
                )}
                <Typography variant="caption" color="text.secondary" component="div" sx={{ mb: 1 }}>
                  HTTPQL — fields <code>req.method</code>, <code>req.host</code>, <code>req.path</code>,{" "}
                  <code>req.body</code>, <code>req.header[&quot;Name&quot;]</code>, <code>resp.code</code>,{" "}
                  <code>resp.body</code>; operators <code>eq ne cont regex gt gte lt lte</code>; combine with{" "}
                  <code>and</code> <code>or</code> <code>not</code>.{" "}
                  <Link
                    component="button"
                    type="button"
                    variant="caption"
                    onClick={() => openExternal("https://github.com/Vibe-Coding-Base/Hettix/blob/main/docs/HTTPQL.md")}
                  >
                    Full reference
                  </Link>
                </Typography>
                <Typography variant="overline" color="text.secondary" component="div">
                  Request type
                </Typography>
                <FormGroup>
                  <FormControlLabel
                    control={
                      <Checkbox
                        size="small"
                        checked={filter?.onlyInScope ? true : false}
                        disabled={filterResult.loading || setFilterResult.loading}
                        onChange={(e) =>
                          setFilterMutate({
                            variables: {
                              filter: {
                                ...withoutTypename(filter),
                                onlyInScope: e.target.checked,
                              },
                            },
                          })
                        }
                      />
                    }
                    label={<Typography variant="body2">Only in-scope requests</Typography>}
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        size="small"
                        checked={filters.hideNoResponse}
                        onChange={(e) => onFiltersChange({ ...filters, hideNoResponse: e.target.checked })}
                      />
                    }
                    label={<Typography variant="body2">Hide items without responses</Typography>}
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        size="small"
                        checked={filters.onlyParameterized}
                        onChange={(e) => onFiltersChange({ ...filters, onlyParameterized: e.target.checked })}
                      />
                    }
                    label={<Typography variant="body2">Only parameterized requests</Typography>}
                  />
                </FormGroup>

                <Divider sx={{ my: 1 }} />

                <Typography variant="overline" color="text.secondary" component="div">
                  Status
                </Typography>
                <ToggleButtonGroup
                  size="small"
                  value={filters.statuses}
                  onChange={(_, value: string[]) => onFiltersChange({ ...filters, statuses: value })}
                  sx={{ mb: 1 }}
                >
                  {STATUS_CLASSES.map((s) => (
                    <ToggleButton key={s} value={s} sx={{ px: 1.5, py: 0.2 }}>
                      {s}
                    </ToggleButton>
                  ))}
                </ToggleButtonGroup>

                <Typography variant="overline" color="text.secondary" component="div">
                  File types
                </Typography>
                <FormGroup row>
                  {TYPE_CATEGORIES.map((t) => (
                    <FormControlLabel
                      key={t.id}
                      sx={{ width: "48%", m: 0 }}
                      control={
                        <Checkbox
                          size="small"
                          checked={filters.types.includes(t.id)}
                          onChange={() => toggleType(t.id)}
                        />
                      }
                      label={<Typography variant="body2">{t.label}</Typography>}
                    />
                  ))}
                </FormGroup>

                <Typography variant="overline" color="text.secondary" component="div" sx={{ mt: 1 }}>
                  File extension
                </Typography>
                <Stack direction="row" spacing={1}>
                  <TextField
                    size="small"
                    fullWidth
                    variant="outlined"
                    label="Show only"
                    placeholder="php, asp, jsp"
                    value={filters.showExtensions}
                    onChange={(e) => onFiltersChange({ ...filters, showExtensions: e.target.value })}
                    inputProps={{ spellCheck: false, autoCapitalize: "off" }}
                  />
                  <TextField
                    size="small"
                    fullWidth
                    variant="outlined"
                    label="Hide"
                    placeholder="css, png, js"
                    value={filters.hideExtensions}
                    onChange={(e) => onFiltersChange({ ...filters, hideExtensions: e.target.value })}
                    inputProps={{ spellCheck: false, autoCapitalize: "off" }}
                  />
                </Stack>

                {filtersActive && (
                  <Link
                    component="button"
                    type="button"
                    variant="caption"
                    onClick={() => onFiltersChange(emptyLogFilters)}
                    sx={{ mt: 1, display: "inline-block" }}
                  >
                    Clear filters
                  </Link>
                )}
              </Paper>
            </Popper>
          </Paper>
        </ClickAwayListener>
      </Box>
    </Box>
  );
}

function Error(props: { prefix: string; error?: Error }) {
  if (!props.error) return null;

  return (
    <Box mb={4}>
      <Alert severity="error">
        {props.prefix}: {props.error.message}
      </Alert>
    </Box>
  );
}

export default Search;
