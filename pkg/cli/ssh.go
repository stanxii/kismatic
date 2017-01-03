package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/spf13/cobra"
)

type sshOpts struct {
	planFilename string
	host         string
	arguments    []string
}

// NewCmdSSH returns an ssh shell
func NewCmdSSH(out io.Writer) *cobra.Command {
	opts := &sshOpts{}

	cmd := &cobra.Command{
		Use:   "ssh HOST [commands]",
		Short: "ssh into a node in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}
			// get optional arguments
			if len(args) > 1 {
				opts.arguments = args[1:]
			}

			opts.host = args[0]

			planner := &install.FilePlanner{File: opts.planFilename}

			err := doSSH(out, planner, opts)
			// 130 = terminated by Control-C, so not an actual error
			if err != nil && !strings.Contains(err.Error(), "130") {
				return fmt.Errorf("Error trying to connect to host %q: %v", opts.host, err)
			}
			return nil
		},
	}

	// PersistentFlags
	cmd.PersistentFlags().StringVarP(&opts.planFilename, "plan-file", "f", "kismatic-cluster.yaml", "path to the installation plan file")

	return cmd
}

func doSSH(out io.Writer, planner install.Planner, opts *sshOpts) error {
	// Check if plan file exists
	if !planner.PlanExists() {
		return fmt.Errorf("plan does not exist")
	}
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}

	// find node
	con, err := plan.GetSSHConnection(opts.host)
	if err != nil {
		return err
	}

	// validate node is able to SSH
	ok, errs := install.ValidateSSHConnection(con, "")
	if !ok {
		printValidationErrors(out, errs)
		return fmt.Errorf("cannot validate SSH connection to node %q", opts.host)
	}

	client, err := createSSHClient(con)
	if err != nil {
		return fmt.Errorf("error creating SSH client: %v", err)
	}

	return client.Shell(strings.Join(opts.arguments, " "))
}

func createSSHClient(con *install.SSHConnection) (ssh.Client, error) {
	addr := con.GetSSHAddress()
	port := con.GetSSHPort()
	auth := &ssh.Auth{}
	if con.GetSSHKeyPath() != "" {
		auth.Keys = []string{con.GetSSHKeyPath()}
	}

	return ssh.NewClient(con.GetSSHUsername(), addr, port, auth)
}
