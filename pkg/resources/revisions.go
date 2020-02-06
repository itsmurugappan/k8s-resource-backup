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

type Revision struct {
	Client           kv1alpha1.ServingV1alpha1Interface
	Result           *v1alpha1.RevisionList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitRevision(config *util.Config) (*Revision, error) {

	return &Revision{
		Client:           config.Kn_clientset.ServingV1alpha1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "revision",
	}, nil
}

func (revision *Revision) Retrieve(opts v1.ListOptions, ns string) {

	list, err := revision.Client.Revisions(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	revision.Result = list
}

func (revision *Revision) Store(ns string) {
	if len(revision.Result.Items) == 0 {
		return
	}

	dbType := storage.GetDBType(revision.DB, revision.ConnectionString)
	data, err := json.Marshal(revision.Result)

	if err != nil {
		log.Printf("error marshalling : %s", err)
	}

	dbType.Store(data, ns, revision.SnapshotID, revision.ResourceName)
}

func (revision *Revision) RetrievefromDataStore(ns string) {
	var resultList v1alpha1.RevisionList

	dbType := storage.GetDBType(revision.DB, revision.ConnectionString)
	result := dbType.Retrieve(ns, revision.SnapshotID, revision.ResourceName)
	if string(result) == "" {
		log.Print("Revision not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	revision.Result = &resultList
}

func (krevision *Revision) Restore(ns string) {
	var rs *v1alpha1.RevisionList
	if krevision.Result == rs {
		return
	}
	for _, revision := range krevision.Result.Items {
		modifyRevisionForCreation(&revision)
		_, err := krevision.Client.Revisions(ns).Create(&revision)

		if err != nil {
			log.Printf("error creating services : %s", err)
		}
	}
}

func modifyRevisionForCreation(revision *v1alpha1.Revision) {
	revision.ResourceVersion = ""
	revision.UID = ""
	revision.SelfLink = ""
}
