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

// Package nfs to manage an NFS export.
package cluster

import (
	"fmt"
	"reflect"
	s "strings"

	"github.com/coreos/pkg/capnslog"
	opkit "github.com/rook/operator-kit"
	clusterv1alpha1 "github.com/alita/alita/pkg/apis/cluster/v1alpha1"
	"github.com/alita/alita/pkg/clusterd"
	"github.com/alita/alita/pkg/operator/k8sutil"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/cache"
)

const (
	customResourceName       = "clustercontroller"
	customResourceNamePlural = "clustercontrollers"
	appName                  = "clustercontroller"
)

var logger = capnslog.NewPackageLogger("github.com/alita/alita", "cluster-controller-operator")

// NFSResource represents the nfs export custom resource
var ClusterResource = opkit.CustomResource{
	Name:    customResourceName,
	Plural:  customResourceNamePlural,
	Group:   clusterv1alpha1.CustomResourceGroup,
	Version: clusterv1alpha1.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(nfsv1alpha1.NFSServer{}).Name(),
}

// Controller represents a controller object for nfs server custom resources
type ClusterController struct {
	context        *clusterd.Context
	containerImage string
}

// NewController create controller for watching nfsserver custom resources created
func NewClusterController(context *clusterd.Context, containerImage string) *ClusterController {
	return &ClusterController{
		context:        context,
		containerImage: containerImage,
	}
}

// StartWatch watches for instances of nfsserver custom resources and acts on them
func (c *ClusterController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	logger.Infof("start watching cluster controller in namespace %s", namespace)
	watcher := opkit.NewWatcher(ClusterResource, namespace, resourceHandlerFuncs, c.context.AlitaClientset.SlurmV1alpha1())
	go watcher.Watch(&nfsv1alpha1.NFSServer{}, stopCh)

	return nil
}

func getServerConfig(exports []nfsv1alpha1.ExportsSpec) map[string]map[string]string {
	claimConfigOpt := make(map[string]map[string]string)
	configOpt := make(map[string]string)

	for _, export := range exports {
		claimName := export.PersistentVolumeClaim.ClaimName
		if claimName != "" {
			configOpt["accessMode"] = export.Server.AccessMode
			configOpt["squash"] = export.Server.Squash
			claimConfigOpt[claimName] = configOpt
		}
	}

	return claimConfigOpt
}

func createAppLabels() map[string]string {
	return map[string]string{
		k8sutil.AppAttr: appName,
	}
}

func createServicePorts() []v1.ServicePort {
	return []v1.ServicePort{
		{
			Name:       "nfs",
			Port:       int32(nfsPort),
			TargetPort: intstr.FromInt(int(nfsPort)),
		},
		{
			Name:       "rpc",
			Port:       int32(rpcPort),
			TargetPort: intstr.FromInt(int(rpcPort)),
		},
	}
}


func (c *Controller) onAdd(obj interface{}) {
	nfsObj := obj.(*nfsv1alpha1.NFSServer).DeepCopy()

	nfsServer := newNfsServer(nfsObj, c.context)

	logger.Infof("new NFS server %s added to namespace %s", nfsObj.Name, nfsServer.namespace)

	logger.Infof("validating nfs server spec in namespace %s", nfsServer.namespace)
	if err := validateNFSServerSpec(nfsServer.spec); err != nil {
		logger.Errorf("Invalid NFS Server spec: %+v", err)
		return
	}

	logger.Infof("creating nfs server service in namespace %s", nfsServer.namespace)
	if err := c.createNFSService(nfsServer); err != nil {
		logger.Errorf("Unable to create NFS service %+v", err)
	}

	logger.Infof("creating nfs server configuration in namespace %s", nfsServer.namespace)
	if err := c.createNFSConfigMap(nfsServer); err != nil {
		logger.Errorf("Unable to create NFS ConfigMap %+v", err)
	}

	logger.Infof("creating nfs server stateful set in namespace %s", nfsServer.namespace)
	if err := c.createNfsStatefulSet(nfsServer, int32(nfsServer.spec.Replicas)); err != nil {
		logger.Errorf("Unable to create NFS stateful set %+v", err)
	}
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oldNfsServ := oldObj.(*nfsv1alpha1.NFSServer).DeepCopy()

	logger.Infof("Received update on NFS server %s in namespace %s. This is currently unsupported.", oldNfsServ.Name, oldNfsServ.Namespace)
}

func (c *Controller) onDelete(obj interface{}) {
	cluster := obj.(*nfsv1alpha1.NFSServer).DeepCopy()
	logger.Infof("cluster %s deleted from namespace %s", cluster.Name, cluster.Namespace)
}

