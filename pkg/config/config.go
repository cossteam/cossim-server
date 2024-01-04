package config

type LogConfig struct {
	Stdout bool   `mapstructure:"stdout"`
	V      int    `mapstructure:"v"`
	Format string `mapstructure:"format"`
}

type MySQLConfig struct {
	Addr         string   `mapstructure:"addr"`
	DSN          string   `mapstructure:"dsn"`
	ReadDSN      []string `mapstructure:"readDSN"`
	IdleTimeout  string   `mapstructure:"idleTimeout"`
	QueryTimeout string   `mapstructure:"queryTimeout"`
	ExecTimeout  string   `mapstructure:"execTimeout"`
	TranTimeout  string   `mapstructure:"tranTimeout"`
}

type RedisConfig struct {
	Name         string `mapstructure:"name"`
	Proto        string `mapstructure:"proto"`
	Addr         string `mapstructure:"addr"`
	DialTimeout  string `mapstructure:"dialTimeout"`
	ReadTimeout  string `mapstructure:"readTimeout"`
	WriteTimeout string `mapstructure:"writeTimeout"`
	IdleTimeout  string `mapstructure:"idleTimeout"`
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
	Name    string `mapstructure:"name"`
	Addr    string `mapstructure:"addr"`
	Dial    string `mapstructure:"dial"`
	Timeout string `mapstructure:"timeout"`
}

type AppConfig struct {
	Log       LogConfig       `mapstructure:"log"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	HTTP      HTTPConfig      `mapstructure:"http"`
	GRPC      GRPCConfig      `mapstructure:"grpc"`
	Register  RegisterConfig  `mapstructure:"register"`
	Discovers DiscoversConfig `mapstructure:"discovers"`
}
