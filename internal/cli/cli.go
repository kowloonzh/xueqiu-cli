package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kowloonzh/xueqiu-cli/xueqiu"
	"github.com/spf13/cobra"
)

type Config struct {
	ChromeRemoteURL string
}

const Version = "1.0.0"

type QuoteService interface {
	Quote(context.Context, string) (xueqiu.Quote, error)
	Quotes(context.Context, []string) ([]xueqiu.Quote, error)
}

func NewRootCommand(out io.Writer) *cobra.Command {
	return newRootCommand(out, func(config Config) QuoteService {
		provider := xueqiu.NewChromeCookieProvider(xueqiu.ChromeConfig{
			RemoteURL: config.ChromeRemoteURL,
			Timeout:   30 * time.Second,
		})
		return xueqiu.NewClient(http.DefaultClient, provider)
	})
}

func NewRootCommandWithService(service QuoteService, out io.Writer) *cobra.Command {
	return newRootCommand(out, func(Config) QuoteService {
		return service
	})
}

func newRootCommand(out io.Writer, serviceFactory func(Config) QuoteService) *cobra.Command {
	config := Config{}
	var service QuoteService

	getService := func() QuoteService {
		if service == nil {
			service = serviceFactory(config)
		}
		return service
	}

	cmd := &cobra.Command{
		Use:           "xueqiu-cli",
		Short:         "Query Xueqiu realtime stock quotes",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&config.ChromeRemoteURL, "chrome-remote-url", "", "Chrome remote debugging URL. Empty means local Chrome mode.")

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(out, Version)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stock <symbol>",
		Short: "Query one realtime quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbols, err := xueqiu.ParseSymbols(args[0])
			if err != nil {
				return err
			}
			if len(symbols) != 1 {
				return fmt.Errorf("stock expects one symbol")
			}

			quote, err := getService().Quote(cmd.Context(), symbols[0])
			if err != nil {
				return err
			}
			return writeJSON(out, quote)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stocks <symbol1,symbol2>",
		Short: "Query multiple realtime quotes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbols, err := xueqiu.ParseSymbols(args[0])
			if err != nil {
				return err
			}

			quotes, err := getService().Quotes(cmd.Context(), symbols)
			if err != nil {
				return err
			}
			return writeJSON(out, quotes)
		},
	})

	return cmd
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
