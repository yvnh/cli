package resolver

import (
	"errors"
	"os"

	"github.com/cli/cli/internal/config"
	"github.com/spf13/cobra"
)

type Strat func() (string, error)

type Resolver struct {
	strats map[string][]Strat
}

func New() *Resolver {
	return &Resolver{
		strats: make(map[string][]Strat),
	}
}

func (r *Resolver) AddStrat(name string, strat Strat) {
	r.strats[name] = append(r.strats[name], strat)
}

func (r *Resolver) GetStrats(name string) []Strat {
	return r.strats[name]
}

func (r *Resolver) ResolverFunc(name string) Strat {
	return ResolverFunc(r.GetStrats(name)...)
}

func (r *Resolver) Resolve(name string) (string, error) {
	return ResolverFunc(r.GetStrats(name)...)()
}

func ResolverFunc(strategies ...Strat) func() (string, error) {
	return func() (string, error) {
		for _, strat := range strategies {
			out, err := strat()
			if err != nil {
				return "", err
			}

			if out != "" {
				return out, nil
			}
		}

		return "", errors.New("No strategies resulted in output")
	}
}

func EnvStrat(name string) Strat {
	return func() (string, error) {
		return os.Getenv(name), nil
	}
}

func ConfigStrat(config func() (config.Config, error), hostname, name string) Strat {
	return func() (string, error) {
		cfg, err := config()
		if err != nil {
			return "", err
		}

		return cfg.Get(hostname, name)
	}
}

func ArgsStrat(args []string, index int) Strat {
	return func() (string, error) {
		if len(args) > index {
			return args[index], nil
		}
		return "", nil
	}
}

func StringFlagStrat(cmd *cobra.Command, name string) Strat {
	return func() (string, error) {
		return cmd.Flags().GetString(name)
	}
}

func MutuallyExclusiveBoolFlagsStrat(cmd *cobra.Command, names ...string) Strat {
	return func() (string, error) {
		flags := cmd.Flags()
		enabledFlagCount := 0
		enabledFlag := ""
		for _, name := range names {
			val, err := flags.GetBool(name)
			if err != nil {
				return "", err
			}

			if val {
				enabledFlagCount = enabledFlagCount + 1
				enabledFlag = name
			}

			if enabledFlagCount > 1 {
				break
			}
		}

		if enabledFlagCount == 0 {
			return "", nil
		} else if enabledFlagCount == 1 {
			return enabledFlag, nil
		}

		return "", errors.New("expected exactly one of boolean flags to be true")
	}
}

func ValueStrat(name string) Strat {
	return func() (string, error) {
		return name, nil
	}
}
