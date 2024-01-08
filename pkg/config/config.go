package config

type LogConfig struct {
	Stdout bool   `mapstructure:"stdout"`
	V      int    `mapstructure:"v"`
	Format string `mapstructure:"format"`
}

type MySQLConfig struct {
	Addr         string `mapstructure:"addr"`
	DSN          string `mapstructure:"dsn"`
	RootPassword string `mapstructure:"root_password"`
	Database     string `mapstructure:"database"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
}

type RedisConfig struct {
	Name  string `mapstructure:"name"`
	Proto string `mapstructure:"proto"`
	Addr  string `mapstructure:"addr"`
}

type HTTPConfig struct {
	Port string `mapstructure:"port"`
	Addr string `mapstructure:"addr"`
}

type GRPCConfig struct {
	Port string `mapstructure:"port"`
	Addr string `mapstructure:"addr"`
}

type RegisterConfig struct {
	Name string   `mapstructure:"name"`
	Addr string   `mapstructure:"addr"`
	Tags []string `mapstructure:"tags"`
}

type DiscoversConfig map[string]ServiceConfig

type ServiceConfig struct {
	Name string `mapstructure:"name"`
	Addr string `mapstructure:"addr"`
}

type AppConfig struct {
	Log        LogConfig        `mapstructure:"log"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	HTTP       HTTPConfig       `mapstructure:"http"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
	Register   RegisterConfig   `mapstructure:"register"`
	Discovers  DiscoversConfig  `mapstructure:"discovers"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

type EncryptionConfig struct {
	Enable     bool   `mapstructure:"enable"`
	Name       string `mapstructure:"name"`
	Email      string `mapstructure:"email"`
	RsaBits    int    `mapstructure:"rsaBits"`
	Passphrase string `mapstructure:"passphrase"`
}
