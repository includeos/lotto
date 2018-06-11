package environment

type UplinkInfo struct {
	Name     string
	FileName string
	Tag      string
}

type EnvSettings struct {
	Vcloud Vcloud
	Fusion Fusion
}

type Environment interface {
	Name() string
	Create() error
	Delete() error
	GetUplinkInfo() (UplinkInfo, error)
	LaunchCmdOptions(string) []string
	RunClientCmd(cmd string) (string, error)
	RunClientCmdScript(file string) ([]byte, error)
}
