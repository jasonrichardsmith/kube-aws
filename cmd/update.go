package cmd

import (
	"fmt"

	"bufio"
	"github.com/kubernetes-incubator/kube-aws/core/root"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	cmdUpdate = &cobra.Command{
		Use:          "update",
		Short:        "Update an existing Kubernetes cluster",
		Long:         ``,
		RunE:         runCmdUpdate,
		SilenceUsage: true,
	}

	updateOpts = struct {
		awsDebug, prettyPrint, skipWait bool
		force                           bool
		targets                         []string
	}{}
)

func init() {
	RootCmd.AddCommand(cmdUpdate)
	cmdUpdate.Flags().BoolVar(&updateOpts.awsDebug, "aws-debug", false, "Log debug information from aws-sdk-go library")
	cmdUpdate.Flags().BoolVar(&updateOpts.prettyPrint, "pretty-print", false, "Pretty print the resulting CloudFormation")
	cmdUpdate.Flags().BoolVar(&updateOpts.skipWait, "skip-wait", false, "Don't wait the resources finish")
	cmdUpdate.Flags().BoolVar(&updateOpts.force, "force", false, "Don't ask for confirmation")
	cmdUpdate.Flags().StringSliceVar(&updateOpts.targets, "targets", root.AllOperationTargetsAsStringSlice(), "Update nothing but specified sub-stacks.  Specify `all` or any combination of `etcd`, `control-plane`, and node pool names. Defaults to `all`")
}

func runCmdUpdate(_ *cobra.Command, _ []string) error {
	if !updateOpts.force && !updateConfirmation() {
		fmt.Println("Operation cancelled")
		return nil
	}

	opts := root.NewOptions(updateOpts.prettyPrint, updateOpts.skipWait)

	cluster, err := root.ClusterFromFile(configPath, opts, updateOpts.awsDebug)
	if err != nil {
		return fmt.Errorf("Failed to read cluster config: %v", err)
	}

	targets := root.OperationTargetsFromStringSlice(updateOpts.targets)

	if _, err := cluster.ValidateStack(targets); err != nil {
		return err
	}

	report, err := cluster.Update(targets)
	if err != nil {
		return fmt.Errorf("Error updating cluster: %v", err)
	}
	if report != "" {
		fmt.Printf("Update stack: %s\n", report)
	}

	info, err := cluster.Info()
	if err != nil {
		return fmt.Errorf("Failed fetching cluster info: %v", err)
	}

	successMsg :=
		`Success! Your AWS resources are being updated:
%s
`
	fmt.Printf(successMsg, info.String())

	return nil
}

func updateConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("This operation will update the cluster. Are you sure? [y,n]: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(strings.ToLower(text), "\n")

	return text == "y" || text == "yes"
}
