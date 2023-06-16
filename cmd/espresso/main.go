package main

import (
	"github.com/gregorychen3/espresso-controller/cmd/espresso/cmdutil"
	"github.com/gregorychen3/espresso-controller/cmd/espresso/config"
	"github.com/gregorychen3/espresso-controller/cmd/espresso/log"
	"github.com/gregorychen3/espresso-controller/internal/espresso"
	serverLogger "github.com/gregorychen3/espresso-controller/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configKeys = []config.Key{
	{Path: "Port", ShortFlag: "p", Description: "Port on which the espresso server should listen", Default: "8080"},
	{Path: "RelayPin", ShortFlag: "r", Description: "The GPIO connected to the relay", Default: 21},
	{Path: "BoilerThermCsPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 chip select, aka chip enable", Default: 29},
	{Path: "BoilerThermClkPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 clock", Default: 23},
	{Path: "BoilerThermMisoPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 data output", Default: 21},
	/* {Path: "BoilerThermCsPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 chip select, aka chip enable", Default: 3},
	{Path: "BoilerThermClkPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 clock", Default: 4},
	{Path: "BoilerThermMisoPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31855 data output", Default: 2},*/
	{Path: "Verbose", ShortFlag: "v", Description: "verbose output", Default: false},
}

func newRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "espresso",
		Short: "Control and monitor an espresso machine",
		Long:  "Control and monitor an espresso machine",
		PreRun: func(cmd *cobra.Command, args []string) {
			// Bind config in PreRun() to avoid collisions with other commands' flags
			for _, k := range configKeys {
				if err := viper.BindPFlag(k.Path, cmd.Flags().Lookup(k.Flag())); err != nil {
					log.Fatal("Failed to bind flag to config: %+v", k)
				}
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info(cmdutil.Logo)
			log.Info("For more information, go to https://github.com/gregorychen3/espresso-controller\n")

			if verbose := viper.GetBool("Verbose"); verbose {
				serverLogger.UseDevLogger()
			} else {
				serverLogger.UseProdLogger(
					viper.GetString(config.KeyLogFilePath),
					viper.GetInt(config.KeyLogFileMaxSize),
					viper.GetInt(config.KeyLogFileMaxAge),
					viper.GetInt(config.KeyLogFileMaxBackups),
				)
			}

			c := espresso.Configuration{}
			if err := viper.Unmarshal(&c); err != nil {
				log.Fatal("Unmarshalling configuration: %s\n", err.Error())
			}

			server := espresso.New(c)
			return server.Run()
		},
	}

	for _, k := range configKeys {
		if k.Default != nil {
			viper.SetDefault(k.Path, k.Default)
		}
		k.BindFlag(&cmd)
	}
	for _, k := range configKeys {
		viper.BindEnv(k.Path, k.EnvKey())
	}
	return &cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
