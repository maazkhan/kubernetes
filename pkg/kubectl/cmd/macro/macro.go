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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

// commandParamVisitorFunc has an opportunity to visit each parameter in specified command
type commandParamVisitorFunc func(cmd *cobra.Command, params map[string]interface{})

// commandParamStringSliceVisitor modifies the parameter value for each enumerated
// parameter to be a string slice rather than a single string.  useful for
// command parameters that may be specified multiple times.
func commandParamStringSliceVisitor(stringSliceParams []string) commandParamVisitorFunc {
	return func(cmd *cobra.Command, params map[string]interface{}) {
		for _, paramName := range stringSliceParams {
			params[paramName] = cmdutil.GetFlagStringSlice(cmd, paramName)
		}
	}
}

// createCommandInternal provides the common utility code for create subcommands.
func createCommandInternal(f *cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command,
	args []string, defaultGenerator string, visitor commandParamVisitorFunc) error {
	if len(args) == 0 {
		return cmdutil.UsageError(cmd, "NAME is required")
	}
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	generatorName := cmdutil.GetFlagString(cmd, "generator")
	if len(generatorName) == 0 {
		generatorName = defaultGenerator
	}
	generator, found := f.Generator(generatorName)
	if !found {
		return cmdutil.UsageError(cmd, fmt.Sprintf("Generator: %s not found.", generatorName))
	}
	names := generator.ParamNames()
	params := kubectl.MakeParams(cmd, names)
	params["name"] = args[0]
	if len(args) > 1 {
		params["args"] = args[1:]
	}
	if visitor != nil {
		visitor(cmd, params)
	}
	err = kubectl.ValidateParams(names, params)
	if err != nil {
		return err
	}
	obj, err := generator.Generate(params)
	if err != nil {
		return err
	}
	mapper, typer := f.Object()
	version, kind, err := typer.ObjectVersionAndKind(obj)
	if err != nil {
		return err
	}
	mapping, err := mapper.RESTMapping(kind, version)
	if err != nil {
		return err
	}
	client, err := f.RESTClient(mapping)
	if err != nil {
		return err
	}
	if !cmdutil.GetFlagBool(cmd, "dry-run") {
		resourceMapper := &resource.Mapper{ObjectTyper: typer, RESTMapper: mapper, ClientMapper: f.ClientMapperForCommand()}
		info, err := resourceMapper.InfoForObject(obj)
		if err != nil {
			return err
		}
		// Serialize the configuration into an annotation.
		if err := kubectl.UpdateApplyAnnotation(info); err != nil {
			return err
		}
		obj, err = resource.NewHelper(client, mapping).Create(namespace, false, info.Object)
		if err != nil {
			return err
		}
	}
	outputFormat := cmdutil.GetFlagString(cmd, "output")
	if outputFormat != "" {
		return f.PrintObject(cmd, obj, cmdOut)
	}
	cmdutil.PrintSuccess(mapper, false, cmdOut, mapping.Resource, args[0], "created")
	return nil
}
