/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package macro

import (
	"io"

	"github.com/spf13/cobra"

	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	namespaceLong = `
Create a namespace with the specified name.`

	namespaceExample = `  // Create a new namespace named my-namespace
  $ kubectl create namespace my-namespace`
)

// NewCmdCreateNamespace is a macro command to create a new namespace
func NewCmdCreateNamespace(f *cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace NAME [--dry-run=bool]",
		Aliases: []string{"ns"},
		Short:   "Create a namespace with the specified name.",
		Long:    namespaceLong,
		Example: namespaceExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateNamespace(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddPrinterFlags(cmd)
	cmd.Flags().String("generator", "", "The name of the API generator to use.  If not specified, default to namespace/v1")
	cmd.Flags().Bool("dry-run", false, "If true, only print the object that would be sent, without sending it.")
	return cmd
}

// CreateNamespace implements the behavior to run the create namespace command
func CreateNamespace(f *cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	return createCommandInternal(f, cmdOut, cmd, args, "namespace/v1", nil)
}
