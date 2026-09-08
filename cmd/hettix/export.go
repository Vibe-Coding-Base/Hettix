package main

import (
	"net/http"

	"github.com/Vibe-Coding-Base/Hettix/pkg/export"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
)

// exportHandler streams the active project's request logs as HAR or CSV for
// download.
func exportHandler(reqLogSvc *reqlog.Service, format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := reqLogSvc.FindByQuery(r.Context(), nil, 0)
		if err != nil {
			http.Error(w, "no active project or failed to load traffic", http.StatusConflict)
			return
		}

		switch format {
		case "har":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", `attachment; filename="hettix-export.har"`)

			if err := export.WriteHAR(w, version, logs); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", `attachment; filename="hettix-export.csv"`)

			if err := export.WriteCSV(w, logs); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "unknown export format", http.StatusBadRequest)
		}
	}
}
