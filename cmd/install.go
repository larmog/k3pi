/*
Copyright © 2019 The Nature of Software Nordic AB <lars@thenatureofsoftware.se>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg"
	cmd2 "github.com/TheNatureOfSoftware/k3pi/pkg/cmd"
	"github.com/TheNatureOfSoftware/k3pi/pkg/misc"
	"github.com/kubernetes-sigs/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs k3os on selected nodes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fn := viper.GetString("filename")

		var bytes []byte
		var err error

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			//fmt.Println("data is being piped to stdin")
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			if fn == "" {
				misc.ErrorExitWithMessage("must specify --filename|-f")
			}
			bytes, err = ioutil.ReadFile(fn)
		}
		misc.PanicOnError(err, "error reading input file")

		nodes := []*pkg.Node{}
		err = yaml.Unmarshal(bytes, &nodes)
		misc.ExitOnError(err, "error parsing nodes from file")

		if len(nodes) == 0 {
			misc.ErrorExitWithMessage("No nodes found in file")
		}

		sshKeys := viper.GetStringSlice("install-ssh-key")
		server := viper.GetString("server")
		token := viper.GetString("token")
		dryRun := viper.GetBool("dry-run")

		if len(sshKeys) == 0 {

			misc.ErrorExitWithMessage("at least one ssh key is required")

		} else if len(sshKeys) == 1 && sshKeys[0] == cmd2.DefaultSSHAuthorizedKey {

			idRsaPubFile, err := homedir.Expand(cmd2.DefaultSSHAuthorizedKey)
			msg := fmt.Sprintf("failed to read default ssh public key: %s", cmd2.DefaultSSHAuthorizedKey)
			misc.ExitOnError(err, msg)

			f, err := os.Open(idRsaPubFile)
			defer f.Close()
			misc.ExitOnError(err, msg)

			b, err := ioutil.ReadAll(f)
			misc.ExitOnError(err, msg)

			key := strings.Split(strings.TrimSpace(string(b)), " ")
			sshKeys = []string{fmt.Sprintf("%s %s", key[0], key[1])}
		}

		err = cmd2.Install(nodes, sshKeys, server, token, dryRun)
		misc.ExitOnError(err)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("dry-run", false, "if true will print the install commands but never run them")
	installCmd.Flags().StringP("filename", "f", "", "YAML file with all nodes")
	installCmd.Flags().Lookup("filename").NoOptDefVal = ""
	installCmd.Flags().StringP("server", "s", "", "ip address or hostname of the server node")
	installCmd.Flags().StringP("token", "t", "", "token or cluster secret for joining a server")

	installCmd.Flags().StringSliceP("ssh-key", "k", []string{cmd2.DefaultSSHAuthorizedKey}, "ssh authorized key that should be added to the rancher user")
	_ = viper.BindPFlag("dry-run", installCmd.Flags().Lookup("dry-run"))
	_ = viper.BindPFlag("filename", installCmd.Flags().Lookup("filename"))
	_ = viper.BindPFlag("server", installCmd.Flags().Lookup("server"))
	_ = viper.BindPFlag("install-ssh-key", installCmd.Flags().Lookup("ssh-key"))
	_ = viper.BindPFlag("token", installCmd.Flags().Lookup("token"))
}
