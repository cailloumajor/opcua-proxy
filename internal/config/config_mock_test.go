package config

type configInitializerMock struct {
	loadEnvFileError   error
	initEnvConfigError error
}

func (c configInitializerMock) loadEnvFile() error {
	return c.loadEnvFileError
}

func (c configInitializerMock) initEnvConfig(*Config) error {
	return c.initEnvConfigError
}

func MockInitializer(loadEnvFileError, initEnvConfigError error) {
	di = configInitializerMock{
		loadEnvFileError,
		initEnvConfigError,
	}
}
