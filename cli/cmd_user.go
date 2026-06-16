package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) userCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "user USER_ID",
		Short: "Fetch a Juejin user profile",
		Long:  `Look up a Juejin user by their numeric user ID and print profile fields.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]
			a.progressf("fetching user %s...", userID)
			profile, err := a.client.User(cmd.Context(), userID)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty([]any{profile}, 1)
		},
	}
}
