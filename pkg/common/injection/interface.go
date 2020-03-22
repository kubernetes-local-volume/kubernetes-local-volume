package injection

import (
	"context"
	"sync"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/controller"

	"k8s.io/client-go/rest"
)

// Interface is the interface for interacting with injection
// implementations, such as our Default and Fake below.
type Interface interface {
	// RegisterClient registers a new injector callback for associating
	// a new client with a context.
	RegisterClient(ClientInjector)

	// GetClients fetches all of the registered client injectors.
	GetClients() []ClientInjector

	// RegisterInformerFactory registers a new injector callback for associating
	// a new informer factory with a context.
	RegisterInformerFactory(InformerFactoryInjector)

	// GetInformerFactories fetches all of the registered informer factory injectors.
	GetInformerFactories() []InformerFactoryInjector

	// RegisterDuck registers a new duck.InformerFactory for a particular type.
	RegisterDuck(ii DuckFactoryInjector)

	// GetDucks accesses the set of registered ducks.
	GetDucks() []DuckFactoryInjector

	// RegisterInformer registers a new injector callback for associating
	// a new informer with a context.
	RegisterInformer(InformerInjector)

	// GetInformers fetches all of the registered informer injectors.
	GetInformers() []InformerInjector

	// SetupInformers runs all of the injectors against a context, starting with
	// the clients and the given rest.Config.  The resulting context is returned
	// along with a list of the .Informer() for each of the injected informers,
	// which is suitable for passing to controller.StartInformers().
	// This does not setup or start any controllers.
	SetupInformers(context.Context, *rest.Config) (context.Context, []controller.Informer)
}

var (
	// Check that impl implements Interface
	_ Interface = (*impl)(nil)

	// Default is the injection interface with which informers should register
	// to make themselves available to the controller process when reconcilers
	// are being run for real.
	Default Interface = &impl{}

	// Fake is the injection interface with which informers should register
	// to make themselves available to the controller process when it is being
	// unit tested.
	Fake Interface = &impl{}
)

type impl struct {
	m sync.RWMutex

	clients   []ClientInjector
	factories []InformerFactoryInjector
	informers []InformerInjector
	ducks     []DuckFactoryInjector
}
