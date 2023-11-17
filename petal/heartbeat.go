package petal

import (
	"encoding/json"
	"github.com/google/uuid"
	"sunflower/common/models"
	"time"
)

func heartbeatPush() {
	for {
		time.Sleep(20 * time.Second)
		bytes, _ := json.Marshal(models.Message{Mid: uuid.NewString(), Types: models.HeartbeatPublish})
		write(thisConnect, models.HeartbeatPublish, bytes)
	}
}
