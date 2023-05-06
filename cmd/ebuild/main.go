package main

import (
	"context"
	"fmt"
	"os"

	"github.com/edgetx/cloudbuild/cmd/ebuild/auth"
	"github.com/edgetx/cloudbuild/cmd/ebuild/db"
	"github.com/edgetx/cloudbuild/cmd/ebuild/run"
	"github.com/edgetx/cloudbuild/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd *cobra.Command

type cobraRunFunc func(*cobra.Command, []string)

func unmarshalConfig(o *config.CloudbuildOpts) cobraRunFunc {
	return func(c *cobra.Command, _ []string) {
		_ = o.Viper.BindPFlags(c.Flags())
		err := o.Unmarshal()
		if err != nil {
			fmt.Println("Can't unmarshal config:", err)
			os.Exit(1)
		}
	}
}

func init() {
	v := viper.New()
	o := config.NewOpts(v)

	rootCmd = &cobra.Command{
		Use:              "ebuild",
		Short:            "ebuild is the EdgeTX cloud build server",
		PersistentPreRun: unmarshalConfig(o),
	}

	cobra.OnInitialize(config.InitConfig(v))

	ctx := context.Background()
	rootCmd.AddCommand(run.NewRunCommand(ctx, o))
	rootCmd.AddCommand(db.NewDBCommand(ctx, o))
	rootCmd.AddCommand(auth.NewAuthCommand(ctx, o))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
