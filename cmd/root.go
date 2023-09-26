/*
Copyright © 2023 Ted van Riel

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/mitaka8/playlist-bot/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "playlist-bot",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
}
var discordbotCmd = &cobra.Command{
	Use: "discord",
	Run: func(cmd *cobra.Command, args []string) {
		app.DiscordBot()
	},
}
var webCmd = &cobra.Command{
	Use: "web",
	Run: func(cmd *cobra.Command, args []string) {
		app.Web()
	},
}

var saveCmd = &cobra.Command{
	Use: "save <-p playlist> <-g guildId> <-u url>",

	Run: func(cmd *cobra.Command, args []string) {
		url, uerr := cmd.Flags().GetString("url")
		playlist, perr := cmd.Flags().GetString("playlist")
		guild, gerr := cmd.Flags().GetString("guild")

		if multierr.Combine(uerr, gerr, perr) != nil {
			os.Exit(1)
		}
		app.Save(url, guild, playlist)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.playlist-bot.yaml)")

	saveCmd.Flags().StringP("playlist", "p", "", "Playlist name to insert in")
	saveCmd.Flags().StringP("guild", "g", "", "Guild to download for")
	saveCmd.Flags().StringP("url", "u", "", "URL to download")
	saveCmd.MarkFlagRequired("playlist")
	saveCmd.MarkFlagRequired("guild")
	saveCmd.MarkFlagRequired("url")

	rootCmd.AddCommand(discordbotCmd, webCmd, saveCmd)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".playlist-bot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".playlist-bot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
