package olm

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListClusterServiceVersion returns clusterserviceversion inventory in the given namespace.
func ListClusterServiceVersion(
	apiClient *clients.Settings,
	nsname string,
	options ...metav1.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	if nsname == "" {
		glog.V(100).Infof("clusterserviceversion 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list clusterserviceversion, 'nsname' parameter is empty")
	}

	passedOptions := metav1.ListOptions{}
	logMessage := fmt.Sprintf("Listing clusterserviceversion in the namespace %s", nsname)

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	csvList, err := apiClient.OperatorsV1alpha1Interface.ClusterServiceVersions(nsname).List(
		context.TODO(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list clusterserviceversion in the nsname %s due to %s", nsname, err.Error())

		return nil, err
	}

	var csvObjects []*ClusterServiceVersionBuilder

	for _, runningCSV := range csvList.Items {
		copiedCSV := runningCSV
		csvBuilder := &ClusterServiceVersionBuilder{
			apiClient:  apiClient,
			Object:     &copiedCSV,
			Definition: &copiedCSV,
		}

		csvObjects = append(csvObjects, csvBuilder)
	}

	return csvObjects, nil
}

// ListClusterServiceVersionWithNamePattern returns a cluster-wide clusterserviceversion inventory
// filtered by the name pattern.
func ListClusterServiceVersionWithNamePattern(
	apiClient *clients.Settings,
	namePattern string,
	nsname string,
	options ...metav1.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	if namePattern == "" {
		glog.V(100).Info(
			"The namePattern field to filter out all relevant clusterserviceversion cannot be empty")

		return nil, fmt.Errorf(
			"the namePattern field to filter out all relevant clusterserviceversion cannot be empty")
	}

	glog.V(100).Infof("Listing clusterserviceversion filtered by the name pattern %s in %s namespace",
		namePattern, nsname)

	notFilteredCsvList, err := ListClusterServiceVersion(apiClient, nsname, options...)

	if err != nil {
		glog.V(100).Infof("Failed to list all clusterserviceversions in namespace %s due to %s",
			nsname, err.Error())

		return nil, err
	}

	var finalCsvList []*ClusterServiceVersionBuilder

	for _, foundCsv := range notFilteredCsvList {
		if strings.Contains(foundCsv.Definition.Name, namePattern) {
			finalCsvList = append(finalCsvList, foundCsv)
		}
	}

	return finalCsvList, nil
}

// ListClusterServiceVersionInAllNamespaces returns cluster-wide clusterserviceversion inventory.
func ListClusterServiceVersionInAllNamespaces(
	apiClient *clients.Settings,
	options ...metav1.ListOptions) ([]*ClusterServiceVersionBuilder, error) {
	passedOptions := metav1.ListOptions{}
	logMessage := "Listing CSVs in all namespaces"

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	csvList, err := apiClient.ClusterServiceVersions("").List(context.TODO(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list CSVs in all namespaces due to %s", err.Error())

		return nil, err
	}

	var csvObjects []*ClusterServiceVersionBuilder

	for _, csvs := range csvList.Items {
		copiedCSV := csvs
		csvBuilder := &ClusterServiceVersionBuilder{
			apiClient:  apiClient,
			Object:     &copiedCSV,
			Definition: &copiedCSV,
		}

		csvObjects = append(csvObjects, csvBuilder)
	}

	return csvObjects, nil
}
