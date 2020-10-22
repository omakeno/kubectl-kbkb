package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bashoverwriter "github.com/omakeno/bashoverwriter/pkg"
	kbkb "github.com/omakeno/kbkb/pkg"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kbkbExample = `
	# view pods in default namespace
	%[1]s kbkb
	
	# view pods in specified namespace
	%[1]s kbkb --namespace your-namespace

	# watch pods
	%[1]s kbkb --watch

	# view pods with large size (monospaced font required)
	%[1]s kbkb --large

	# color pods by labels, not annotation (for demonstration)
	%[1]s kbkb --demo
	`
)

type KbkbOptions struct {
	namespace  string
	kubeconfig string
	watch      bool
	large      bool
	demo       bool
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
	rootCmd.PersistentFlags().BoolVarP(&o.large, "large", "L", false, "view on large size")
	rootCmd.PersistentFlags().BoolVarP(&o.demo, "demo", "", false, "demonstrate kbkb with color (label-hash)")
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

	if o.watch {
		o.Watch(clientset)
	} else {
		o.Get(clientset)
	}

	return nil
}
func (o *KbkbOptions) Get(clientset *kubernetes.Clientset) {
	podList, err := clientset.CoreV1().Pods(o.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	var kf kbkb.KbkbField
	if o.demo {
		kf = kbkb.BuildKbkbFieldFromList(podList, nodeList, &kbkb.HashedPodGenerator{})
	} else {
		kf = kbkb.BuildKbkbFieldFromList(podList, nodeList, &kbkb.AnnotatedPodGenerator{})
	}
	writer := bashoverwriter.GetBashoverwriter()
	var kcs kbkb.KbkbCharSet
	if o.large {
		kcs = kbkb.GetKbkbCharSetWide()
	} else {
		kcs = kbkb.GetKbkbCharSet()
	}
	kcs.PrintKbkb(&writer, kf)
}

func (o *KbkbOptions) Watch(clientset *kubernetes.Clientset) {
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	podInformer := informerFactory.Core().V1().Pods()
	nodeInformer := informerFactory.Core().V1().Nodes()

	writer := bashoverwriter.GetBashoverwriter()

	informedFunc := func() {
		pods, err := podInformer.Lister().Pods(o.namespace).List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}
		nodes, err := nodeInformer.Lister().List(labels.NewSelector())
		if err != nil {
			panic(err.Error())
		}

		var kf kbkb.KbkbField
		if o.demo {
			kf = kbkb.BuildKbkbField(pods, nodes, &kbkb.HashedPodGenerator{})
		} else {
			kf = kbkb.BuildKbkbField(pods, nodes, &kbkb.AnnotatedPodGenerator{})
		}
		var kcs kbkb.KbkbCharSet
		if o.large {
			kcs = kbkb.GetKbkbCharSetWide()
		} else {
			kcs = kbkb.GetKbkbCharSet()
		}
		kcs.PrintKbkb(&writer, kf)
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

	for {
		time.Sleep(time.Second)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
