package wsgateway

import "github.com/panjf2000/ants"

var (
	analyzeProtocolPool, _ = ants.NewPool(128)
	handlePool, _          = ants.NewPool(128)
	acceptPool, _          = ants.NewPool(128)
	writeHandlePool, _     = ants.NewPool(128)
)
