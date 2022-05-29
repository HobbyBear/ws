package wsgateway

import (
	"encoding/json"
	"ws/msg"
)

func initConsumer(s *Server) {
	if s.Broker == nil {
		s.Logger.Errorf("init consumer fail broker is nil")
		return
	}
	go func() {
		ch := s.Broker.Sub()
		for true {
			select {
			case chanMsg := <-ch:
				var pushMsg msg.PushMsg
				json.Unmarshal(chanMsg, &pushMsg)
				s.Logger.Infof("收到sub 消息%+v data=%s", pushMsg, string(pushMsg.Data))
				switch pushMsg.Type {
				case msg.PushSingle:
					if connlist := s.ConnMgr.GetConnByUids(pushMsg.Uids); len(connlist) != 0 {
						for _, conns := range connlist {
							for _, conn := range conns {
								c, mt, data := conn, pushMsg.WsType, pushMsg.Data
								writeHandlePool.Submit(func() {
									c.WriteMsg(mt, data)
								})
							}
						}
					}

				case msg.PushGroup:
					if connlist := s.ConnMgr.GetConnByRoomId(pushMsg.RoomId); len(connlist) != 0 {
						for _, conn := range connlist {
							c, mt, data := conn, pushMsg.WsType, pushMsg.Data
							writeHandlePool.Submit(func() {
								c.WriteMsg(mt, data)
							})
						}
					}
				case msg.PushAll:
					if connlist := s.ConnMgr.GetAllConn(); len(connlist) != 0 {
						for _, conn := range connlist {
							c, mt, data := conn, pushMsg.WsType, pushMsg.Data
							writeHandlePool.Submit(func() {
								c.WriteMsg(mt, data)
							})
						}
					}
				}
			}
		}
	}()
}
