package resources

import (
	"encoding/json"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	iv1alpha1 "istio.io/client-go/pkg/clientset/versioned/typed/authentication/v1alpha1"

	"github.com/itsmurugappan/k8s-resource-backup/pkg/storage"
	"github.com/itsmurugappan/k8s-resource-backup/pkg/util"
)

type Policy struct {
	Client           iv1alpha1.AuthenticationV1alpha1Interface
	Result           *v1alpha1.PolicyList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitPolicy(config *util.Config) (*Policy, error) {

	return &Policy{
		Client:           config.Istio_clientset.AuthenticationV1alpha1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "policy",
	}, nil
}

func (policy *Policy) Retrieve(opts v1.ListOptions, ns string) {

	list, err := policy.Client.Policies(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	policy.Result = list
}

func (policy *Policy) Store(ns string) {
	if len(policy.Result.Items) == 0 {
		return
	}

	dbType := storage.GetDBType(policy.DB, policy.ConnectionString)

	data, err := json.Marshal(policy.Result)

	if err != nil {
		log.Printf("error marshalling : %s", err)
	}

	dbType.Store(data, ns, policy.SnapshotID, policy.ResourceName)
}

func (policy *Policy) RetrievefromDataStore(ns string) {
	var resultList v1alpha1.PolicyList

	dbType := storage.GetDBType(policy.DB, policy.ConnectionString)
	result := dbType.Retrieve(ns, policy.SnapshotID, policy.ResourceName)
	if string(result) == "" {
		log.Print("Policy not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	policy.Result = &resultList
}

func (kpolicy *Policy) Restore(ns string) {
	var rs *v1alpha1.PolicyList
	if kpolicy.Result == rs {
		return
	}
	for _, policy := range kpolicy.Result.Items {
		modifyPolicyForCreation(&policy)
		_, err := kpolicy.Client.Policies(ns).Create(&policy)

		if err != nil {
			log.Printf("error creating services : %s", err)
		}
	}
}

func modifyPolicyForCreation(policy *v1alpha1.Policy) {
	policy.ResourceVersion = ""
	policy.UID = ""
	policy.SelfLink = ""
}
