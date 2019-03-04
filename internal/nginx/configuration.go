package nginx

type Item struct {
	Checksum string `json:"checksum"`
	Data     []byte `json:"data"`
}

type Configuration struct {
	Configuration *Item `json:"configuration,omitempty"`
	Certificates  *Item `json:"certificates,omitempty"`
	LUA           *Item `json:"lua,omitempty"`
}
