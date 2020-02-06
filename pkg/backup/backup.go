package backup

import (
	"log"
	"sync"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackUpInterface interface {
	Retrieve(opts v1.ListOptions, ns string)
	Store(ns string)
	RetrievefromDataStore(ns string)
	Restore(ns string)
}

func BackUp(resourceList []BackUpInterface, namespaces []string) {
	log.Println("start backing up")
	var wg sync.WaitGroup

	for _, resource := range resourceList {
		for _, ns := range namespaces {
			wg.Add(1)
			go backUpWorker(resource, ns, &wg)
		}
	}

	wg.Wait()
}

func backUpWorker(resource BackUpInterface, ns string, wg *sync.WaitGroup) {

	defer wg.Done()

	log.Printf("backing up %s ---- starting", ns)

	listOptions := v1.ListOptions{}

	resource.Retrieve(listOptions, ns)
	resource.Store(ns)

	log.Printf("backing up %s ---- done", ns)
}

func Restore(resourceList []BackUpInterface, namespaces []string) {
	log.Println("start restoring")
	var wg sync.WaitGroup

	for _, resource := range resourceList {
		for _, ns := range namespaces {
			wg.Add(1)
			go restoreWorker(resource, ns, &wg)
		}
	}
	wg.Wait()
}

func restoreWorker(resource BackUpInterface, ns string, wg *sync.WaitGroup) {

	defer wg.Done()

	log.Printf("restoring %s ---- starting", ns)

	resource.RetrievefromDataStore(ns)
	resource.Restore(ns)

	log.Printf("restoring %s ---- done", ns)
}
