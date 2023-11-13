package sdk

import (
	"encoding/json"
	"github.com/google/uuid"
	"sunflower/common/models"
	"time"
)

func heartbeatPush() {
	for {
		time.Sleep(10 * time.Second)
		bytes, _ := json.Marshal(models.Message{Mid: uuid.NewString(), Types: models.HeartbeatPublish})
		write(thisConnect, bytes)
	}
}
