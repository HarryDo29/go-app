package initialize

import (
	"log"

	"go-app/global"
	"go-app/internal/websocket"
	"go-app/internal/wire"
)

func InitWebSocket() {
	// init websocket Hub (singleton - dùng chung toàn app)
	global.WsHub = websocket.NewHub()
	go global.WsHub.Run()

	// init message service for binding func into hub
	messageService, err := wire.InitMessageService()
	if err != nil {
		log.Fatalf("failed to initialize message service for websocket: %v", err)
	}

	global.WsHub.GetChannelMembersFunc = messageService.GetMemberIds
}
