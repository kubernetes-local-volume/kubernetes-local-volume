package agent

import (
	"context"
	"flag"
	"math"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/apis/storage/v1alpha1"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/client"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/injection/informers/storage/v1alpha1/nodelocalvolumestorage"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/client/kube/injection/informers/core/v1/persistentvolume"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/lvm"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
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
	nlvsInformer := nodelocalvolumestorage.Get(ctx)
	pvInformer := persistentvolume.Get(ctx)

	// create vg
	_, err := lvm.CreateVG(types.VGName)
	if err != nil {
		logger.Fatalf("Create vg(%s) error = %s", types.VGName, err.Error())
	}

	r := &Reconciler{
		nodeID:       *nodeID,
		client:       client,
		nlvsInformer: nlvsInformer,
		nlvsLister:   nlvsInformer.Lister(),
		pvLister:     pvInformer.Lister(),
	}

	// register node local volume storage resource
	registerNodeLocalVolumeStorage(r)

	impl := controller.NewImpl(r, logger, ReconcilerName)

	nlvsInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	pvInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filter(*nodeID),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	logger.Info("Agent Started")
	return impl
}

func registerNodeLocalVolumeStorage(r *Reconciler) {
	logger := logging.GetLogger()

	_, err := r.client.LocalV1alpha1().NodeLocalVolumeStorages(v1.NamespaceDefault).Get(r.nodeID, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		return
	} else if err != nil {
		logger.Fatalf("Register get node local volume storage(%s) error = %s", r.nodeID, err.Error())
	}

	// register node local volume storage
	totalSize, _ := lvm.VGTotalSize(types.VGName)
	freeSize, _ := lvm.VGFreeSize(types.VGName)
	nlvs := &v1alpha1.NodeLocalVolumeStorage{}
	nlvs.Name = r.nodeID
	nlvs.Status.TotalSize = uint64(math.Floor(float64(totalSize) / 1024))
	nlvs.Status.UsedSize = uint64(math.Floor(float64(totalSize-freeSize) / 1024))
	nlvs.Status.PreAllocated = make(map[string]string)
	_, err = r.client.LocalV1alpha1().NodeLocalVolumeStorages(v1.NamespaceDefault).Create(nlvs)
	if err != nil {
		logger.Fatalf("Register create node local volume storage(%s) error = %s", r.nodeID, err.Error())
	} else {
		logger.Infof("Register node local volume storage(%s) success", r.nodeID)
	}
}

func filter(nodeID string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		pv, ok := obj.(v1.PersistentVolume)
		if !ok {
			return false
		}

		return isPVInMyNode(&pv, nodeID)
	}
}
