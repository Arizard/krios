package entity

import (
	"arieoldman/arieoldman/krios/common"
)

// Config is the entity which describes an SDN configuration.
type Config struct {
	DPIDs []common.EthAddr
	L2Switching bool
	DPIEnabled bool
}
