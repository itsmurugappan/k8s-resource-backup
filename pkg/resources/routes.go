package resources

import (
	"encoding/json"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	kv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"github.com/itsmurugappan/k8s-resource-backup/pkg/storage"
	"github.com/itsmurugappan/k8s-resource-backup/pkg/util"
)

type Route struct {
	Client           kv1alpha1.ServingV1alpha1Interface
	Result           *v1alpha1.RouteList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitRoute(config *util.Config) (*Route, error) {
	return &Route{
		Client:           config.Kn_clientset.ServingV1alpha1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "route",
	}, nil
}

func (route *Route) Retrieve(opts v1.ListOptions, ns string) {

	list, err := route.Client.Routes(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	route.Result = list
}

func (route *Route) Store(ns string) {
	if len(route.Result.Items) == 0 {
		return
	}

	dbType := storage.GetDBType(route.DB, route.ConnectionString)

	data, err := json.Marshal(route.Result)

	if err != nil {
		log.Printf("error marshalling : %s", err)
	}

	dbType.Store(data, ns, route.SnapshotID, route.ResourceName)
}

func (route *Route) RetrievefromDataStore(ns string) {
	var resultList v1alpha1.RouteList

	dbType := storage.GetDBType(route.DB, route.ConnectionString)
	result := dbType.Retrieve(ns, route.SnapshotID, route.ResourceName)
	if string(result) == "" {
		log.Print("Route not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	route.Result = &resultList
}

func (kroute *Route) Restore(ns string) {
	var rs *v1alpha1.RouteList
	if kroute.Result == rs {
		return
	}

	for _, route := range kroute.Result.Items {
		modifyRouteForCreation(&route)
		_, err := kroute.Client.Routes(ns).Create(&route)

		if err != nil {
			log.Printf("error creating services : %s", err)
		}
	}
}

func modifyRouteForCreation(route *v1alpha1.Route) {
	route.ResourceVersion = ""
	route.UID = ""
	route.SelfLink = ""
}
