package checks

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/na4ma4/config"
	"github.com/na4ma4/rsca/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Checks slice of checks with NextRun method for kicking clients to update.
type Checks []*Info

// NextRun sets the next run property of all checks to specified timestamp.
func (c Checks) NextRun(t time.Time) {
	for _, check := range c {
		check.NextRun = t
	}
}

// GetChecksFromViper gets all the checks from the viper.Viper config.
func GetChecksFromViper(cfg config.Conf, vcfg *viper.Viper, logger *zap.Logger, hostName string) Checks {
	checkListMap := make(map[string]bool)

	for _, key := range vcfg.AllKeys() {
		if strings.HasPrefix(key, "check.") {
			token := strings.SplitN(key, ".", 3)
			checkListMap[token[1]] = true
		}
	}

	checkList := make(Checks, len(checkListMap))
	i := 0
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for v := range checkListMap {
		checkList[i] = GetCheckFromViper(cfg, logger, v, hostName)
		logger.Info("adding check", zap.String("check.name", checkList[i].Name),
			zap.Duration("check.period", checkList[i].Period))
		i++
	}

	return checkList
}

// GetCheckFromViper returns a check with the specified name from the config file.
func GetCheckFromViper(cfg config.Conf, logger *zap.Logger, name, hostName string) *Info {
	check := &Info{
		Name:     cfg.GetString(fmt.Sprintf("check.%s.name", name)),
		Period:   cfg.GetDuration(fmt.Sprintf("check.%s.period", name)),
		Command:  cfg.GetString(fmt.Sprintf("check.%s.command", name)),
		Hostname: hostName,
	}

	switch cfg.GetString(fmt.Sprintf("check.%s.type", name)) {
	case "host":
		check.Type = api.CheckType_HOST
	case "", "service":
		check.Type = api.CheckType_SERVICE
	default:
		logger.Warn("unknown check type, defaulting to 'service'",
			zap.String("check", check.Name),
			zap.String("check-type-supplied", cfg.GetString(fmt.Sprintf("check.%s.type", name))),
		)

		check.Type = api.CheckType_SERVICE
	}

	return check
}
