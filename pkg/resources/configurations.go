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

type Configuration struct {
	Client           kv1alpha1.ServingV1alpha1Interface
	Result           *v1alpha1.ConfigurationList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitConfiguration(config *util.Config) (*Configuration, error) {

	return &Configuration{
		Client:           config.Kn_clientset.ServingV1alpha1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "configuration",
	}, nil
}

func (configuration *Configuration) Retrieve(opts v1.ListOptions, ns string) {

	list, err := configuration.Client.Configurations(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	configuration.Result = list
}

func (configuration *Configuration) Store(ns string) {
	if len(configuration.Result.Items) == 0 {
		return
	}

	dbType := storage.GetDBType(configuration.DB, configuration.ConnectionString)

	data, err := json.Marshal(configuration.Result)

	if err != nil {
		log.Printf("error marshalling : %s", err)
	}

	dbType.Store(data, ns, configuration.SnapshotID, configuration.ResourceName)
}

func (configuration *Configuration) RetrievefromDataStore(ns string) {
	var resultList v1alpha1.ConfigurationList

	dbType := storage.GetDBType(configuration.DB, configuration.ConnectionString)
	result := dbType.Retrieve(ns, configuration.SnapshotID, configuration.ResourceName)
	if string(result) == "" {
		log.Print("Configuration not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	configuration.Result = &resultList
}

func (kconfiguration *Configuration) Restore(ns string) {
	var rs *v1alpha1.ConfigurationList
	if kconfiguration.Result == rs {
		return
	}
	for _, configuration := range kconfiguration.Result.Items {
		modifyConfigurationForCreation(&configuration)
		_, err := kconfiguration.Client.Configurations(ns).Create(&configuration)

		if err != nil {
			log.Printf("error creating services : %s", err)
		}
	}
}

func modifyConfigurationForCreation(configuration *v1alpha1.Configuration) {
	configuration.ResourceVersion = ""
	configuration.UID = ""
	configuration.SelfLink = ""
}
