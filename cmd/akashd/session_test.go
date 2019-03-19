package main

import (
	"os"
	"testing"

	"github.com/ovrclk/akash/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_RootDir_Env(t *testing.T) {
	testutil.WithTempDirEnv(t, "AKASHD_DATA", func(basedir string) {
		assertCommand(t, func(session Session, cmd *cobra.Command, args []string) error {
			assert.Equal(t, basedir, session.RootDir())
			return nil
		})
	})
}

func TestSession_RootDir_Arg(t *testing.T) {
	testutil.WithTempDir(t, func(basedir string) {
		assertCommand(t, func(session Session, cmd *cobra.Command, args []string) error {
			assert.Equal(t, basedir, session.RootDir())
			return nil
		}, "-d", basedir)
	})
}

func TestSession_EnvOverrides(t *testing.T) {
	testutil.WithTempDir(t, func(basedir string) {
		defer os.Unsetenv("AKASHD_DATA")
		defer os.Unsetenv("AKASHD_GENESIS")
		defer os.Unsetenv("AKASHD_VALIDATOR")
		defer os.Unsetenv("AKASHD_MONIKER")
		defer os.Unsetenv("AKASHD_P2P_SEEDS")

		gpath := "/foo/bar/genesis.json"
		vpath := "/foo/bar/private_validator.json"
		seeds := "a,b,c"
		moniker := "foobar"
		laddr := "tcp://0.0.0.0:25"

		os.Setenv("AKASHD_DATA", basedir)
		os.Setenv("AKASHD_GENESIS", gpath)
		os.Setenv("AKASHD_VALIDATOR", vpath)
		os.Setenv("AKASHD_MONIKER", moniker)
		os.Setenv("AKASHD_P2P_SEEDS", seeds)
		os.Setenv("AKASHD_RPC_LADDR", laddr)

		assertCommand(t, func(session Session, cmd *cobra.Command, args []string) error {
			cfg, err := session.TMConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, basedir, session.RootDir())
			assert.Equal(t, gpath, cfg.GenesisFile())
			assert.Equal(t, vpath, cfg.PrivValidatorKeyFile())
			assert.Equal(t, moniker, cfg.Moniker)
			assert.Equal(t, seeds, cfg.P2P.Seeds)
			assert.Equal(t, laddr, cfg.RPC.ListenAddress)
			return nil
		})
	})
}

func assertCommand(t *testing.T, fn sessionRunner, args ...string) {
	viper.Reset()

	ran := false

	base := baseCommand()

	cmd := &cobra.Command{
		Use: "test",
		RunE: withSession(func(session Session, cmd *cobra.Command, args []string) error {
			ran = true
			return fn(session, cmd, args)
		}),
	}

	base.AddCommand(cmd)
	base.SetArgs(append([]string{"test"}, args...))
	require.NoError(t, base.Execute())
	assert.True(t, ran)
}
