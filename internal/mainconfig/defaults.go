package mainconfig

import "github.com/spf13/viper"

// ConfigInit is the common config initialisation for the commands.
func ConfigInit() {
	viper.SetConfigName("rsca")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./artifacts")
	viper.AddConfigPath("./test")
	viper.AddConfigPath("$HOME/.rsca")
	viper.AddConfigPath("/etc/rsca")
	viper.AddConfigPath("/usr/local/rsca/etc")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("/run/secrets")
	viper.AddConfigPath("/etc/nsca")
	viper.AddConfigPath("/etc/nagios")
	viper.AddConfigPath(".")

	viper.SetDefault("general.jitter", "10s")
	viper.SetDefault("general.timeout", "3s")
	viper.SetDefault("general.retry", true)
	viper.SetDefault("general.max-retries", 3)
	viper.SetDefault("general.check-tick", "9s")
	viper.SetDefault("general.tags", []string{})

	viper.RegisterAlias("command.timeout", "general.timeout")

	viper.SetDefault("nagios.command-file", "/tmp/nagios.cmd")

	viper.SetDefault("client.server", "127.0.0.1:15888")
	viper.SetDefault("client.server-cert-type", false)

	viper.SetDefault("server.listen", "0.0.0.0:15888")
	viper.SetDefault("server.tick", "15s")
	viper.SetDefault("server.server-cert-type", true)

	viper.SetDefault("watchdog.enabled", false)
	viper.SetDefault("watchdog.tick", "30s")

	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.listen", "localhost:2112")

	_ = viper.ReadInConfig()
}