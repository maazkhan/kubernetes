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

package kubectl

import (
	"encoding/json"
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"k8s.io/kubernetes/pkg/runtime"
)

type SecretForDockerRegistryGeneratorV1 struct{}

func (s SecretForDockerRegistryGeneratorV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"name", true},
		{"docker-username", true},
		{"docker-email", true},
		{"docker-password", true},
		{"docker-server", true},
	}
}

func (s SecretForDockerRegistryGeneratorV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	// TODO: we seem inconsistent on where parameter validation should exist
	// We need to standardize if this should happen in the Generator or outside of the Generator
	err := ValidateParams(s.ParamNames(), genericParams)
	if err != nil {
		return nil, err
	}
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected string, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}

	dockercfgContent, err := handleDockercfgContent(params["docker-username"], params["docker-password"], params["docker-email"], params["docker-server"])

	if err != nil {
		return nil, err
	}

	secret := &api.Secret{}
	secret.Name = params["name"]
	secret.Type = api.SecretTypeDockercfg
	secret.Data = map[string][]byte{}
	secret.Data[api.DockerConfigKey] = dockercfgContent

	return secret, nil
}

func handleDockercfgContent(username, password, email, server string) ([]byte, error) {
	dockercfgAuth := credentialprovider.DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
	}

	dockerCfg := map[string]credentialprovider.DockerConfigEntry{server: dockercfgAuth}

	return json.Marshal(dockerCfg)
}
