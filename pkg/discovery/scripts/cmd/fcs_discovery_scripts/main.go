package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/discovery/scripts"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger  = logrus.StandardLogger()
	rootCmd = &cobra.Command{
		Use:   "fcs_discovery_scripts",
		Short: "Functional Conformance Suite Server -  Discovery Scripts",
		Long: `Generates all the non-manadatory fields for each endpoint, e.g.,

$ go run pkg/discovery/scripts/cmd/fcs_discovery_scripts/main.go --swagger_path 'pkg/schema/spec/v3.1.2/account-info-swagger.flattened.json' --output_file 'pkg/discovery/scripts/generated/v3.1.2_account-info-discovery.json'
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logger.WithFields(logrus.Fields{
				"app": "fcs_discovery_scripts",
			})

			swaggerPath := viper.GetString("swagger_path")
			outputFile := viper.GetString("output_file")
			logger.WithFields(logrus.Fields{
				"swagger_path": swaggerPath,
				"output_file":  outputFile,
			}).Infof("Parsing started ...")

			nonManadatoryFields, err := scripts.ParseSchema(swaggerPath, logger.Logger)
			if err != nil {
				return err
			}

			// logger.WithFields(logrus.Fields{
			// 	"nonManadatoryFields": fmt.Sprintf("%#v", nonManadatoryFields),
			// }).Infof("Parsing finished ...")
			data, err := json.MarshalIndent(nonManadatoryFields, "", "  ")
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(outputFile, data, 0644)
			if err != nil {
				return err
			}

			logger.WithFields(logrus.Fields{
				"output_file": outputFile,
			}).Infof("Finished writing to file ...")

			return nil
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func init() {
	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.String("log_level", "INFO", "Log level")
	persistentFlags.StringP("swagger_path", "s", "", "Swagger file path, e.g., 'pkg/schema/spec/v3.1.2/account-info-swagger.flattened.json'")
	persistentFlags.StringP("output_file", "o", "", "Swagger file path, e.g., 'pkg/discovery/scripts/generated/v3.1.2_account-info-discovery.json'")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	cobra.MarkFlagRequired(persistentFlags, "swagger_path")
	cobra.MarkFlagRequired(persistentFlags, "output_file")
	cobra.OnInitialize(onInitialize)
}

func onInitialize() {
	logger.SetNoLock()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    false,
		ForceColors:      true,
		FullTimestamp:    false,
		DisableTimestamp: true,
	})
	level, err := logrus.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		printCommandFlags()
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	logger.SetLevel(level)

	printCommandFlags()
}

func printCommandFlags() {
	rootCmd.PersistentFlags().PrintDefaults()

	logger.WithFields(logrus.Fields{
		"swagger_path": viper.GetString("swagger_path"),
		"output_file":  viper.GetString("output_file"),
		"log_level":    viper.GetString("log_level"),
	}).Info("configuration flags")
}
