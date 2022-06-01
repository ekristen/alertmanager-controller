//go:generate go run github.com/ibuildthecloud/baaah/cmd/deepcopy ./pkg/apis/promam.ekristen.dev/v1/
//#go:generate go run k8s.io/kube-openapi/cmd/openapi-gen -i github.com/ekristen/prom-am-operator/pkg/apis/promam.ekristen.dev/v1,github.com/ekristen/prom-am-operator/pkg/apis/promam.ekristen.dev/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/version,k8s.io/apimachinery/pkg/api/resource,k8s.io/api/core/v1 -p ./pkg/openapi/generated -h tools/header.txt

package main
