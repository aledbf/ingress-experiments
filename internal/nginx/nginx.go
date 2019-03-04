package nginx

type NGINX struct{}

func (n *NGINX) Start() {}

func (n *NGINX) Stop() {}

func (n *NGINX) Reload() {}

func (n *NGINX) Update(cfg *Configuration) error {
	return nil
}
