package opcua

// Config holds the OPC-UA part of the configuration
type Config struct {
	ServerURL string
	User      string `envconfig:"optional"`
	Password  string `envconfig:"optional"`
	CertFile  string `envconfig:"optional"`
	KeyFile   string `envconfig:"optional"`
}
