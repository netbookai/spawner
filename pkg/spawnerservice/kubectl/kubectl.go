package kubectl

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getObjectKind(obj runtime.Object) string {
	switch obj.(type) {
	case *appsv1.Deployment:
		return "deployment"
	case *rbacv1.ClusterRole:
		return "clusterrole"
	case *rbacv1.ClusterRoleBinding:
		return "clusterrolebinding"
	case *corev1.Secret:
		return "secret"
	case *corev1.Namespace:
		return "namespace"
	case *corev1.ServiceAccount:
		return "serviceaccount"
	default:
		return "unknown"
	}
}

func getObjectMap(byts []byte) [][]byte {
	splits := strings.Split(string(byts), "---")
	objectList := make([][]byte, 0, len(splits))

	for _, v := range splits {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		objectList = append(objectList, []byte(v))
	}
	return objectList
}

func GetObjects(manifestData []byte) ([]runtime.Object, error) {

	list := getObjectMap(manifestData)
	m := make([]runtime.Object, 0, len(list))

	for _, v := range list {
		var Codec = serializer.NewCodecFactory(Scheme).
			UniversalDecoder(Scheme.PrioritizedVersionsAllGroups()...)
		//LegacyCodec(Scheme.PrioritizedVersionsAllGroups()...)
		data, err := runtime.Decode(Codec, v)
		if err != nil {
			return nil, err
		}
		m = append(m, data)
	}
	return m, nil
}

func GetManifestFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		byts, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		return byts, err
	}
	return nil, errors.New("failed to read yaml")
}

func Apply(ctx context.Context, client *kubernetes.Clientset, dynamicClient dynamic.Interface, objects []runtime.Object) error {

	for _, obj := range objects {
		gvk := obj.GetObjectKind().GroupVersionKind()
		fmt.Printf("applying %+v\n ", gvk)

		gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
		groupResources, err := restmapper.GetAPIGroupResources(client.Discovery())
		if err != nil {
			return err
		}

		rm := restmapper.NewDiscoveryRESTMapper(groupResources)
		mapping, err := rm.RESTMapping(gk, gvk.Version)
		if err != nil {
			return err
		}

		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}

		resp, err := dynamicClient.Resource(mapping.Resource).Create(ctx, &unstructured.Unstructured{Object: objMap}, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("failed to apply resources ...", err)
			continue
		}
		fmt.Println(resp)
		//result := client.RESTClient().Post().VersionedParams(obj, runtime.NewParameterCodec(Scheme)).Do(ctx)
		//fmt.Printf("Err %+v --- %+v\n\n", result.Error(), result)

	}
	return nil
}
