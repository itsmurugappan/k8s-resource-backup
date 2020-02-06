package resources

import (
	"encoding/json"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	iv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"

	"github.com/itsmurugappan/k8s-resource-backup/pkg/storage"
	"github.com/itsmurugappan/k8s-resource-backup/pkg/util"
)

type AuthorizationPolicy struct {
	Client           iv1beta1.SecurityV1beta1Interface
	Result           *v1beta1.AuthorizationPolicyList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitAuthorizationPolicy(config *util.Config) (*AuthorizationPolicy, error) {

	return &AuthorizationPolicy{
		Client:           config.Istio_clientset.SecurityV1beta1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "authorizationpolicy",
	}, nil
}

func (policy *AuthorizationPolicy) Retrieve(opts v1.ListOptions, ns string) {

	list, err := policy.Client.AuthorizationPolicies(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	policy.Result = list
}

func (policy *AuthorizationPolicy) Store(ns string) {
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

func (policy *AuthorizationPolicy) RetrievefromDataStore(ns string) {
	var resultList v1beta1.AuthorizationPolicyList

	dbType := storage.GetDBType(policy.DB, policy.ConnectionString)
	result := dbType.Retrieve(ns, policy.SnapshotID, policy.ResourceName)
	if string(result) == "" {
		log.Print("AuthorizationPolicy not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	policy.Result = &resultList
}

func (kpolicy *AuthorizationPolicy) Restore(ns string) {
	var rs *v1beta1.AuthorizationPolicyList
	if kpolicy.Result == rs {
		return
	}
	for _, policy := range kpolicy.Result.Items {
		modifyAuthPolicyForCreation(&policy)
		_, err := kpolicy.Client.AuthorizationPolicies(ns).Create(&policy)

		if err != nil {
			log.Printf("error creating services : %s", err)
		}
	}
}

func modifyAuthPolicyForCreation(policy *v1beta1.AuthorizationPolicy) {
	policy.ResourceVersion = ""
	policy.UID = ""
	policy.SelfLink = ""
}
