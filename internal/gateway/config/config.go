package config

type Config struct {
	WSAddr        string `env:"WS_ADDR,notEmpty"`
	PrivateAddr   string `env:"PRIVATE_HTTP_ADDR,notEmpty"`
	MessengerURL  string `env:"MESSENGER_URL,notEmpty"`
	InternalToken string `env:"INTERNAL_TOKEN,notEmpty"`
}
