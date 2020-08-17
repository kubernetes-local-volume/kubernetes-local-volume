package scheduler

import (
	"math/rand"
	"time"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/apis/extender/v1"
)

func (lvs *LocalVolumeScheduler) PrioritizeHandler(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	return lvs.prioritize(*args.Pod, args.Nodes.Items)
}

func (lvs *LocalVolumeScheduler) prioritize(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	logger := logging.FromContext(lvs.ctx)
	requestSize := lvs.getPodLocalVolumeRequestSize(&pod)

	var priorityList schedulerapi.HostPriorityList
	priorityList = make([]schedulerapi.HostPriority, len(nodes))
	for i, node := range nodes {
		freeSize := lvs.getNodeFreeSize(node.Name)
		logger.Infof("local volume scheduler handle pod(%s, namespace = %s) requestsize(%d) prioritize: node(%s) free size(%d)",
			pod.Namespace, pod.Name, requestSize, node.Name, freeSize)

		priorityList[i] = schedulerapi.HostPriority{
			Host: node.Name,
		}

		if requestSize == 0 && freeSize == 0 {
			priorityList[i].Score = randInt64Range(1, 10)

		} else if requestSize == 0 && freeSize > 0 {
			priorityList[i].Score = randInt64Range(1, 5)

		} else if freeSize > requestSize {
			priorityList[i].Score = getScoreByNodeLocalVolumeSize(int64(freeSize))

		} else {
			priorityList[i].Score = 0
		}
	}

	return &priorityList, nil
}

func getScoreByNodeLocalVolumeSize(localvolumeSize int64) int64 {
	score := localvolumeSize % 10

	if score == 0 {
		score = 10
	}
	return score
}

func randInt64Range(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}
