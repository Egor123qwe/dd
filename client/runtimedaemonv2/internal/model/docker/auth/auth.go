package auth

type Settings struct {
	// auth enabled when credentials is not nil & [[]Port > 0]
	Credentials *Credentials
}

type Credentials struct {
	Login    string
	Password string
}

type Port struct {
	InPort  string
	OutPort string
}

type State struct {
	Enabled bool
	Ports   []Port
}
