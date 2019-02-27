/*
Copyright 2018 The Alita Authors. All rights reserved.

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

package scheduler

import (
	"fmt"
	"github.com/alita/alita/pkg/clusterd"
	"github.com/alita/alita/pkg/operator/k8sutil"
	"github.com/alita/alita/pkg/util/flags"

	"github.com/alita/alita/cmd/alita/alita"
	operator "github.com/alita/alita/pkg/operator/scheduler"
	"github.com/spf13/cobra"
)

const containerName = "alita-scheduler-operator"

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Runs alita-scheduler for non-native HPC workload management system in kubernetes clusters",
	Long: `Runs alita-scheduler for non-native HPC workload management system in kubernetes clusters`,
}

func init() {
	flags.SetFlagsFromEnv(operatorCmd.Flags(), rook.RookEnvVarPrefix)
	flags.SetLoggingFlags(operatorCmd.Flags())

	operatorCmd.RunE = startOperator
}

func startOperator(cmd *cobra.Command, args []string) error {
	alita.SetLogLevel()
	rook.LogStartupInfo(operatorCmd.Flags())

	clientset, apiExtClientset, alitaClientset, err := alita.GetClientset()
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to get k8s clients. %+v\n", err))
	}

	logger.Infof("starting scheduler operator")
	context := createContext()
	context.NetworkInfo = clusterd.NetworkInfo{}
	context.ConfigDir = k8sutil.DataDir
	context.Clientset = clientset
	context.APIExtensionClientset = apiExtClientset
	context.AlitaClientset = alitaClientset

	alitaImage := ""
	op := operator.New(context, alitaImage)
	err = op.Run()
	if err != nil {
		rook.TerminateFatal(fmt.Errorf("failed to run operator. %+v\n", err))
	}

	return nil
}
