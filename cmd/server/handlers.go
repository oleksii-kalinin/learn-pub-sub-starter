package server

import (
	"fmt"

	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/gamelogic"
	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/pubsub"
	"github.com/oleksii-kalinin/learn-pub-sub-starter/internal/routing"
)

func handlerLog(gl routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(gl)
	if err != nil {
		return pubsub.NackRequeue
	}
	return pubsub.Ack
}
