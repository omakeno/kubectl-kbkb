/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	kbkb "github.com/omakeno/kbkb/pkg"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kbkbExample = `
	# view pods in default namespace
	%[1]s kbkb
	
	# view pods in specified namespace
	%[1]s kbkb --namespace your-namespace
	`
)

type KbkbOptions struct {
	namespace  string
	kubeconfig string
	watch      bool
}

func NewKbkbOptions() *KbkbOptions {
	return &KbkbOptions{}
}

func CreateCmd() *cobra.Command {
	o := NewKbkbOptions()
	var rootCmd = &cobra.Command{
		Use:          "kbkb [flags]",
		Short:        "Show pods as kbkb format.",
		Example:      fmt.Sprintf(kbkbExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Execute(c, args); err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "specify namespace to show as kbkb format.")
	rootCmd.PersistentFlags().BoolVarP(&o.watch, "watch", "w", false, "watch kbkb")
	rootCmd.PersistentFlags().StringVarP(&o.kubeconfig, "kubeconfig", "", filepath.Join(homeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	return rootCmd
}

func (o *KbkbOptions) Execute(cmd *cobra.Command, args []string) error {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	o.Watch(clientset)

	return nil
}

func (o *KbkbOptions) Watch(clientset *kubernetes.Clientset) {
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	podInformer := informerFactory.Core().V1().Pods()
	nodeInformer := informerFactory.Core().V1().Nodes()

	printer := kbkb.BashOverwritePrinter{
		Row: 0,
	}

	informedFunc := func() {
		pods, err := podInformer.Lister().Pods(o.namespace).List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}
		nodes, err := nodeInformer.Lister().List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}

		kf := kbkb.BuildKbkbField(pods, nodes)
		kf.PrintAsKbkbOverwrite(&printer)
	}
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { informedFunc() },
		UpdateFunc: func(oldObj, newObj interface{}) { informedFunc() },
		DeleteFunc: func(obj interface{}) { informedFunc() },
	})
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { informedFunc() },
		UpdateFunc: func(oldObj, newObj interface{}) { informedFunc() },
		DeleteFunc: func(obj interface{}) { informedFunc() },
	})

	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	if o.watch {
		for {
			time.Sleep(time.Second)
		}
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
