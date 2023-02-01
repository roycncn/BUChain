package main

import (
	"fmt"
	"github.com/roycncn/BUChain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "chaind",
		Short: "HKBU COMP7200 Project Chain Server",
	}

	rootCmd.AddCommand(VersionCmd())
	rootCmd.AddCommand(StartCmd())
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("err:%v", err)
		os.Exit(1)
	}

}

func StartCmd() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the chaind server",
		Long:  `Run the chaind server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Starting chaind server")
			cfg, err := config.New(configFile)
			if err != nil {
				panic(err)
			}

			if cfg.LogLevel == "INFO" {
				log.SetLevel(log.InfoLevel)
			} else if cfg.LogLevel == "DEBUG" {
				log.SetLevel(log.DebugLevel)
			} else if cfg.LogLevel == "TRACE" {
				log.SetLevel(log.TraceLevel)
			} else {
				log.SetLevel(log.InfoLevel)
			}

			BUChain.Start(cfg)

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&configFile, "config", "f", "./config.yaml", "config file (default is ../config.yaml)")

	return cmd
}

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get chaind  server version",
		Long:  `Get chaind  server version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("chaind v0.1")
			return nil
		},
	}
	return cmd
}
