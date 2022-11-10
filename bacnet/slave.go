package bacnet

type Slave struct {
	cfg SlaveConfig
}

func NewSlave(cfg SlaveConfig) (*Slave, error) {
	return &Slave{
		cfg: cfg,
	}, nil
}
