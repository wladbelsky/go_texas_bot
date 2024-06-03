package config

import (
	"flag"
	"os"
	"strconv"
)

var (
	token            = flag.String("token", "", "token for authorization")
	lavalinkHost     = flag.String("lavalink-host", "", "lavalink host")
	lavalinkPort     = flag.String("lavalink-port", "", "lavalink port")
	lavalinkPassword = flag.String("lavalink-password", "", "lavalink password")
)

func getFlagOrEnvString(flagName string, envName string) string {
	fl := flag.Lookup(flagName)
	if os.Getenv(envName) != "" {
		return os.Getenv(envName)
	}
	if fl != nil {
		return fl.Value.String()
	}
	return ""
}

func getFlagOrEnvInt(flagName string, envName string) int {
	fl := flag.Lookup(flagName)
	if os.Getenv(envName) != "" {
		i, _ := strconv.Atoi(os.Getenv(envName))
		return i
	} else if fl != nil {
		return fl.Value.(flag.Getter).Get().(int)
	}
	return 0
}

func Token() string {
	return getFlagOrEnvString("token", "BOT_TOKEN")
}

func LavalinkHost() string {
	return getFlagOrEnvString("lavalink-host", "LAVALINK_HOST")
}

func LavalinkPort() int {
	return getFlagOrEnvInt("lavalink-port", "LAVALINK_PORT")
}

func LavalinkPassword() string {
	return getFlagOrEnvString("lavalink-password", "LAVALINK_PASSWORD")
}
