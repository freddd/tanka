package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/go-clix/cli"
	"github.com/spf13/pflag"

	"github.com/grafana/tanka/pkg/tanka"
)

func evalCmd() *cli.Command {
	cmd := &cli.Command{
		Short: "evaluate the jsonnet to json",
		Use:   "eval <path>",
		Args:  workflowArgs,
	}

	getExtCode, getTLACode := cliCodeParser(cmd.Flags())

	cmd.Run = func(cmd *cli.Command, args []string) error {
		raw, err := tanka.Eval(args[0],
			tanka.WithExtCode(getExtCode()),
			tanka.WithTLACode(getTLACode()),
		)

		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return err
		}

		pageln(string(out))
		return nil
	}

	return cmd
}

func cliCodeParser(fs *pflag.FlagSet) (func() map[string]string, func() map[string]string) {
	// need to use StringArray instead of StringSlice, because pflag attempts to
	// parse StringSlice using the csv parser, which breaks when passing objects
	extCode := fs.StringArray("ext-code", nil, "Set code value of extVar (Format: key=<code>)")
	extStr := fs.StringArrayP("ext-str", "V", nil, "Set string value of extVar (Format: key=value)")

	tlaCode := fs.StringArray("tla-code", nil, "Set code value of top level function (Format: key=<code>)")
	tlaStr := fs.StringArrayP("tla-str", "A", nil, "Set string value of top level function (Format: key=value)")

	newParser := func(kind string, code, str *[]string) func() map[string]string {
		return func() map[string]string {
			m := make(map[string]string)
			for _, s := range *code {
				split := strings.SplitN(s, "=", 2)
				if len(split) != 2 {
					log.Fatalf(kind+"-code argument has wrong format: `%s`. Expected `key=<code>`", s)
				}
				m[split[0]] = split[1]
			}

			for _, s := range *str {
				split := strings.SplitN(s, "=", 2)
				if len(split) != 2 {
					log.Fatalf(kind+"-str argument has wrong format: `%s`. Expected `key=<value>`", s)
				}
				m[split[0]] = fmt.Sprintf(`"%s"`, split[1])
			}
			return m
		}
	}

	return newParser("ext", extCode, extStr),
		newParser("tla", tlaCode, tlaStr)
}
