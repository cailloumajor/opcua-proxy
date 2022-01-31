package opcua

// Config holds the OPC-UA part of the configuration.
type Config struct {
	ServerURL string
	User      string
	Password  string
	CertFile  string
	KeyFile   string
}
