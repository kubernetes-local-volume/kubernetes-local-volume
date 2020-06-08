package agent

import (
	"context"
	"flag"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/apis/storage/v1alpha1"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/client"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/informers/storage/v1alpha1/localvolume"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/persistentvolume"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/lvm"
	internaltypes "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
	lvtypes "github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

var (
	nodeID = flag.String("nodeid", "", "node id")
)

func NewAgent(
	ctx context.Context,
) *controller.Impl {
	flag.Parse()
	logger := logging.FromContext(ctx)
	client := client.Get(ctx)
	lvInformer := localvolume.Get(ctx)
	pvInformer := persistentvolume.Get(ctx)

	// create vg
	_, err := lvm.CreateVG(lvtypes.VGName)
	if err != nil {
		logger.Fatalf("Create vg(%s) error = %s", lvtypes.VGName, err.Error())
	}

	r := &Reconciler{
		nodeID:     *nodeID,
		client:     client,
		lvInformer: lvInformer,
		lvLister:   lvInformer.Lister(),
		pvLister:   pvInformer.Lister(),
	}

	// register node local volume storage resource
	registerNodeLocalVolumeStorage(r)

	impl := controller.NewImpl(r, logger, ReconcilerName)

	lvInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	pvInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filter(*nodeID),
		Handler:    controller.HandleAll(impl.Enqueue),
	})

	logger.Info("Agent Started")
	return impl
}

func registerNodeLocalVolumeStorage(r *Reconciler) {
	logger := logging.GetLogger()

	_, err := r.client.LocalV1alpha1().LocalVolumes(v1.NamespaceDefault).Get(r.nodeID, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// register node local volume storage
		nlvs := &v1alpha1.LocalVolume{}
		nlvs.Name = r.nodeID
		_, err = r.client.LocalV1alpha1().LocalVolumes(v1.NamespaceDefault).Create(nlvs)
		if err == nil {
			logger.Infof("Register node local volume storage(%s) success", r.nodeID)
		}
	}
}

func filter(nodeID string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		pv, ok := obj.(*v1.PersistentVolume)
		if !ok {
			return false
		}

		return internaltypes.IsPVInMyNode(pv, nodeID)
	}
}
