package cli

import (
	"fmt"
	"io"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

// NewCmdPlan creates a new install plan command
func NewCmdPlan(in io.Reader, out io.Writer, options *installOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "plan your Kubernetes cluster and generate a plan file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("Unexpected args: %v", args)
			}
			planner := &install.FilePlanner{File: options.planFilename}
			return doPlan(in, out, planner, options.planFilename)
		},
	}

	return cmd
}

func doPlan(in io.Reader, out io.Writer, planner install.Planner, planFile string) error {
	fmt.Fprintln(out, "Plan your Kubernetes cluster:")

	etcdNodes, err := util.PromptForInt(in, out, "Number of etcd nodes", 3)
	if err != nil {
		return fmt.Errorf("Error reading number of etcd nodes: %v", err)
	}
	if etcdNodes <= 0 {
		return fmt.Errorf("The number of etcd nodes must be greater than zero")
	}
	masterNodes, err := util.PromptForInt(in, out, "Number of master nodes", 2)
	if err != nil {
		return fmt.Errorf("Error reading number of master nodes: %v", err)
	}
	if masterNodes <= 0 {
		return fmt.Errorf("The number of master nodes must be greater than zero")
	}
	workerNodes, err := util.PromptForInt(in, out, "Number of worker nodes", 3)
	if err != nil {
		return fmt.Errorf("Error reading number of worker nodes: %v", err)
	}
	if workerNodes <= 0 {
		return fmt.Errorf("The number of worker nodes must be greater than zero")
	}
	ingressNodes, err := util.PromptForInt(in, out, "Number of ingress nodes (optional, set to 0 if not required)", 2)
	if err != nil {
		return fmt.Errorf("Error reading number of ingress nodes: %v", err)
	}
	if ingressNodes < 0 {
		return fmt.Errorf("The number of ingress nodes must be greater than or equal to zero")
	}
	storageNodes, err := util.PromptForInt(in, out, "Number of storage nodes (optional, set to 0 if not required)", 0)
	if err != nil {
		return fmt.Errorf("Error reading number of storage nodes: %v", err)
	}
	if storageNodes < 0 {
		return fmt.Errorf("The number of storage nodes must be greater than or equal to zero")
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Generating installation plan file template with: \n")
	fmt.Fprintf(out, "- %d etcd nodes\n", etcdNodes)
	fmt.Fprintf(out, "- %d master nodes\n", masterNodes)
	fmt.Fprintf(out, "- %d worker nodes\n", workerNodes)
	fmt.Fprintf(out, "- %d ingress nodes\n", ingressNodes)
	fmt.Fprintf(out, "- %d storage nodes\n", storageNodes)
	fmt.Fprintln(out)

	plan := buildPlan(etcdNodes, masterNodes, workerNodes, ingressNodes)
	// Write out the plan
	if err = install.WritePlanTemplate(plan, planner); err != nil {
		return fmt.Errorf("error planning installation: %v", err)
	}
	fmt.Fprintf(out, "Wrote plan file template to %q\n", planFile)
	fmt.Fprintf(out, "Edit the plan file to further describe your cluster. Once ready, execute the \"install validate\" command to proceed.\n")
	return nil
}

func buildPlan(etcdNodes int, masterNodes int, workerNodes int, ingressNodes int) install.Plan {
	// Create a plan
	masterNodeGroup := install.MasterNodeGroup{}
	masterNodeGroup.ExpectedCount = masterNodes
	plan := install.Plan{
		Etcd: install.NodeGroup{
			ExpectedCount: etcdNodes,
		},
		Master: masterNodeGroup,
		Worker: install.NodeGroup{
			ExpectedCount: workerNodes,
		},
	}

	if ingressNodes > 0 {
		plan.Ingress = install.OptionalNodeGroup{
			ExpectedCount: ingressNodes,
		}
	}

	return plan
}
