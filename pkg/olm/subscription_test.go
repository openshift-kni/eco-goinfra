package olm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	oplmV1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewSubscriptionBuilder(t *testing.T) { //nolint:funlen
	testCases := []struct {
		name                   string
		namespace              string
		catalogSource          string
		catalogSourceNamespace string
		packageName            string
		expectedError          string
		client                 bool
	}{
		{
			name:                   "subscription",
			namespace:              "test-namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "test-namespace",
			packageName:            "package-test",
			client:                 true,
			expectedError:          "",
		},
		{
			name:                   "",
			namespace:              "test-namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "test-namespace",
			packageName:            "package-test",
			client:                 true,
			expectedError:          "subscription 'subName' cannot be empty",
		},
		{
			name:                   "subscription",
			namespace:              "",
			catalogSource:          "test-source",
			catalogSourceNamespace: "test-namespace",
			packageName:            "package-test",
			client:                 true,
			expectedError:          "subscription 'subNamespace' cannot be empty",
		},
		{
			name:                   "subscription",
			namespace:              "test-namespace",
			catalogSource:          "",
			catalogSourceNamespace: "test-namespace",
			packageName:            "package-test",
			client:                 true,
			expectedError:          "subscription 'catalogSource' cannot be empty",
		},
		{
			name:                   "subscription",
			namespace:              "test-namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "",
			packageName:            "package-test",
			client:                 true,
			expectedError:          "subscription 'catalogSourceNamespace' cannot be empty",
		},
		{
			name:                   "subscription",
			namespace:              "test-namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "test-namespace",
			packageName:            "",
			client:                 true,
			expectedError:          "subscription 'packageName' cannot be empty",
		},
		{
			name:                   "subscription",
			namespace:              "test-namespace",
			catalogSource:          "test-source",
			catalogSourceNamespace: "test-namespace",
			packageName:            "package-test",
			client:                 false,
			expectedError:          "",
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		}

		subscription := NewSubscriptionBuilder(
			testSettings,
			testCase.name,
			testCase.namespace,
			testCase.catalogSource,
			testCase.catalogSourceNamespace,
			testCase.packageName)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, subscription.errorMsg)

			if testCase.expectedError == "" {
				assert.NotNil(t, subscription.Definition)
				assert.Equal(t, testCase.name, subscription.Definition.Name)
				assert.Equal(t, testCase.namespace, subscription.Definition.Namespace)
			}
		} else {
			assert.Nil(t, subscription)
		}
	}
}

func TestPullSubscription(t *testing.T) {
	subscription := func(name, namespace string) *oplmV1alpha1.Subscription {
		return &oplmV1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: &oplmV1alpha1.SubscriptionSpec{
				CatalogSource: "test",
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "subscription",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("subscription 'subName' cannot be empty"),
			client:              true,
		},
		{
			name:                "subscription",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("subscription 'subNamespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "subscription",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"subscription object named subscription does not exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                "subscription",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("subscription 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		subscription := subscription(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, subscription)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := PullSubscription(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestSubscriptionGet(t *testing.T) {
	testCases := []struct {
		subscription  *SubscriptionBuilder
		expectedError string
	}{
		{
			subscription:  buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: "",
		},
		{
			subscription:  buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: "subscription 'subNamespace' cannot be empty",
		},
		{
			subscription: buildValidSubscriptionBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "subscriptions.operators.coreos.com \"subscription\" not found",
		},
	}

	for _, testCase := range testCases {
		subscription, err := testCase.subscription.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, subscription.Name, testCase.subscription.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestSubscriptionExist(t *testing.T) {
	testCases := []struct {
		subscription   *SubscriptionBuilder
		expectedStatus bool
	}{
		{
			subscription:   buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			subscription: buildValidSubscriptionBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.subscription.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestSubscriptionDelete(t *testing.T) {
	testCases := []struct {
		subscription  *SubscriptionBuilder
		expectedError error
	}{
		{
			subscription:  buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: nil,
		},
		{
			subscription: buildValidSubscriptionBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.subscription.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.subscription.Object)
		}
	}
}

func TestSubscriptionUpdate(t *testing.T) {
	testCases := []struct {
		subscription  *SubscriptionBuilder
		expectedError error
		startingCSV   string
	}{
		{
			subscription:  buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: nil,
			startingCSV:   "test",
		},
		{
			subscription:  buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: fmt.Errorf("subscription 'subNamespace' cannot be empty"),
			startingCSV:   "",
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.subscription.Definition.Spec.StartingCSV)
		assert.Nil(t, nil, testCase.subscription.Object)
		testCase.subscription.Definition.Spec.StartingCSV = testCase.startingCSV
		testCase.subscription.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.subscription.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.startingCSV, testCase.subscription.Object.Spec.StartingCSV)
		}
	}
}

func TestSubscriptionCreate(t *testing.T) {
	testCases := []struct {
		subscription  *SubscriptionBuilder
		expectedError error
	}{
		{
			subscription:  buildValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: nil,
		},
		{
			subscription:  buildInValidSubscriptionBuilder(buildTestSubscriptionClientWithDummyObject()),
			expectedError: fmt.Errorf("subscription 'subNamespace' cannot be empty"),
		},
		{
			subscription: buildValidSubscriptionBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		catalogSourceBuilder, err := testCase.subscription.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, catalogSourceBuilder.Definition.Name, catalogSourceBuilder.Object.Name)
		}
	}
}

func TestSubscriptionWithChannel(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
		channel           string
	}{
		{
			channel:           "test",
			expectedErrorText: "",
		},
		{
			channel:           "",
			expectedErrorText: "can not redefine subscription with empty channel",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestSubscriptionClientWithDummyObject()
		subscriptionBuilder := buildValidSubscriptionBuilder(testSettings).WithChannel(testCase.channel)
		assert.Equal(t, subscriptionBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, subscriptionBuilder.Definition.Spec.Channel, testCase.channel)
		}
	}
}

func TestSubscriptionWithStartingCSV(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
		csv               string
	}{
		{
			csv:               "test",
			expectedErrorText: "",
		},
		{
			csv:               "",
			expectedErrorText: "can not redefine subscription with empty startingCSV",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestSubscriptionClientWithDummyObject()
		subscriptionBuilder := buildValidSubscriptionBuilder(testSettings).WithStartingCSV(testCase.csv)
		assert.Equal(t, subscriptionBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, subscriptionBuilder.Definition.Spec.StartingCSV, testCase.csv)
		}
	}
}

func TestSubscriptionWithInstallPlanApproval(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
		installPlan       oplmV1alpha1.Approval
	}{
		{
			installPlan:       oplmV1alpha1.ApprovalAutomatic,
			expectedErrorText: "",
		},
		{
			installPlan:       "",
			expectedErrorText: "Subscription 'installPlanApproval' must be either \"Automatic\" or \"Manual\"",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestSubscriptionClientWithDummyObject()
		subscriptionBuilder := buildValidSubscriptionBuilder(testSettings).WithInstallPlanApproval(testCase.installPlan)
		assert.Equal(t, subscriptionBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, subscriptionBuilder.Definition.Spec.InstallPlanApproval, testCase.installPlan)
		}
	}
}

func buildInValidSubscriptionBuilder(apiClient *clients.Settings) *SubscriptionBuilder {
	return NewSubscriptionBuilder(
		apiClient,
		"subscription",
		"",
		"test-source",
		"test-namespace",
		"package-test")
}

func buildValidSubscriptionBuilder(apiClient *clients.Settings) *SubscriptionBuilder {
	return NewSubscriptionBuilder(
		apiClient,
		"subscription",
		"test-namespace",
		"test-source",
		"test-namespace",
		"package-test")
}

func buildTestSubscriptionClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummySubscription(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummySubscription() []runtime.Object {
	return append([]runtime.Object{}, &oplmV1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subscription",
			Namespace: "test-namespace",
		},
		Spec: &oplmV1alpha1.SubscriptionSpec{
			Package: "test",
		},
	})
}
