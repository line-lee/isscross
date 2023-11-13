package example

import (
	"encoding/json"
	"log"
	"sunflower/common/models"
	"testing"
)

func TestMessageUnmarshal(t *testing.T) {
	str := "{\"mid\":\"d746e187-49aa-4494-bd60-64b1e94056ec\",\"types\":\"HeartbeatPublish\",\"mutex\":{}}"
	m := new(models.Message)
	json.Unmarshal([]byte(str), m)
	switch m.Types {

	case models.HeartbeatPublish:
		log.Printf("types找到HeartbeatPublish\n")
	default:
		log.Printf("types找不到\n")

	}

}
