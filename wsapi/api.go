package wsapi

import (
	"github.com/gobwas/ws"
	"ws/broker"
	"ws/msg"
)

type Client struct {
	broker.Producer
}

func (c *Client) SendMsgToAll(data []byte, opCode ws.OpCode) error {

	return c.Pub((&msg.PushMsg{
		Type:   msg.PushAll,
		Data:   data,
		WsType: opCode,
	}).Marshal())
}

func (c *Client) SendMsg(data []byte, opCode ws.OpCode, uids []string, groupId, topic string) error {
	t := msg.PushSingle
	if len(groupId) != 0 && len(uids) == 0 {
		t = msg.PushGroup
	}

	return c.Pub((&msg.PushMsg{
		Type:   t,
		Uids:   uids,
		RoomId: groupId,
		Topic:  topic,
		Data:   data,
		WsType: opCode,
	}).Marshal())
}
