/*
Copyright 2018 The Rook Authors. All rights reserved.

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

package rook

import (
	"fmt"
	"os"
	"strings"

	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	alitaclient "github.com/alita/alita/pkg/client/clientset/versioned"
	"github.com/alita/alita/pkg/util/flags"
	"github.com/alita/alita/pkg/version"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

const (
	RookEnvVarPrefix = "ALITA"
	terminationLog   = "/dev/termination-log"
)

var RootCmd = &cobra.Command{
	Use: "alita",
}

var (
	logLevelRaw string
	Cfg         = &Config{}
	logger      = capnslog.NewPackageLogger("github.com/alita/alita", "alitacmd")
)

type Config struct {
	LogLevel capnslog.LogLevel
}

// Initialize the configuration parameters. The precedence from lowest to highest is:
//  1) default value (at compilation)
//  2) environment variables (upper case, replace - with _, and rook prefix. For example, discovery-url is ROOK_DISCOVERY_URL)
//  3) command line parameter
func init() {
	RootCmd.PersistentFlags().StringVar(&logLevelRaw, "log-level", "INFO", "logging level for logging/tracing output (valid values: CRITICAL,ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE)")

	// load the environment variables
	flags.SetFlagsFromEnv(RootCmd.Flags(), AlitaEnvVarPrefix)
	flags.SetFlagsFromEnv(RootCmd.PersistentFlags(), AlitaEnvVarPrefix)
}

// SetLogLevel set log level based on provided log option.
func SetLogLevel() {
	// parse given log level string then set up corresponding global logging level
	ll, err := capnslog.ParseLevel(logLevelRaw)
	if err != nil {
		logger.Warningf("failed to set log level %s. %+v", logLevelRaw, err)
	}
	Cfg.LogLevel = ll
	capnslog.SetGlobalLogLevel(Cfg.LogLevel)
}

// LogStartupInfo log the version number, arguments, and all final flag values (environment variable overrides have already been taken into account)
func LogStartupInfo(cmdFlags *pflag.FlagSet) {

	flagValues := flags.GetFlagsAndValues(cmdFlags, "secret|keyring")
	logger.Infof("starting Alita %s with arguments '%s'", version.Version, strings.Join(os.Args, " "))
	logger.Infof("flag values: %s", strings.Join(flagValues, ", "))
}

// GetClientset create the k8s client
func GetClientset() (kubernetes.Interface, apiextensionsclient.Interface, alitaclient.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get k8s config. %+v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s clientset. %+v", err)
	}
	apiExtClientset, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s API extension clientset. %+v", err)
	}
	alitaClientset, err := alitaclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create rook clientset. %+v", err)
	}
	return clientset, apiExtClientset, alitaClientset, nil
}

// TerminateFatal terminates the process with an exit code of 1 and writes the given reason to stderr and // the termination log file.
func TerminateFatal(reason error) {
	fmt.Fprintln(os.Stderr, reason)

	file, err := os.OpenFile(terminationLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to write message to termination log: %+v", err))
	} else {
		defer file.Close()
		if _, err = file.WriteString(reason.Error()); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("failed to write message to termination log: %+v", err))
		}
	}

	os.Exit(1)
}
