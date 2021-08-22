package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeConfigFile := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	// Here we intend to run our cli outside of the K8S cluster (we will run it from our local terminal)
	// So it is ideal to provide a non-empty kubeConfigFile
	//
	// If you want to run your tool inside cluster, you can use empty params, like
	// config, err := clientcmd.BuildConfigFromFlags("", "")
	// or use the package "k8s.io/client-go/rest"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Connected to " + config.Host)

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	api := clientSet.CoreV1()

	listOptions := v1.ListOptions{
		LabelSelector: "",
		FieldSelector: "",
	}
	pvcs, err := api.PersistentVolumeClaims("monitoring").List(context.Background(), listOptions)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(pvcs.Items)
}
