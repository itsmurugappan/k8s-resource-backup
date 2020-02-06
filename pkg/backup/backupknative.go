package backup

import (
	"log"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/itsmurugappan/k8s-resource-backup/pkg/resources"
	"github.com/itsmurugappan/k8s-resource-backup/pkg/util"
)

func BackUpKnative() {
	config, err := util.GetConfig()
	labels, _ := os.LookupEnv("labels")

	if err != nil {
		log.Printf("error getting config : %s", err)
		return
	}
	//"istio-injection=enabled"
	listOptions := metav1.ListOptions{
		LabelSelector: labels,
	}

	//get ns
	nsList, err := config.K8s_clientset.CoreV1().Namespaces().List(listOptions)

	namespaces := extractNSList(nsList)

	// construct models
	resourceList, err := getResourcestoBackUp(config)

	if err != nil {
		log.Print("please provide db info")
		return
	}

	BackUp(resourceList, namespaces)

}

func RestoreKnative() {
	config, err := util.GetConfig()
	labels, _ := os.LookupEnv("labels")

	if err != nil {
		log.Printf("error getting go client : %s", err)
		return
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels,
	}

	//get ns
	nsList, err := config.K8s_clientset.CoreV1().Namespaces().List(listOptions)

	namespaces := extractNSList(nsList)

	resourceList, err := getResourcestoRestore(config)

	if err != nil {
		log.Print("please provide db info")
		return
	}
	Restore(resourceList, namespaces)

}

func extractNSList(nsList *corev1.NamespaceList) []string {
	var namespaces []string
	exList, present := os.LookupEnv("excluded_ns")
	var excludedNS []string
	if !present {
		log.Println("no exclude list")
	} else {
		excludedNS = strings.Split(exList, ",")
	}

	for _, item := range nsList.Items {
		if !contains(excludedNS, item.ObjectMeta.Name) {
			namespaces = append(namespaces, item.ObjectMeta.Name)
		}
	}

	return namespaces
}

func contains(excludeList []string, ns string) bool {
	for _, e := range excludeList {
		if ns == e {
			return true
		}
	}
	return false
}

func getResourcestoBackUp(config *util.Config) ([]BackUpInterface, error) {

	// istio resources
	// policy, _ := resources.InitPolicy(config)
	// authpolicy, _ := resources.InitAuthorizationPolicy(config)
	rev, _ := resources.InitRevision(config)
	configuration, _ := resources.InitConfiguration(config)
	service, _ := resources.InitKnativeService(config)
	route, _ := resources.InitRoute(config)

	resourceList := []BackUpInterface{rev, configuration, service, route}
	return resourceList, nil
}

func getResourcestoRestore(config *util.Config) ([]BackUpInterface, error) {

	// istio resources
	// policy, _ := resources.InitPolicy(config)
	// authpolicy, _ := resources.InitAuthorizationPolicy(config)
	service, _ := resources.InitKnativeService(config)

	resourceList := []BackUpInterface{service}
	return resourceList, nil
}
