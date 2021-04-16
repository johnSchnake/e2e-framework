/*
Copyright 2021 The Kubernetes Authors.

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

package conf

import (
	"k8s.io/client-go/rest"

	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

type Config = types.Config

type TestConfig struct {
	namespace string
	kubecfg *rest.Config
}

func New() Config {
	return &TestConfig{}
}

// NewWithKubeCfgFile is a convenience constructor that will
// create a Kubernetes *rest.Config from a file
func NewWithKubecfgFile(filePath string) (Config, error){
	return nil, nil
}

func (c *TestConfig) WithConfig(cfg *rest.Config) Config{
	c.kubecfg = cfg
	return c
}

func (c *TestConfig) WithNamespace(ns string) Config{
	c.namespace = ns
	return c
}

func (c *TestConfig) Env() (types.Environment, error) {
	return nil, nil
}
