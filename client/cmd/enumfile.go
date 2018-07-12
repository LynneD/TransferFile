// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/LynneD/TransferFile/client_routine"
)

var regexp string

// clientCmd represents the client command
var enumfileCmd = &cobra.Command{
	Use:   "enumfile",
	Short: "Enumerate Files stored on server",
	Long: `The client will connect to the server(host:port) and transfer file by truncating it to several chunks whose 
           size is defined by the parameter`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(port) == 0 {
			cmd.Help()
			return
		}
		clientroutine.EnumFile(host, port, regexp)
	},
}

func init() {
	rootCmd.AddCommand(enumfileCmd)

	enumfileCmd.Flags().StringVarP(&host, "host", "H", "", "The host you want to connect")
	enumfileCmd.Flags().StringVarP(&port, "port", "p", "", "The port you use to talk to the host")
	enumfileCmd.Flags().StringVarP(&regexp, "regular expression", "r", "", "The file you want to store remotely")

}
