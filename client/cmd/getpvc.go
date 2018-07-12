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


// clientCmd represents the client command
var getpvcCmd = &cobra.Command{
	Use:   "getpvc",
	Short: "Get the pvc list of files stored on server",
	Long: `The client will connect to the server(host:port) and transfer file by truncating it to several chunks whose 
           size is defined by the parameter`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("=====================")
		if len(port) == 0 {
			cmd.Help()
			return
		}
		files = append(files, args...)
		//fmt.Println(files)
		clientroutine.GetPVC(host, port, files)
	},
}

func init() {
	rootCmd.AddCommand(getpvcCmd)

	getpvcCmd.Flags().StringVarP(&host, "host", "H", "", "The host you want to connect")
	getpvcCmd.Flags().StringVarP(&port, "port", "p", "", "The port you use to talk to the host")
	getpvcCmd.Flags().StringSliceVarP(&files,"file list", "f", nil, "The list of files you want to send")
}
