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
	Kind:    reflect.TypeOf(clusterv1alpha1.Cluster{}).Name(),
}

// Controller represents a controller object for nfs server custom resources
type Controller struct {
	context        *clusterd.Context
	containerImage string
}

// NewController create controller for watching nfsserver custom resources created
func NewController(context *clusterd.Context, containerImage string) *Controller {
	return &Controller{
		context:        context,
		containerImage: containerImage,
	}
}

// StartWatch watches for instances of nfsserver custom resources and acts on them
func (c *Controller) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	logger.Infof("start watching cluster controller in namespace %s", namespace)
	watcher := opkit.NewWatcher(ClusterResource, namespace, resourceHandlerFuncs, c.context.AlitaClientset.ClusterV1alpha1().RESTClient())
	go watcher.Watch(&clusterv1alpha1.Cluster{}, stopCh)

	return nil
}

func (c *Controller) onAdd(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster).DeepCopy()
	logger.Infof("cluster %s deleted from namespace %s", cluster.Name, cluster.Namespace)
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oldCluster := oldObj.(*clusterv1alpha1.Cluster).DeepCopy()

	logger.Infof("Received update on NFS server %s in namespace %s. This is currently unsupported.", oldCluster.Name, oldCluster.Namespace)
}

func (c *Controller) onDelete(obj interface{}) {
	cluster := obj.(*clusterv1alpha1.Cluster).DeepCopy()
	logger.Infof("cluster %s deleted from namespace %s", cluster.Name, cluster.Namespace)
}

