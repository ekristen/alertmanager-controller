package scheme

import (
	v1 "github.com/ekristen/prom-am-operator/pkg/apis/promam.ekristen.dev/v1"
	"github.com/rancher/wrangler/pkg/merr"
	"github.com/rancher/wrangler/pkg/schemes"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	Scheme         = runtime.NewScheme()
	Codecs         = serializer.NewCodecFactory(Scheme)
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

func AddToScheme(scheme *runtime.Scheme) error {
	var errs []error
	errs = append(errs, v1.AddToScheme(scheme))
	errs = append(errs, corev1.AddToScheme(scheme))
	errs = append(errs, appsv1.AddToScheme(scheme))
	errs = append(errs, batchv1.AddToScheme(scheme))
	errs = append(errs, networkingv1.AddToScheme(scheme))
	errs = append(errs, storagev1.AddToScheme(scheme))
	return merr.NewErrors(errs...)
}

func init() {
	utilruntime.Must(schemes.Register(AddToScheme))
	utilruntime.Must(AddToScheme(Scheme))
}
