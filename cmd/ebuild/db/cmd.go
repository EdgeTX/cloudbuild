package db

import (
	"context"
	"fmt"
	"os"

	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	"github.com/spf13/cobra"
)

func NewDBCommand(ctx context.Context, o *config.CloudbuildOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "EdgeTX build server DB maintenance",
	}
	o.BindCliOpts(cmd)
	o.BindDBOpts(cmd)
	cmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Create / migrate DB schema",
		Run: func(cmd *cobra.Command, args []string) {
			if err := database.Migrate(o.DatabaseDSN); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Println("Database migrated successfully")
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "drop",
		Short: "Drop DB schema",
		Run: func(cmd *cobra.Command, args []string) {
			if err := database.DropSchema(o.DatabaseDSN); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Println("Database schema dropped successfully")
		},
	})
	return cmd
}
