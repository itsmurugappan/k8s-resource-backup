package resources

import (
	"encoding/json"
	"log"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	kv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"github.com/itsmurugappan/k8s-resource-backup/pkg/storage"
	"github.com/itsmurugappan/k8s-resource-backup/pkg/util"
)

type KnativeService struct {
	Client           kv1alpha1.ServingV1alpha1Interface
	Result           *v1alpha1.ServiceList
	ConnectionString string
	DB               string
	SnapshotID       string
	ResourceName     string
}

func InitKnativeService(config *util.Config) (*KnativeService, error) {

	return &KnativeService{
		Client:           config.Kn_clientset.ServingV1alpha1(),
		ConnectionString: config.CS,
		DB:               config.DB,
		SnapshotID:       config.SnapID,
		ResourceName:     "ksvc",
	}, nil
}

func (svc *KnativeService) Retrieve(opts v1.ListOptions, ns string) {

	list, err := svc.Client.Services(ns).List(opts)

	if err != nil {
		log.Printf("error retrieving services : %s", err)
	}
	svc.Result = list
}

func (svc *KnativeService) Store(ns string) {
	if len(svc.Result.Items) == 0 {
		return
	}

	dbType := storage.GetDBType(svc.DB, svc.ConnectionString)

	data, err := json.Marshal(svc.Result)
	if err != nil {
		log.Printf("error marshalling : %s", err)
	}

	dbType.Store(data, ns, svc.SnapshotID, svc.ResourceName)
}

func (svc *KnativeService) RetrievefromDataStore(ns string) {
	var resultList v1alpha1.ServiceList

	dbType := storage.GetDBType(svc.DB, svc.ConnectionString)
	result := dbType.Retrieve(ns, svc.SnapshotID, svc.ResourceName)
	if string(result) == "" {
		log.Print("KnativeService not record found")
		return
	}

	json.Unmarshal(result, &resultList)

	svc.Result = &resultList
}

func (ksvc *KnativeService) Restore(ns string) {
	var rs *v1alpha1.ServiceList
	if ksvc.Result == rs {
		return
	}

	revList := ksvc.retrieveRevsFromDataStore(ns)

	for _, svc := range ksvc.Result.Items {
		if len(svc.Spec.RouteSpec.Traffic) > 1 {
			revMap := getRevstoRestore(svc)
			var svcCreated *v1alpha1.Service
			for k, _ := range revMap {
				if svcCreated == nil {
					tmpSvc := svc.DeepCopy()
					modifyForCreation(tmpSvc)
					svcCreated = submitService(ksvc.Client, tmpSvc, revList, k, ns, true)
				} else {
					svcCreated = submitService(ksvc.Client, svcCreated, revList, k, ns, false)
				}
			}
			//set percentages
			setPercentages(revMap, svcCreated)
			if _, err := ksvc.Client.Services(ns).Update(svcCreated); err != nil {
				log.Printf("error updating services percentages in %s : %s", ns, err)
			}
		} else {
			modifyForCreation(&svc)
			if _, err := ksvc.Client.Services(ns).Create(&svc); err != nil {
				log.Printf("error creating services in %s : %s", ns, err)
			}
		}
	}
}

func modifyForCreation(svc *v1alpha1.Service) {
	svc.ResourceVersion = ""
	svc.UID = ""
	svc.SelfLink = ""
}

func modifyForUpdate(svc *v1alpha1.Service) {
	svc.UID = ""
}

func setRevision(revName string, svc *v1alpha1.Service, revList *v1alpha1.RevisionList) {
	log.Printf("revsion name : %s", revName)
	percent := int64(100)
	traffic := []v1alpha1.TrafficTarget{
		{
			TrafficTarget: servingv1.TrafficTarget{
				RevisionName: revName,
				Percent:      &percent,
			},
		},
	}
	for _, rev := range revList.Items {
		if revName == rev.ObjectMeta.Name {
			svc.Spec.RouteSpec = v1alpha1.RouteSpec{Traffic: traffic}
			svc.Spec.ConfigurationSpec.Template.Spec = rev.Spec
			svc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = revName
			// svc.Status = v1alpha1.ServiceStatus{}
		}
	}
}

func setPercentages(revMap map[string]*int64, svc *v1alpha1.Service) {
	var trafficList []v1alpha1.TrafficTarget
	for k, v := range revMap {
		traffic := v1alpha1.TrafficTarget{
			TrafficTarget: servingv1.TrafficTarget{
				RevisionName: k,
				Percent:      v,
			},
		}
		trafficList = append(trafficList, traffic)
	}
	svc.Spec.RouteSpec = v1alpha1.RouteSpec{Traffic: trafficList}
}

func getRevstoRestore(svc v1alpha1.Service) map[string]*int64 {
	trafficList := svc.Spec.RouteSpec.Traffic
	revsMap := make(map[string]*int64)

	for _, traffic := range trafficList {
		revsMap[traffic.TrafficTarget.RevisionName] = traffic.TrafficTarget.Percent
	}
	return revsMap
}

func (svc *KnativeService) retrieveRevsFromDataStore(ns string) *v1alpha1.RevisionList {
	var resultList v1alpha1.RevisionList

	dbType := storage.GetDBType(svc.DB, svc.ConnectionString)
	result := dbType.Retrieve(ns, svc.SnapshotID, "revision")
	if string(result) == "" {
		log.Print("revision not record found")
		return &v1alpha1.RevisionList{}
	}

	json.Unmarshal(result, &resultList)

	return &resultList
}

func submitService(client kv1alpha1.ServingV1alpha1Interface, svc *v1alpha1.Service, revList *v1alpha1.RevisionList, revName string, ns string, create bool) *v1alpha1.Service {

	var tmpSvc *v1alpha1.Service
	var err error

	setRevision(revName, svc, revList)

	if create {
		_, err = client.Services(ns).Create(svc)
		if err != nil {
			log.Printf("error creating services in %s : %s", ns, err)
		}
	} else {
		_, err = client.Services(ns).Update(svc)
		if err != nil {
			log.Printf("error updating services in %s : %s", ns, err)
		}
	}
	revReady, err := util.IsRevisionReady(time.Now(), time.Duration(60)*time.Second, client, svc.ObjectMeta.Name, ns)
	if revReady {
		getOpts := v1.GetOptions{}
		tmpSvc, err = client.Services(ns).Get(svc.ObjectMeta.Name, getOpts)
		return tmpSvc.DeepCopy()
	} else {
		log.Printf("error waiting for revision to be ready : %s", err)
		return &v1alpha1.Service{}
	}
}
