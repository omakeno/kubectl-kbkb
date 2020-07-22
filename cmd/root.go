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
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
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

	printer := BashOverwritePrinter{row: 0}

	informedFunc := func() {
		pods, err := podInformer.Lister().Pods(o.namespace).List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}
		nodes, err := nodeInformer.Lister().List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}

		kf := BuildKubeField(pods, nodes)
		kf.printAsKbkbOverwrite(&printer)
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

type kubeState struct {
	Color    string
	IsStable bool
}

func (ks *kubeState) String() string {
	return ks.Color
}

func (ks *kubeState) ColoredString() string {
	var colorMap = map[string]string{
		"red":    "\033[0;31m",
		"green":  "\033[0;32m",
		"yellow": "\033[0;33m",
		"blue":   "\033[0;34m",
		"purple": "\033[0;35m",
		"white":  "",
	}

	var iconMap = map[bool]string{
		true:  "●",
		false: "o",
	}
	return colorMap[ks.Color] + iconMap[ks.IsStable] + "\033[0m"
}

type kubeField [][]kubeState

func BuildKubeField(p []*v1.Pod, n []*v1.Node) kubeField {
	var nodes []*v1.Node = make([]*v1.Node, len(n))
	copy(nodes, n)
	var pods []*v1.Pod = make([]*v1.Pod, len(p))
	copy(pods, p)

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].CreationTimestamp.UnixNano() == nodes[j].CreationTimestamp.UnixNano() {
			return nodes[i].Name < nodes[j].Name
		} else {
			return nodes[i].CreationTimestamp.UnixNano() < nodes[j].CreationTimestamp.UnixNano()
		}
	})
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].CreationTimestamp.UnixNano() == pods[j].CreationTimestamp.UnixNano() {
			return pods[i].Name < pods[j].Name
		} else {
			return pods[i].CreationTimestamp.UnixNano() < pods[j].CreationTimestamp.UnixNano()
		}
	})

	var kf [][]kubeState
	kf = make([][]kubeState, len(nodes))

	nodenameToIndex := map[string]int{}
	for i, node := range nodes {
		nodenameToIndex[node.Name] = i
	}

	for _, pod := range pods {
		kf[nodenameToIndex[pod.Spec.NodeName]] = append(kf[nodenameToIndex[pod.Spec.NodeName]], getKube(pod))
	}

	return kubeField(kf)
}

func (kf kubeField) printAsKbkbOverwrite(p *BashOverwritePrinter) {
	out := strings.Repeat("-", len(kf)+2) + "\n"
	i := 0
	for {
		line := "|"
		empty := true
		for _, col := range kf {
			if len(col) > i {
				line += col[i].ColoredString()
				empty = false
			} else {
				line += " "
			}
		}
		out = line + "|\n" + out
		i++
		if empty {
			break
		}
	}
	p.Print(out)
}

func getKube(pod *v1.Pod) kubeState {

	var IsStable bool = true
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			IsStable = false
			break
		}
	}

	var color string = "white"
	c, ok := pod.ObjectMeta.Annotations["kubeColor"]
	if ok {
		color = c
	}

	return kubeState{
		Color:    color,
		IsStable: IsStable,
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
