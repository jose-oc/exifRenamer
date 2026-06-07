package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jose-oc/imagesExifRenamer/pkg/config"
	"github.com/jose-oc/imagesExifRenamer/pkg/renamer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Version is the application version, set at compile time using ldflags.
	Version = "dev"
	// RootCmd exposes the root Cobra command for use by main or tests.
	RootCmd = &cobra.Command{
		Use:     "exifrenamer [source_paths...]",
		Version: Version,
		Short:   "ExifRenamer is a CLI tool to rename and organize images/videos based on EXIF metadata.",
		Long: `A high-performance Go CLI utility to rename and organize image and video files
into a structured directory hierarchy using their EXIF/metadata creation dates.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration, resolving file configurations, environment variables and overrides.
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Set color disabling option globally
			color.NoColor = cfg.NoColor

			// If quiet mode is enabled, silence output unless errors occur.
			if cfg.Quiet {
				cfg.Verbose = false
			}

			if cfg.Verbose {
				fmt.Printf("Configured destination: %s\n", cfg.Destination)
				fmt.Printf("Recursive search: %t\n", cfg.Recursive)
				fmt.Printf("Dry run mode: %t\n", cfg.DryRun)
				fmt.Printf("Force lowercase extensions: %t\n", cfg.LowercaseExtension)
				fmt.Printf("Folder pattern: %s\n", cfg.Naming.FolderPattern)
				fmt.Printf("File pattern: %s\n", cfg.Naming.FilePattern)
				if cfg.Naming.DefaultPrefix != "" {
					fmt.Printf("Prefix: %s\n", cfg.Naming.DefaultPrefix)
				}
				if cfg.Naming.DefaultSuffix != "" {
					fmt.Printf("Suffix: %s\n", cfg.Naming.DefaultSuffix)
				}
			}

			if len(args) == 0 {
				if !cfg.Quiet {
					fmt.Println("No source paths specified. Use --help for usage details.")
				}
				return nil
			}

			if cfg.Verbose {
				fmt.Printf("Processing source paths: %v\n", args)
			}

			// Configure and instantiate the renamer
			opts := renamer.Options{
				Destination:        cfg.Destination,
				Recursive:          cfg.Recursive,
				DryRun:             cfg.DryRun,
				LowercaseExtension: cfg.LowercaseExtension,
				Verbose:            cfg.Verbose,
				FolderPattern:      cfg.Naming.FolderPattern,
				FilePattern:        cfg.Naming.FilePattern,
				DefaultPrefix:      cfg.Naming.DefaultPrefix,
				DefaultSuffix:      cfg.Naming.DefaultSuffix,
			}
			r := renamer.New(opts)
			if err := r.Process(args); err != nil {
				return err
			}

			if !cfg.Quiet {
				if cfg.DryRun {
					fmt.Println("Dry-run execution completed (no changes made).")
				}
			}

			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// CLI Flags registration
	RootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to a specific configuration file (overrides default search paths)")
	RootCmd.Flags().BoolP("recursive", "r", true, "Recursively traverse source directories (overrides config)")
	RootCmd.Flags().StringP("dest", "d", "", "Base destination directory (overrides config)")
	RootCmd.Flags().Bool("dry-run", false, "Dry-run mode. Displays what actions would be taken without making changes")
	RootCmd.Flags().BoolP("lower-ext", "l", true, "Force lowercase file extensions, e.g. .JPG -> .jpg (overrides config)")
	RootCmd.Flags().StringP("prefix", "p", "", "Temporary prefix to insert via %p or %P token (overrides config default)")
	RootCmd.Flags().StringP("suffix", "s", "", "Temporary suffix to insert via %s or %S token (overrides config default)")
	RootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	RootCmd.Flags().BoolP("quiet", "q", false, "Silence output except for critical errors")
	RootCmd.Flags().Bool("no-color", false, "Disable colorized console output")

	// Bind flags to Viper configuration values
	viper.BindPFlag("recursive", RootCmd.Flags().Lookup("recursive"))
	viper.BindPFlag("destination", RootCmd.Flags().Lookup("dest"))
	viper.BindPFlag("dry_run", RootCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("lowercase_extension", RootCmd.Flags().Lookup("lower-ext"))
	viper.BindPFlag("naming.default_prefix", RootCmd.Flags().Lookup("prefix"))
	viper.BindPFlag("naming.default_suffix", RootCmd.Flags().Lookup("suffix"))
	viper.BindPFlag("verbose", RootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("quiet", RootCmd.Flags().Lookup("quiet"))
	viper.BindPFlag("no_color", RootCmd.Flags().Lookup("no-color"))
}
