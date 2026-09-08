import { Box, MenuItem, TextField, Typography } from "@mui/material";

import { FONT_FAMILIES, useAppearance } from "lib/AppearanceContext";

export default function AppearanceSettings(): JSX.Element {
  const { fontFamily, fontSize, setAppearance } = useAppearance();

  return (
    <Box sx={{ maxWidth: 640 }}>
      <Typography variant="h5" sx={{ mb: 1 }}>
        Appearance
      </Typography>
      <Typography paragraph color="text.secondary">
        Font used by the request and response editors. Saved in this browser.
      </Typography>

      <Box sx={{ display: "flex", gap: 2 }}>
        <TextField
          select
          size="small"
          label="Editor font"
          value={fontFamily}
          onChange={(e) => setAppearance({ fontFamily: e.target.value })}
          sx={{ flex: 1 }}
        >
          {FONT_FAMILIES.map((f) => (
            <MenuItem key={f.value} value={f.value} sx={{ fontFamily: f.value }}>
              {f.label}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          type="number"
          size="small"
          label="Font size"
          value={fontSize}
          onChange={(e) => {
            const size = parseInt(e.target.value, 10);
            if (!Number.isNaN(size) && size >= 8 && size <= 32) {
              setAppearance({ fontSize: size });
            }
          }}
          inputProps={{ min: 8, max: 32 }}
          sx={{ width: 120 }}
        />
      </Box>

      <Box
        sx={{
          mt: 2,
          p: 1.5,
          borderRadius: 1,
          bgcolor: "action.hover",
          fontFamily,
          fontSize,
          whiteSpace: "pre",
          overflowX: "auto",
        }}
      >
        {`GET /api/users?id=1 HTTP/1.1\nHost: target.example\nAuthorization: Bearer eyJhbGci…`}
      </Box>
    </Box>
  );
}
