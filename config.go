package helloworld

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	DiscordPublicKey string `env:"DISCORD_PUBLIC_KEY" env-required:"true"`
	PetpetAPIBaseURL string `env:"PETPET_API_BASE_URL" env-required:"true"`
	MaxImageSize     int64  `env:"MAX_IMAGE_SIZE" env-default:"1048576"` // in bytes. default 1 MiB
}

func LoadConfig() Config {
	var cfg Config
	cleanenv.ReadEnv(&cfg)
	return cfg
}
