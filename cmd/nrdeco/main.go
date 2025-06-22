package main

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	"github.com/miyamo2/nrdeco/internal"
	"github.com/spf13/cobra"
)

var (
	// version of the nrdeco
	version string
	// revision of the nrdeco
	revision string
)

func main() {
	comand, err := rootCmd()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	if err := comand.Execute(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

func rootCmd() (*cobra.Command, error) {
	var (
		sourceFlag  string
		destFlag    string
		versionFlag bool
	)
	command := &cobra.Command{
		Use:   "nrdeco",
		Short: "nrdeco generates decorated implementations with New Relic segments from interfaces.",
		Long:  `nrdeco generates decorated implementations with New Relic segments from interfaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				cmd.Printf("[nrdeco] version %s-%s\n", version, revision)
				return nil
			}
			cmd.Printf("[nrdeco] input: %s", sourceFlag)
			dest := cmp.Or(destFlag, strings.Replace(sourceFlag, ".go", ".nrdeco.go", -1))

			b, err := internal.Generate(cmd.Context(), sourceFlag, dest, version)
			if err != nil {
				return fmt.Errorf("[nrdeco] failed to generate code from %s: %w", sourceFlag, err)
			}

			f, err := os.Create(dest)
			defer func() {
				_ = f.Close()
			}()
			if err != nil {
				return fmt.Errorf("[nrdeco] failed to create %s: %w", dest, err)
			}
			_, err = f.Write(b)
			if err != nil {
				return fmt.Errorf("[nrdeco] failed to write to %s: %w", dest, err)
			}
			cmd.Printf("[nrdeco] wrote: %s\n", dest)
			return nil
		},
	}
	command.Flags().BoolVar(&versionFlag, "version", false, `Print the version of nrdeco.`)
	command.Flags().
		StringVarP(&sourceFlag, "source", "s", "", `A file containing interfaces to be decorate.`)
	command.Flags().
		StringVarP(&destFlag, "dest", "d", "", `A file to which the resulting source code will be written. If not provided, the code will be written to <source>.nrdeco.go instead.`)
	err := command.MarkFlagFilename("source", "go")
	if err != nil {
		return nil, err
	}
	err = command.MarkFlagFilename("dest", "go")
	if err != nil {
		return nil, err
	}
	command.MarkFlagsOneRequired("source", "version")
	command.MarkFlagsMutuallyExclusive("source", "version")
	return command, nil
}
