package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) pinsCmd() *cobra.Command {
	var cursor string
	cmd := &cobra.Command{
		Use:   "pins",
		Short: "List Juejin pins (沸点 moments)",
		Long:  `Fetch public pins/moments (沸点) from Juejin. Use --cursor to paginate.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d pins (cursor=%s)...", n, cursor)
			pins, err := a.client.Pins(cmd.Context(), cursor, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(pins, len(pins))
		},
	}
	cmd.Flags().StringVar(&cursor, "cursor", "0", "pagination cursor (0 for first page)")
	return cmd
}
