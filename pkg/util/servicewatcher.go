package util

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"

	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	kv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

func IsRevisionReady(start time.Time, timeout time.Duration, kcs kv1alpha1.ServingV1alpha1Interface, revName string, ns string) (bool, error) {

	listOptions := metav1.ListOptions{
		FieldSelector: "metadata.name=" + revName,
	}

	watcher, _ := kcs.Services(ns).Watch(listOptions)

	defer watcher.Stop()
	for {
		select {
		case <-time.After(timeout):
			return false, nil
		case event, ok := <-watcher.ResultChan():
			if !ok || event.Object == nil {
				return true, nil
			}

			// Skip event if generations has not yet been consolidated
			inSync, err := isGivenEqualsObservedGeneration(event.Object)
			if err != nil {
				return false, err
			}
			if !inSync {
				continue
			}

			conditions, err := serviceConditionExtractor(event.Object)
			if err != nil {
				return false, err
			}
			for _, cond := range conditions {
				if cond.Type == apis.ConditionReady {
					switch cond.Status {
					case corev1.ConditionTrue:
						return true, nil
					case corev1.ConditionFalse:
						return false, fmt.Errorf("%s: %s", cond.Reason, cond.Message)
					}
				}
			}
		}
	}
}

func isGivenEqualsObservedGeneration(object runtime.Object) (bool, error) {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return false, err
	}
	meta, ok := unstructured["metadata"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("cannot extract metadata from %v", object)
	}
	status, ok := unstructured["status"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("cannot extract status from %v", object)
	}
	observedGeneration, ok := status["observedGeneration"]
	if !ok {
		// Can be the case if not status has been attached yet
		return false, nil
	}
	givenGeneration, ok := meta["generation"]
	if !ok {
		return false, fmt.Errorf("no field 'generation' in metadata of %v", object)
	}
	return givenGeneration == observedGeneration, nil
}

// func serviceConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
// 	rev, ok := obj.(*v1alpha1.Revision)
// 	if !ok {
// 		return nil, fmt.Errorf("%v is not a revision", obj)
// 	}
// 	return apis.Conditions(rev.Status.Status.Conditions), nil
// }

func serviceConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	service, ok := obj.(*v1alpha1.Service)
	if !ok {
		return nil, fmt.Errorf("%v is not a service", obj)
	}
	return apis.Conditions(service.Status.Conditions), nil
}
