package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// categoryIDs maps friendly names to Juejin category IDs.
var categoryIDs = map[string]string{
	"frontend": "6809637767543259144",
	"backend":  "6809637769959178254",
	"ai":       "6809637773935329294",
	"android":  "6809635626879549453",
	"ios":      "6809635632167337997",
	"devtools": "6809637776263217166",
	"career":   "6809637771511070734",
}

func (a *App) latestCmd() *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Latest Juejin articles",
		Long:  "Fetch the newest articles from Juejin. Use --category to filter by topic.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)

			var categoryID string
			if category != "" {
				id, ok := categoryIDs[category]
				if !ok {
					return codeError(exitUsage, fmt.Errorf("unknown category %q (valid: frontend, backend, ai, android, ios, devtools, career)", category))
				}
				categoryID = id
			}

			if category != "" {
				a.progressf("fetching %d latest %s articles...", n, category)
			} else {
				a.progressf("fetching %d latest articles...", n)
			}

			articles, err := a.client.Feed(cmd.Context(), 3, categoryID, "0", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by category: frontend|backend|ai|android|ios|devtools|career")
	return cmd
}
