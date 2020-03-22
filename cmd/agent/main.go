package main

import (
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/agent"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/sharemain"
)

func main() {
	sharemain.Main(
		agent.NewAgent,
	)
}
