package routers

import (
	"go-app/internal/router/auth"
	"go-app/internal/router/channel"
	"go-app/internal/router/connection"
	"go-app/internal/router/group"
	"go-app/internal/router/message"
	"go-app/internal/router/rf"
	"go-app/internal/router/role"
	"go-app/internal/router/user"
	"go-app/internal/router/websocket"
	"go-app/internal/router/upload"
)

type RouterGroup struct {
	User       user.UserRouterGroup
	Auth       auth.AuthRouterGroup
	Rf         rf.RfRouterGroup
	Ws         websocket.WebsocketRouterGroup
	Role       role.RoleRouterGroup
	Channel    channel.ChannelRouterGroup
	Connection connection.ConnectionRouterGroup
	Group      group.GroupRouterGroup
	Message    message.MessageRouterGroup
	Upload     upload.UploadRouterGroup
}

var RouterGroupApp = new(RouterGroup)
