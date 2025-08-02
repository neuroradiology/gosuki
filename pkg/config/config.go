//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"context"
	"fmt"

	"github.com/blob42/gosuki/pkg/logging"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/cli/v3"
)

type Hook func(context.Context, *cli.Command) error

var GlobalConfig = struct {
	//TODO!: generate usage text with //go:zzz directives
	WatchAll bool `toml:"watch-all" mapstructure:"watch-all"`
}{
	false,
}

var (
	log            = logging.GetLogger("CONF")
	ConfReadyHooks []Hook
	configs        = make(map[string]Configurator)
)

const (
	GlobalConfigName = "global"
)

// A Configurator allows multiple packages and modules to set and access configs
// which can be mapped to any output format (toml, cli flags, env variables ...)
type Configurator interface {
	Set(string, any) error
	Get(string) (any, error)
	Dump() map[string]any
	MapFrom(any) error
}

// Global config holder
type Config map[string]any

func (c Config) Set(opt string, v any) error {
	c[opt] = v
	return nil
}

func (c Config) Get(opt string) (any, error) {
	return c[opt], nil
}

func (c Config) Dump() map[string]any {
	return c
}

func (c Config) MapFrom(src any) error {
	mapDecoderConfig.Result = &c
	decoder, err := mapstructure.NewDecoder(mapDecoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}

type AutoConfigurator struct {
	c any
}

func (ac AutoConfigurator) Set(opt string, v any) error {
	// log.Debugf("setting option %s = %v", opt, v)
	s := structs.New(ac.c)
	f, ok := s.FieldOk(opt)
	if !ok {
		return fmt.Errorf("%s option not defined", opt)
	}

	return f.Set(v)
}

func (ac AutoConfigurator) Get(opt string) (any, error) {
	s := structs.New(ac.c)
	f, ok := s.FieldOk(opt)
	if !ok {
		return nil, fmt.Errorf("%s : option not defined", opt)
	}

	return f.Value(), nil
}

func (ac AutoConfigurator) Dump() map[string]any {
	s := structs.New(ac.c)
	return s.Map()
}

func (ac AutoConfigurator) MapFrom(src any) error {
	// log.Debugf("mapping from:  %#v ", src)
	// log.Debugf("mapping to:  %#v ", ac.c)
	mapDecoderConfig.Result = &ac.c
	decoder, err := mapstructure.NewDecoder(mapDecoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}

// AsConfigurator generates implements a default Configurator for a given struct
// or custom type. Use this to handle module options.
func AsConfigurator(c any) Configurator {
	return AutoConfigurator{c}
}

// Register a global option ie. under [global] in toml file
func RegisterGlobalOption(key string, val any) {
	log.Debugf("Registring global option %s = %v", key, val)
	configs[GlobalConfigName].Set(key, val)
}

// Get global option
func GetGlobalOption(key string) (any, error) {
	return configs[GlobalConfigName].Get(key)
}

// GetModuleOption returns a module option value given a module name and option name
func GetModuleOption(module string, opt string) (any, error) {
	if c, ok := configs[module]; ok {
		return c.Get(opt)
	}
	return nil, fmt.Errorf("module %s not found", module)
}


// Register a module option ie. under [module] in toml file
// If the module is not a configurator, a simple map[string]any will be
// created for it. use [GetModuleOption]
func RegisterModuleOpt(module string, opt string, val any) error {
	log.Debugf("Setting option for module <%s>: %s = %v", module, opt, val)
	if _, ok := configs[module]; !ok {
		log.Debugf("Creating new default config for module <%s>", module)
		configs[module] = make(Config)
	}
	dest := configs[module]
	if err := dest.Set(opt, val); err != nil {
		return err
	}

	return nil
}

// Get all configs as a map[string]any
func GetAll() Config {
	result := make(Config)
	for k, c := range configs {
		// if its an AutoConfigurator, use its c field
		if ac, ok := c.(AutoConfigurator); ok {
			result[k] = ac.c
		} else {
			result[k] = c
		}

	}
	return result
}

// Get the config of a module.
func GetModule(module string) Configurator {
	if c, ok := configs[module]; ok {
		return c
	}
	return nil
}

// Hooks registered here will be executed after the config package has finished
// loading the conf
func RegisterConfReadyHooks(hooks ...Hook) {
	ConfReadyHooks = append(ConfReadyHooks, hooks...)
}

// A call to this func will run all registered config hooks
func RunConfHooks(ctx context.Context, cmd *cli.Command) {
	log.Debug("running config hooks")
	for _, f := range ConfReadyHooks {
		err := f(ctx, cmd)
		if err != nil {
			log.Fatalf("error running config hook: %v", err)
		}
	}
}

// A configurator can set options available under it's own module scope
// or under the global scope. A configurator implements the Configurator interface
func RegisterConfigurator(name string, c Configurator) {
	log.Debugf("Registering configurator %s", name)
	configs[name] = c
}

func init() {
	// Initialize the global config
	configs[GlobalConfigName] = AsConfigurator(&GlobalConfig)
}
