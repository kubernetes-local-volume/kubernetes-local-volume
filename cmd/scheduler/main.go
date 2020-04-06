package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/client-go/rest"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/injection"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/kubeconfig"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/signals"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/scheduler"
)

func main() {
	// kube config
	cfg := getKubeConfig()

	// context
	ctx := signals.NewContext()

	// logging
	logger := logging.FromContext(ctx)

	// injection
	ctx, informers := injection.Default.SetupInformers(ctx, cfg)

	// start informers
	logger.Info("Starting informers.")
	if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
		logger.Fatalw("Failed to start informers", err)
	}

	lvs := scheduler.NewLocalVolumeScheduler(ctx)

	router := httprouter.New()

	// add version route
	scheduler.AddVersion(router)

	// add predicate route
	scheduler.AddPredicate(router, lvs)

	// add prioritize route
	scheduler.AddPrioritize(router, lvs)

	// add bind route
	scheduler.AddBind(router, lvs)

	// add preemption route
	scheduler.AddPreemption(router, lvs)

	logger.Infof("local volume scheduler starting on the port :80")
	if err := http.ListenAndServe(":80", router); err != nil {
		logger.Fatal(err)
	}
}

func getKubeConfig() *rest.Config {
	var (
		masterURL = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
		config    = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	)
	flag.Parse()

	cfg, err := kubeconfig.GetConfig(*masterURL, *config)
	if err != nil {
		log.Fatal("Error building kubeconfig", err)
	}
	return cfg
}
