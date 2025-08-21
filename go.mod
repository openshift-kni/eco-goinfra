module github.com/rh-ecosystem-edge/eco-goinfra

go 1.25

toolchain go1.25.0

require (
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/blang/semver/v4 v4.0.0
	github.com/containernetworking/cni v1.3.0
	github.com/go-openapi/errors v0.22.1
	github.com/go-openapi/strfmt v0.23.0
	github.com/go-openapi/swag v0.23.1
	github.com/go-openapi/validate v0.24.0
	github.com/golang/glog v1.2.5
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/vault/api v1.20.0
	github.com/hashicorp/vault/api/auth/approle v0.10.0
	github.com/hashicorp/vault/api/auth/kubernetes v0.10.0
	github.com/k8snetworkplumbingwg/multi-networkpolicy v1.0.1
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.7
	github.com/k8snetworkplumbingwg/sriov-network-operator v1.5.0
	github.com/kedacore/keda-olm-operator v0.0.0-20250709175755-d5e451d12d5d
	github.com/kedacore/keda/v2 v2.17.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20221122204822-d1a8c34382f1
	github.com/lib/pq v1.10.9
	github.com/metal3-io/baremetal-operator/apis v0.10.2
	github.com/nmstate/kubernetes-nmstate/api v0.0.0-20250711164732-0e728986112f
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/openshift-kni/cluster-group-upgrades-operator v0.0.0-20250715163214-56f0876892dc // release-4.19
	github.com/openshift-kni/lifecycle-agent v0.0.0-20250715161102-71395a52711a // release-4.19
	github.com/openshift-kni/numaresources-operator v0.4.18-0.2024100201.0.20250715062915-7cc48e4830bd // release-4.19
	github.com/openshift-kni/oran-o2ims/api/hardwaremanagement v0.0.0-20250728203300-9ac32da76e0e // 9ac32da76e0e93e1005f15a003aec72adbb5b3e6
	github.com/openshift-kni/oran-o2ims/api/provisioning v0.0.0-20250728203300-9ac32da76e0e // 9ac32da76e0e93e1005f15a003aec72adbb5b3e6
	github.com/openshift/api v0.0.0-20250529074221-97812373b6b4 // release-4.19
	github.com/openshift/client-go v0.0.0-20250425165505-5f55ff6979a1 // release-4.19
	github.com/openshift/cluster-nfd-operator v0.0.0-20250619073832-dbf21174e0c0 // release-4.19
	github.com/openshift/cluster-node-tuning-operator v0.0.0-20250408112936-4f58be155c79 // 4f58be155c79b2c92e7e1c36f481f7163b7f2497
	github.com/openshift/custom-resource-status v1.1.3-0.20220503160415-f2fdb4999d87
	github.com/openshift/elasticsearch-operator v0.0.0-20241202223819-cc1a232913d6 // release-5.8
	github.com/openshift/local-storage-operator v0.0.0-20250401053348-567d4745bb07 // release-4.19
	github.com/ovn-org/ovn-kubernetes/go-controller v0.0.0-20250716192743-2700eb06d1e8 // latest
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.82.2
	github.com/red-hat-storage/odf-operator v0.0.0-20250716125006-48092cb5468b // release-4.18
	github.com/sirupsen/logrus v1.9.3
	github.com/stmcginnis/gofish v0.20.0
	github.com/stretchr/testify v1.10.0
	github.com/thoas/go-funk v0.9.3
	golang.org/x/crypto v0.40.0
	golang.org/x/exp v0.0.0-20250711185948-6ae5c78190dc
	golang.org/x/net v0.42.0
	gopkg.in/k8snetworkplumbingwg/multus-cni.v4 v4.2.1
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/gorm v1.30.0
	k8s.io/api v0.32.6
	k8s.io/apiextensions-apiserver v0.32.6
	k8s.io/apimachinery v0.32.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubectl v0.32.6
	k8s.io/kubelet v0.32.6
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397
	maistra.io/api v0.0.0-20240319144440-ffa91c765143
	open-cluster-management.io/api v0.16.1
	open-cluster-management.io/governance-policy-propagator v0.16.0
	open-cluster-management.io/multicloud-operators-subscription v0.15.0
	sigs.k8s.io/container-object-storage-interface-api v0.1.0
	sigs.k8s.io/controller-runtime v0.20.4
	sigs.k8s.io/yaml v1.5.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/carapace-sh/carapace-shlex v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/clarketm/json v1.17.1 // indirect
	github.com/coreos/fcct v0.5.0 // indirect
	github.com/coreos/go-json v0.0.0-20230131223807-18775e0fb4fb // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/coreos/ign-converter v0.0.0-20230417193809-cee89ea7d8ff // indirect
	github.com/coreos/ignition v0.35.0 // indirect
	github.com/coreos/ignition/v2 v2.21.0 // indirect
	github.com/coreos/vcontext v0.0.0-20231102161604-685dc7299dc5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/expr-lang/expr v1.17.2 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/getkin/kin-openapi v0.127.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grafana/loki/operator/apis/loki v0.0.0-20241021105923-5e970e50b166
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/metal3-io/baremetal-operator/pkg/hardwareutils v0.5.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.4.1 // indirect
	github.com/oapi-codegen/runtime v1.1.1
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/openshift-kni/oran-o2ims/api/common v0.0.0-20250716190952-206606893d4d // indirect
	github.com/openshift/cluster-logging-operator/api/observability v0.0.0-20250422180113-5bae4ccfc5ef
	github.com/openshift/library-go v0.0.0-20250313122028-477d5d90df06 // indirect
	github.com/openshift/machine-config-operator v0.0.1-0.20250320230514-53e78f3692ee // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/r3labs/diff/v3 v3.0.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/speakeasy-api/openapi-overlay v0.9.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/vmware-tanzu/velero v1.15.2
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.mongodb.org/mongo-driver v1.17.3 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.32.6 // indirect
	k8s.io/cli-runtime v0.32.6 // indirect
	k8s.io/component-base v0.32.6 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-aggregator v0.32.3 // indirect
	k8s.io/kube-openapi v0.0.0-20250701173324-9bd5c66d9911 // indirect
	knative.dev/pkg v0.0.0-20250326102644-9f3e60a9244c // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/kube-storage-version-migrator v0.0.6-0.20230721195810-5c8923c5ff96 // indirect
	sigs.k8s.io/kustomize/api v0.20.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.20.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.7.0 // indirect
)

require github.com/go-test/deep v1.1.0 // indirect

replace (
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.16
	github.com/k8snetworkplumbingwg/sriov-network-operator => github.com/openshift/sriov-network-operator v0.0.0-20250625093820-3b2381406672 // release-4.18
	k8s.io/client-go => k8s.io/client-go v0.32.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.19.7
)

tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
