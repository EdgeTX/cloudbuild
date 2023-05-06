package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/edgetx/cloudbuild/auth"
	"github.com/edgetx/cloudbuild/config"
	"github.com/spf13/cobra"
)

func newTokenStore(o *config.CloudbuildOpts) *auth.AuthTokenDB {
	tokenStore, err := auth.NewAuthTokenDBFromConfig(o)
	if err != nil {
		fmt.Println("failed to create token store:", err)
		os.Exit(1)
	}
	return tokenStore
}

func NewAuthCommand(ctx context.Context, o *config.CloudbuildOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Token authentication related commands",
	}

	o.BindCliOpts(cmd)
	o.BindDBOpts(cmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create authentication token",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			token, err := newTokenStore(o).CreateToken(args[0], nil)
			if err != nil {
				fmt.Println("failed to create new token:", err)
				os.Exit(1)
			}
			fmt.Println("AccessKey:", token.AccessKey)
			fmt.Println("SecretKey:", token.SecretKey)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "authenticate",
		Short: "Authenticate provided token",
		Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			err := newTokenStore(o).Authenticate(args[0], args[1])
			if err != nil {
				if errors.Is(err, auth.ErrAuthenticationFailed) {
					fmt.Println("authentication failed")
				} else {
					fmt.Println("authentication failed:", err)
				}
				os.Exit(1)
			} else {
				fmt.Println("authentication successful")
			}
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List existing token",
		Run: func(cmd *cobra.Command, args []string) {
			tokens, err := newTokenStore(o).ListTokens()
			if err != nil {
				fmt.Println("failed to list tokens:", err)
				os.Exit(1)
			} else {
				for _, token := range *tokens {
					fmt.Println(
						token.AccessKey,
						token.SecretKey,
						token.ValidUntil,
					)
				}
			}
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "remove",
		Short: "Remove a token",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			accessKey := args[0]
			err := newTokenStore(o).RemoveToken(accessKey)
			if err != nil {
				fmt.Println("failed to remove token:", err)
				os.Exit(1)
			} else {
				fmt.Println("token", accessKey, "removed")
			}
		},
	})

	return cmd
}
