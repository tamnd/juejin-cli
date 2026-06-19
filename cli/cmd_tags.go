package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) tagsCmd() *cobra.Command {
	var cursor string
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List Juejin tags",
		Long:  `Fetch the tag list from Juejin. Use --cursor to paginate.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d tags (cursor=%s)...", n, cursor)
			tags, err := a.client.Tags(cmd.Context(), cursor, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(tags, len(tags))
		},
	}
	cmd.Flags().StringVar(&cursor, "cursor", "0", "pagination cursor (0 for first page)")
	return cmd
}
