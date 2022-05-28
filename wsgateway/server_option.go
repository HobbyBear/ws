package wsgateway

import (
	"time"
	"ws/broker"
	"ws/pkg/logger"
)

type Option func(s *Server) *Server

func OptionSetLogger(l logger.Log) Option {
	return func(s *Server) *Server {
		s.Logger = l
		return s
	}
}

func OptionSetConnMgr(connMgr ConnMgr) Option {
	return func(s *Server) *Server {
		s.ConnMgr = connMgr
		return s
	}
}

func OptionSetBroker(broker broker.Broker) Option {
	return func(s *Server) *Server {
		s.Broker = broker
		return s
	}
}

func OptionSetMux(mux Handler) Option {
	return func(s *Server) *Server {
		s.Mux = mux
		return s
	}
}

func OptionSetTickOpen(open bool) Option {
	return func(s *Server) *Server {
		s.TickOpen = open
		return s
	}
}

func OptionSetCallOnStateChange(f func(c *Conn, state ConnState)) Option {
	return func(s *Server) *Server {
		s.CallConnStateChange = f
		return s
	}
}

func OptionSetTickExpireSec(sec int) Option {
	return func(s *Server) *Server {
		s.conTicker.TickExpireSec = sec
		return s
	}
}

func OptionSetWheelIntervalSec(sec int) Option {
	return func(s *Server) *Server {
		s.conTicker.WheelIntervalSec = sec
		return s
	}
}

func OptionSetUpgradeDeadline(t time.Duration) Option {
	return func(s *Server) *Server {
		s.UpgradeDeadline = t
		return s
	}
}

func OptionSetReadPayloadDeadline(t time.Duration) Option {
	return func(s *Server) *Server {
		s.ReadPayloadDeadline = t
		return s
	}
}

func OptionSetReadHeaderDeadline(t time.Duration) Option {
	return func(s *Server) *Server {
		s.ReadHeaderDeadline = t
		return s
	}
}
