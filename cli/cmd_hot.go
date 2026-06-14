package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) hotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hot",
		Short: "Hot and recommended Juejin articles",
		Long:  "Fetch hot/recommended articles from Juejin using the recommendation engine (sort_type=200).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d hot articles...", n)
			articles, err := a.client.Feed(cmd.Context(), 200, "", "0", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}
}
