package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/na4ma4/config"
	"github.com/na4ma4/rsca/api"
	"github.com/na4ma4/rsca/internal/certs"
	"github.com/na4ma4/rsca/internal/common"
	"github.com/na4ma4/rsca/internal/mainconfig"
	"github.com/na4ma4/rsca/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

//nolint: gochecknoglobals // cobra uses globals in main
var rootCmd = &cobra.Command{
	Use: "rscad",
	Run: mainCommand,
}

//nolint:gochecknoinits // init is used in main for cobra
func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")
}

func main() {
	_ = rootCmd.Execute()
}

func mainCommand(cmd *cobra.Command, args []string) {
	cfg := config.NewViperConfigFromViper(viper.GetViper(), "rsca")

	logger, _ := cfg.ZapConfig().Build()
	defer logger.Sync() //nolint: errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lis, err := net.Listen("tcp", cfg.GetString("server.bind"))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	cp, err := certs.NewFileCertificateProvider(cfg.GetString("server.cert-dir"), true)
	if err != nil {
		logger.Fatal("failed to get certificates", zap.Error(err))
	}

	logger.Info("server listening", zap.String("bind", viper.GetString("server.bind")))

	hostName := getHostname(cfg)
	eg, ctx := errgroup.WithContext(ctx)
	sapi := server.NewServer(logger, hostName)
	gc := grpc.NewServer(cp.ServerOption())

	api.RegisterRSCAServer(gc, sapi)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	eg.Go(common.WaitForOSSignal(ctx, cancel, cfg, logger, c))
	eg.Go(sapi.(*server.Server).Run(ctx, cfg))
	eg.Go(common.ProcessWatchdog(ctx, cancel, cfg, logger))
	eg.Go(func() error { return gc.Serve(lis) })

	if cfg.GetBool("metrics.enabled") {
		go func() {
			http.Handle("/metrics", promhttp.Handler())

			if err := http.ListenAndServe(cfg.GetString("metrics.bind"), nil); err != nil {
				cancel()
			}
		}()
	}

	<-ctx.Done()
}

func getHostname(cfg config.Conf) string {
	hostName := cfg.GetString("general.hostname")
	if hostName == "" {
		hostName, _ = os.Hostname()
	}

	return hostName
}
