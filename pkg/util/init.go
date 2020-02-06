package util

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	iclient "istio.io/client-go/pkg/clientset/versioned"
	kclient "knative.dev/serving/pkg/client/clientset/versioned"
)

type Config struct {
	K8s_clientset   *kubernetes.Clientset
	Istio_clientset *iclient.Clientset
	Kn_clientset    *kclient.Clientset
	CS              string
	DB              string
	SnapID          string
}

func GetConfig() (*Config, error) {
	restConfig, err := kubeConfig()

	if err != nil {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error getting kube config %s :", err)
			return &Config{}, err
		}
	}

	k8s_clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Printf("error getting k8s clientset %s :", err)
		return &Config{}, err
	}

	istio_clientset, err := iclient.NewForConfig(restConfig)
	if err != nil {
		log.Printf("error getting istio clientset %s :", err)
		return &Config{}, err
	}

	kn_clientset, err := kclient.NewForConfig(restConfig)
	if err != nil {
		log.Printf("error getting knative clientset %s :", err)
		return &Config{}, err
	}

	cs, present_cs := os.LookupEnv("storageinfo")
	db, present_db := os.LookupEnv("db")
	snapID, present_sid := os.LookupEnv("snapshot_id")

	if !present_cs || !present_db || !present_sid {
		return &Config{}, fmt.Errorf("please provide db info")
	}

	config := &Config{k8s_clientset, istio_clientset, kn_clientset, cs, db, snapID}

	return config, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func kubeConfig() (*rest.Config, error) {
	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}
