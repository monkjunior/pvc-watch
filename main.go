package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	label     = ""
	field     = ""
	namespace = ""
)

func main() {
	flag.StringVar(&label, "labels", "", "pvc labels")
	flag.StringVar(&field, "fields", "", "pvc fields")
	flag.StringVar(&namespace, "namespace", "", "namespace")
	flag.Parse()

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
		LabelSelector: label,
		FieldSelector: field,
	}
	pvcs, err := api.PersistentVolumeClaims(namespace).List(context.Background(), listOptions)
	if err != nil {
		log.Fatalln(err)
	}
	printPVCs(pvcs)

	watcher, err := api.PersistentVolumeClaims(namespace).Watch(context.Background(), listOptions)
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Stop()
	eventsChan := watcher.ResultChan()
	log.Println("Calculating total claimed capacity of PVCs on namespace " + namespace)
	var totalClaimed resource.Quantity
	for e := range eventsChan {
		switch e.Type {
		case watch.Added:
			totalClaimed.Add(*e.Object.(*coreV1.PersistentVolumeClaim).Spec.Resources.Requests.Storage())
			log.Println("Total claimed PVCs: " + totalClaimed.String())
		case watch.Deleted:
			totalClaimed.Sub(*e.Object.(*coreV1.PersistentVolumeClaim).Spec.Resources.Requests.Storage())
			log.Println("Total claimed PVCs: " + totalClaimed.String())
		}
	}
}

func printPVCs(pvcs *coreV1.PersistentVolumeClaimList) {
	template := "%-96v%-8v%-8v\n"
	fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
	for _, pvc := range pvcs.Items {
		fmt.Printf(template, pvc.Name, string(pvc.Status.Phase), pvc.Status.Capacity.Storage().String())
	}
}
