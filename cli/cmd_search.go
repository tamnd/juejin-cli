package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search Juejin articles by keyword",
		Long:  "Search Juejin articles by keyword using the Juejin search API.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			n := a.effectiveLimit(10)
			a.progressf("fetching up to %d results for %q...", n, query)
			results, err := a.client.Search(cmd.Context(), query, "0", n)
			if err != nil {
				return mapFetchErr(err)
			}
			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "no results found")
				return codeError(exitNoData, nil)
			}
			return a.render(results)
		},
	}
}
