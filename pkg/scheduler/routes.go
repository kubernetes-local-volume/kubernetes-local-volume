package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
)

const (
	versionPath    = "/version"
	apiPrefix      = "/scheduler"
	bindPath       = apiPrefix + "/bind"
	predicatesPath = apiPrefix + "/predicates/always_true"
	prioritiesPath = apiPrefix + "/priorities/zero_score"
	preemptionPath = apiPrefix + "/preemption"
)

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
}

func PredicateRoute(lvs *LocalVolumeScheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		logger := logging.FromContext(context.Background())
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		logger.Infof("local volume scheduler predicate extenderArgs = %s", buf.String())

		var extenderArgs schedulerapi.ExtenderArgs
		var extenderFilterResult *schedulerapi.ExtenderFilterResult

		if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
			extenderFilterResult = &schedulerapi.ExtenderFilterResult{
				Nodes:       nil,
				FailedNodes: nil,
				Error:       err.Error(),
			}
		} else {
			extenderFilterResult = lvs.PredicateHandler(extenderArgs)
		}

		if resultBody, err := json.Marshal(extenderFilterResult); err != nil {
			panic(err)
		} else {
			logger.Infof("local volume scheduler predicate extenderFilterResult = %s", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func PrioritizeRoute(lvs *LocalVolumeScheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		logger := logging.FromContext(context.Background())
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		logger.Infof("local volume scheduler prioritize extenderArgs = ", buf.String())

		var extenderArgs schedulerapi.ExtenderArgs
		var hostPriorityList *schedulerapi.HostPriorityList

		if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
			panic(err)
		}

		if list, err := lvs.PrioritizeHandler(extenderArgs); err != nil {
			panic(err)
		} else {
			hostPriorityList = list
		}

		if resultBody, err := json.Marshal(hostPriorityList); err != nil {
			panic(err)
		} else {
			logger.Infof("local volume scheduler prioritize hostPriorityList = ", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func BindRoute(lvs *LocalVolumeScheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		logger := logging.FromContext(context.Background())
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		logger.Infof("local volume scheduler bind extenderBindingArgs = ", buf.String())

		var extenderBindingArgs schedulerapi.ExtenderBindingArgs
		var extenderBindingResult *schedulerapi.ExtenderBindingResult

		if err := json.NewDecoder(body).Decode(&extenderBindingArgs); err != nil {
			extenderBindingResult = &schedulerapi.ExtenderBindingResult{
				Error: err.Error(),
			}
		} else {
			extenderBindingResult = lvs.BindHandler(extenderBindingArgs)
		}

		if resultBody, err := json.Marshal(extenderBindingResult); err != nil {
			panic(err)
		} else {
			logger.Infof("local volume scheduler extenderBindingResult = ", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func PreemptionRoute(lvs *LocalVolumeScheduler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		logger := logging.FromContext(context.Background())
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)
		logger.Infof("local volume scheduler preemption extenderPreemptionArgs = ", buf.String())

		var extenderPreemptionArgs schedulerapi.ExtenderPreemptionArgs
		var extenderPreemptionResult *schedulerapi.ExtenderPreemptionResult

		if err := json.NewDecoder(body).Decode(&extenderPreemptionArgs); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			extenderPreemptionResult = lvs.PreemptionHandler(extenderPreemptionArgs)
		}

		if resultBody, err := json.Marshal(extenderPreemptionResult); err != nil {
			panic(err)
		} else {
			logger.Infof("local volume scheduler extenderPreemptionResult = ", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

func VersionRoute(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, fmt.Sprint(types.Version))
}

func AddVersion(router *httprouter.Router) {
	router.GET(versionPath, DebugLogging(VersionRoute, versionPath))
}

func DebugLogging(h httprouter.Handle, path string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Print("debug: ", path, " request body = ", r.Body)
		h(w, r, p)
		log.Print("debug: ", path, " response=", w)
	}
}

func AddPredicate(router *httprouter.Router, lvs *LocalVolumeScheduler) {
	router.POST(predicatesPath, DebugLogging(PredicateRoute(lvs), predicatesPath))
}

func AddPrioritize(router *httprouter.Router, lvs *LocalVolumeScheduler) {
	router.POST(prioritiesPath, DebugLogging(PrioritizeRoute(lvs), prioritiesPath))
}

func AddBind(router *httprouter.Router, lvs *LocalVolumeScheduler) {
	if handle, _, _ := router.Lookup("POST", bindPath); handle != nil {
		log.Print("warning: AddBind was called more then once!")
	} else {
		router.POST(bindPath, DebugLogging(BindRoute(lvs), bindPath))
	}
}

func AddPreemption(router *httprouter.Router, lvs *LocalVolumeScheduler) {
	if handle, _, _ := router.Lookup("POST", preemptionPath); handle != nil {
		log.Print("warning: AddPreemption was called more then once!")
	} else {
		router.POST(preemptionPath, DebugLogging(PreemptionRoute(lvs), preemptionPath))
	}
}