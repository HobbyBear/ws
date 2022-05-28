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
				switch pushMsg.Type {
				case msg.PushSingle:
					if connlist := s.ConnMgr.GetConnByUid(pushMsg.Uid); len(connlist) != 0 {
						for _, conn := range connlist {
							// todo 线程池处理
							conn.WriteMsg(pushMsg.WsType, pushMsg.Data)
						}
					}

				case msg.PushGroup:
					if connlist := s.ConnMgr.GetConnByGroupId(pushMsg.GroupId); len(connlist) != 0 {
						for _, conn := range connlist {
							conn.WriteMsg(pushMsg.WsType, pushMsg.Data)
						}
					}
				case msg.PushAll:
					if connlist := s.ConnMgr.GetAllConn(); len(connlist) != 0 {
						for _, conn := range connlist {
							conn.WriteMsg(pushMsg.WsType, pushMsg.Data)
						}
					}
				}
			}
		}
	}()
}
