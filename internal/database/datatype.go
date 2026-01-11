package database

type DataType string

const (
	DefaultDataType DataType = "data"
	PlayerDataType  DataType = "player"
	PartyDataType   DataType = "party"
	ProxyDataType   DataType = "proxy"
	BackendDataType DataType = "backend"
)

func (dt DataType) String() string {
	return string(dt)
}
