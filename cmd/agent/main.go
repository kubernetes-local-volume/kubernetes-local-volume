package main

import (
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/agent"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/sharemain"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/gc"
)

func main() {
	sharemain.Main(
		agent.NewAgent,
		gc.NewGC,
	)
}
