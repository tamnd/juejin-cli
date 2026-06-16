package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) hotCmd() *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:   "hot",
		Short: "Hot and recommended Juejin articles",
		Long:  "Fetch hot/recommended articles from Juejin using the recommendation engine (sort_type=200).",
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
				a.progressf("fetching %d hot %s articles...", n, category)
			} else {
				a.progressf("fetching %d hot articles...", n)
			}

			articles, err := a.client.Feed(cmd.Context(), 200, categoryID, "0", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by category: frontend|backend|ai|android|ios|devtools|career")
	return cmd
}
