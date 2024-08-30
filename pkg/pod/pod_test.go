package pod

import (
	"bytes"
	"context"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func buildValidContainterBuilder() *ContainerBuilder {
	return NewContainerBuilder(
		"test-container",
		"registry.example.com/test/image:latest",
		[]string{"/bin/sh", "-c", "sleep infinity"})
}

func TestWithCustomResourcesRequests(t *testing.T) {
	testCases := []struct {
		testRequests   corev1.ResourceList
		expectedErrMsg string
		expectedValues map[string]string
	}{
		{
			testRequests:   corev1.ResourceList{},
			expectedErrMsg: "container's resource limit var 'resourceList' is empty",
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
			},
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
				corev1.ResourceName("openshift.io/fake2"):  resource.MustParse("2"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
				"openshift.io/fake2":  "2",
			},
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
				corev1.ResourceName("cpu"):                 resource.MustParse("2.5m"),
				corev1.ResourceName("memory"):              resource.MustParse("0.5Gi"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
				"cpu":                 "2.5m",
				"memory":              "0.5Gi",
			},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidContainterBuilder()
		testBuilder = testBuilder.WithCustomResourcesRequests(testCase.testRequests)

		if testCase.expectedErrMsg != "" {
			assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)
		} else {
			assert.Empty(t, testBuilder.errorMsg)
		}

		if len(testCase.testRequests) != 0 && len(testCase.expectedValues) != 0 {
			for k, v := range testCase.expectedValues {
				assert.Equal(t,
					resource.MustParse(v), testBuilder.definition.Resources.Requests[corev1.ResourceName(k)])
			}
		}
	}
}

func TestWaitUntilInStatuses(t *testing.T) {
	testCases := []struct {
		checkedPhases    []corev1.PodPhase
		expectedPodPhase corev1.PodPhase
		expectedErrMsg   string
		pod              *corev1.Pod
	}{
		{
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning},
			expectedPodPhase: "Running",
			expectedErrMsg:   "",
			pod:              generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodReady, false),
		},
		{
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: "Running",
			expectedErrMsg:   "",
			pod:              generateTestPod("test2", "ns1", corev1.PodRunning, corev1.PodReady, false),
		},
		{
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: "Succeeded",
			expectedErrMsg:   "",
			pod:              generateTestPod("test3", "ns2", corev1.PodSucceeded, corev1.PodReady, false),
		},
		{
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: "",
			expectedErrMsg:   "context deadline exceeded",
			pod:              generateTestPod("test1", "ns1", corev1.PodFailed, corev1.PodReady, false),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		runtimeObjects = append(runtimeObjects, testCase.pod)
		testBuilder, err := buildPodTestBuilderWithFakeObjects(runtimeObjects, testCase.pod.Name, testCase.pod.Namespace)
		assert.Nil(t, err)

		var phase *corev1.PodPhase
		phase, err = testBuilder.WaitUntilInOneOfStatuses(testCase.checkedPhases, 2*time.Second)

		assert.Equal(t, testCase.expectedErrMsg, getErrorString(err))
		assert.Equal(t, testCase.expectedPodPhase, *phase)
	}
}

func TestWaitUntilHealthy(t *testing.T) {
	testCases := []struct {
		includeSucceeded bool
		skipReadiness    bool
		ignoreFailedPods bool
		expectedErrMsg   string
		pod              *corev1.Pod
	}{
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "",
			pod:              generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodReady, false),
		},
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "context deadline exceeded",
			pod:              generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodInitialized, false),
		},
		{
			includeSucceeded: false,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "context deadline exceeded",
			pod:              generateTestPod("test1", "ns1", corev1.PodSucceeded, corev1.PodScheduled, false),
		},
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "",
			pod:              generateTestPod("test1", "ns1", corev1.PodSucceeded, corev1.PodScheduled, false),
		},
		{
			includeSucceeded: true,
			skipReadiness:    true,
			ignoreFailedPods: true,
			expectedErrMsg:   "",
			pod:              generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodScheduled, false),
		},
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: false,
			expectedErrMsg:   "context deadline exceeded",
			pod:              generateTestPod("test1", "ns1", corev1.PodFailed, corev1.PodScheduled, false),
		},
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "",
			pod:              generateTestPod("test1", "ns1", corev1.PodFailed, corev1.PodScheduled, true),
		},
		{
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedErrMsg:   "context deadline exceeded",
			pod:              generateTestPod("test1", "ns1", corev1.PodFailed, corev1.PodScheduled, false),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		runtimeObjects = append(runtimeObjects, testCase.pod)
		testBuilder, err := buildPodTestBuilderWithFakeObjects(runtimeObjects, testCase.pod.Name, testCase.pod.Namespace)
		assert.Nil(t, err)

		err = testBuilder.WaitUntilHealthy(2*time.Second, testCase.includeSucceeded, testCase.skipReadiness,
			testCase.ignoreFailedPods)

		assert.Equal(t, testCase.expectedErrMsg, getErrorString(err))
	}
}

func buildPodTestBuilderWithFakeObjects(objects []runtime.Object, name, namespace string) (*Builder, error) {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return Pull(&clients.Settings{
		K8sClient:       fakeClient,
		CoreV1Interface: fakeClient.CoreV1(),
	}, name, namespace)
}

func generateTestPod(name, namespace string, phase corev1.PodPhase, conditionType corev1.PodConditionType,
	neverRestart bool) *corev1.Pod {
	pod := corev1.Pod{}
	pod.Name = name
	pod.Namespace = namespace
	pod.Status.Phase = phase
	condition := corev1.PodCondition{}
	condition.Type = conditionType
	condition.Status = corev1.ConditionTrue

	pod.Status.Conditions = append(pod.Status.Conditions, condition)
	if neverRestart {
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	}

	return &pod
}

func getErrorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func TestBuilder_ExecCommand(t *testing.T) {
	type args struct {
		command       []string
		containerName []string
	}

	tests := []struct {
		name    string
		args    args
		pod     *corev1.Pod
		want    bytes.Buffer
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				command:       []string{"ls"},
				containerName: []string{"test"},
			},
			pod:     generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodReady, false),
			want:    *bytes.NewBufferString("Command fake output"),
			wantErr: false,
		},
	}
	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object
			runtimeObjects = append(runtimeObjects, testcase.pod)
			builder, err := buildPodTestBuilderWithFakeObjects(runtimeObjects, testcase.pod.Name, testcase.pod.Namespace)
			assert.Nil(t, err)

			NewSPDYExecutor = fakeNewSPDYExecutor
			GetRestURL = getFakeRestURL
			got, err := builder.ExecCommand(testcase.args.command, testcase.args.containerName...)

			if (err != nil) != testcase.wantErr {
				t.Errorf("Builder.ExecCommand() error = %v, wantErr %v", err, testcase.wantErr)

				return
			}

			if !reflect.DeepEqual(got, testcase.want) {
				t.Errorf("Builder.ExecCommand() = %v, want %v", got, testcase.want)
			}
		})
	}
}

func getFakeRestURL(builder *Builder, command []string, containerName ...string) (*url.URL, error) {
	if valid, err := builder.validate(); !valid {
		return &url.URL{}, err
	}

	return &url.URL{}, nil
}

var fakeNewSPDYExecutor = func(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	return &fakeExecutor{method: method, url: url}, nil
}

type fakeExecutor struct {
	method string
	url    *url.URL
}

func (f *fakeExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	if options.Stdout != nil {
		buf := new(bytes.Buffer)
		buf.WriteString("Command fake output")

		if _, err := options.Stdout.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (f *fakeExecutor) Stream(options remotecommand.StreamOptions) error {
	return nil
}
