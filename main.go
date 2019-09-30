package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	namespace := flag.String("namespace", "", "(required) namespace to target")
	interval := flag.Int("interval", 30, "(optional) interval in which the logs are retrieved")
	selector := flag.String("selector", "", "(optional) Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. key1=value1,key2=value2")

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if *namespace == "" {
		flag.PrintDefaults()
		log.Fatalln("the namespace argument is required")
	}

	var client *Client
	if _, err := os.Stat(*kubeconfig); os.IsNotExist(err) {
		// if kubeconfig doesn't exist or is not provided, create in-cluster client
		log.Println("attempting to connect using an in-cluster client")
		client, err = NewInclusterClient()
		if err != nil {
			log.Fatalf("error creating an in-cluster client: %v", err.Error())
		}
	} else {
		// if kubeconfig exists, create out-of-cluster client
		client, err = NewKubeconfigClient(*kubeconfig)
		log.Println("attempting to connect using kubeconfig")
		if err != nil {
			log.Fatalf("error creating an out-of-cluster client: %v", err.Error())
		}
	}

	// Get server version to validate we can connect
	version, err := client.Clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("error retrieving cluster version: %v", err.Error())
	}
	log.Printf("connected to cluster, server version is %+v\n", version)

	// check if namespace exists and we have access
	_, err = client.Clientset.CoreV1().Namespaces().Get(*namespace, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	// check if we have access to `pods/log` in the namespace
	sar, err := client.Clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(&authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Resource:  "pods/log",
				Verb:      "get",
				Namespace: *namespace,
			},
		},
	})
	if err != nil {
		log.Fatalf("error validating permissions to get logs: %+v\n", err)
	}
	if !sar.Status.Allowed {
		log.Fatalf("this user is not allowed to get logs for pods in namespace %s", *namespace)
	}

	var since int64
	intervalDuration := time.Duration(*interval)
	since = int64((intervalDuration * time.Second).Seconds())
	ticker := time.NewTicker(intervalDuration * time.Second)
	log.Printf("will retrieve logs every %d seconds", intervalDuration)
	for range ticker.C {
		log.Println("refreshing pods list")
		pods, _ := client.Clientset.CoreV1().Pods(*namespace).List(metav1.ListOptions{LabelSelector: *selector})
		if len(pods.Items) == 0 {
			log.Printf("Warning: no pods found in namespace %s", *namespace)
		}
		log.Printf("retrieving logs from %d pods", len(pods.Items))
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				logs, err := client.GetPodLogs(pod, container, since)
				if err != nil {
					log.Fatalf("error retrieving logs for pod/container %s/%s: %v", pod.Name, container.Name, err)
				}
				fmt.Println(logs)
			}
		}
	}
}
