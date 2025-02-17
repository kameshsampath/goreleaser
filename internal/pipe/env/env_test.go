package env

import (
	"fmt"
	"os"
	"testing"

	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestSetDefaultTokenFiles(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		ctx := context.New(config.Project{})
		setDefaultTokenFiles(ctx)
		require.Equal(t, "~/.config/goreleaser/github_token", ctx.Config.EnvFiles.GitHubToken)
		require.Equal(t, "~/.config/goreleaser/gitlab_token", ctx.Config.EnvFiles.GitLabToken)
		require.Equal(t, "~/.config/goreleaser/gitea_token", ctx.Config.EnvFiles.GiteaToken)
	})
	t.Run("custom config config", func(t *testing.T) {
		cfg := "what"
		ctx := context.New(config.Project{
			EnvFiles: config.EnvFiles{
				GitHubToken: cfg,
			},
		})
		setDefaultTokenFiles(ctx)
		require.Equal(t, cfg, ctx.Config.EnvFiles.GitHubToken)
	})
	t.Run("templates", func(t *testing.T) {
		ctx := context.New(config.Project{
			ProjectName: "foobar",
			Env: []string{
				"FOO=FOO_{{ .Env.BAR }}",
				"FOOBAR={{.ProjectName}}",
				"EMPTY_VAL=",
			},
		})
		ctx.Env["FOOBAR"] = "old foobar"
		os.Setenv("BAR", "lebar")
		os.Setenv("GITHUB_TOKEN", "fake")
		require.NoError(t, Pipe{}.Run(ctx))
		require.Equal(t, "FOO_lebar", ctx.Env["FOO"])
		require.Equal(t, "foobar", ctx.Env["FOOBAR"])
		require.Equal(t, "", ctx.Env["EMPTY_VAL"])
	})

	t.Run("template error", func(t *testing.T) {
		ctx := context.New(config.Project{
			Env: []string{
				"FOO={{ .Asss }",
			},
		})
		require.EqualError(t, Pipe{}.Run(ctx), `template: tmpl:1: unexpected "}" in operand`)
	})

	t.Run("no token", func(t *testing.T) {
		ctx := context.New(config.Project{})
		require.NoError(t, Pipe{}.Run(ctx))
		require.Equal(t, ctx.TokenType, context.TokenTypeGitHub)
	})
}

func TestValidGithubEnv(t *testing.T) {
	require.NoError(t, os.Setenv("GITHUB_TOKEN", "asdf"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "asdf", ctx.Token)
	require.Equal(t, context.TokenTypeGitHub, ctx.TokenType)
	// so the tests do not depend on each other
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
}

func TestValidGitlabEnv(t *testing.T) {
	require.NoError(t, os.Setenv("GITLAB_TOKEN", "qwertz"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "qwertz", ctx.Token)
	require.Equal(t, context.TokenTypeGitLab, ctx.TokenType)
	// so the tests do not depend on each other
	require.NoError(t, os.Unsetenv("GITLAB_TOKEN"))
}

func TestValidGiteaEnv(t *testing.T) {
	require.NoError(t, os.Setenv("GITEA_TOKEN", "token"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "token", ctx.Token)
	require.Equal(t, context.TokenTypeGitea, ctx.TokenType)
	// so the tests do not depend on each other
	require.NoError(t, os.Unsetenv("GITEA_TOKEN"))
}

func TestInvalidEnv(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
	require.NoError(t, os.Unsetenv("GITLAB_TOKEN"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.Error(t, Pipe{}.Run(ctx))
	require.EqualError(t, Pipe{}.Run(ctx), ErrMissingToken.Error())
}

func TestMultipleEnvTokens(t *testing.T) {
	require.NoError(t, os.Setenv("GITHUB_TOKEN", "asdf"))
	require.NoError(t, os.Setenv("GITLAB_TOKEN", "qwertz"))
	require.NoError(t, os.Setenv("GITEA_TOKEN", "token"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.Error(t, Pipe{}.Run(ctx))
	require.EqualError(t, Pipe{}.Run(ctx), "multiple tokens found, but only one is allowed: GITHUB_TOKEN, GITLAB_TOKEN, GITEA_TOKEN\n\nLearn more at https://goreleaser.com/errors/multiple-tokens\n")
	// so the tests do not depend on each other
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
	require.NoError(t, os.Unsetenv("GITLAB_TOKEN"))
	require.NoError(t, os.Unsetenv("GITEA_TOKEN"))
}

func TestEmptyGithubFileEnv(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.Error(t, Pipe{}.Run(ctx))
}

func TestEmptyGitlabFileEnv(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITLAB_TOKEN"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.Error(t, Pipe{}.Run(ctx))
}

func TestEmptyGiteaFileEnv(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITEA_TOKEN"))
	ctx := &context.Context{
		Config: config.Project{},
	}
	require.Error(t, Pipe{}.Run(ctx))
}

func TestEmptyGithubEnvFile(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
	f, err := os.CreateTemp(t.TempDir(), "token")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, os.Chmod(f.Name(), 0o377))
	ctx := &context.Context{
		Config: config.Project{
			EnvFiles: config.EnvFiles{
				GitHubToken: f.Name(),
			},
		},
	}
	require.EqualError(t, Pipe{}.Run(ctx), fmt.Sprintf("failed to load github token: open %s: permission denied", f.Name()))
}

func TestEmptyGitlabEnvFile(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITLAB_TOKEN"))
	f, err := os.CreateTemp(t.TempDir(), "token")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, os.Chmod(f.Name(), 0o377))
	ctx := &context.Context{
		Config: config.Project{
			EnvFiles: config.EnvFiles{
				GitLabToken: f.Name(),
			},
		},
	}
	require.EqualError(t, Pipe{}.Run(ctx), fmt.Sprintf("failed to load gitlab token: open %s: permission denied", f.Name()))
}

func TestEmptyGiteaEnvFile(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITEA_TOKEN"))
	f, err := os.CreateTemp(t.TempDir(), "token")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, os.Chmod(f.Name(), 0o377))
	ctx := &context.Context{
		Config: config.Project{
			EnvFiles: config.EnvFiles{
				GiteaToken: f.Name(),
			},
		},
	}
	require.EqualError(t, Pipe{}.Run(ctx), fmt.Sprintf("failed to load gitea token: open %s: permission denied", f.Name()))
}

func TestInvalidEnvChecksSkipped(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
	ctx := &context.Context{
		Config:      config.Project{},
		SkipPublish: true,
	}
	require.NoError(t, Pipe{}.Run(ctx))
}

func TestInvalidEnvReleaseDisabled(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))

	t.Run("true", func(t *testing.T) {
		ctx := context.New(config.Project{
			Env: []string{},
			Release: config.Release{
				Disable: "true",
			},
		})
		require.NoError(t, Pipe{}.Run(ctx))
	})

	t.Run("tmpl true", func(t *testing.T) {
		ctx := context.New(config.Project{
			Env: []string{"FOO=true"},
			Release: config.Release{
				Disable: "{{ .Env.FOO }}",
			},
		})
		require.NoError(t, Pipe{}.Run(ctx))
	})

	t.Run("tmpl false", func(t *testing.T) {
		ctx := context.New(config.Project{
			Env: []string{"FOO=true"},
			Release: config.Release{
				Disable: "{{ .Env.FOO }}-nope",
			},
		})
		require.EqualError(t, Pipe{}.Run(ctx), ErrMissingToken.Error())
	})

	t.Run("tmpl error", func(t *testing.T) {
		ctx := context.New(config.Project{
			Release: config.Release{
				Disable: "{{ .Env.FOO }}",
			},
		})
		testlib.RequireTemplateError(t, Pipe{}.Run(ctx))
	})
}

func TestInvalidEnvReleaseDisabledTmpl(t *testing.T) {
	require.NoError(t, os.Unsetenv("GITHUB_TOKEN"))
}

func TestLoadEnv(t *testing.T) {
	t.Run("env exists", func(t *testing.T) {
		env := "SUPER_SECRET_ENV"
		require.NoError(t, os.Setenv(env, "1"))
		v, err := loadEnv(env, "nope")
		require.NoError(t, err)
		require.Equal(t, "1", v)
	})
	t.Run("env file exists", func(t *testing.T) {
		env := "SUPER_SECRET_ENV_NOPE"
		require.NoError(t, os.Unsetenv(env))
		f, err := os.CreateTemp(t.TempDir(), "token")
		require.NoError(t, err)
		fmt.Fprintf(f, "123")
		require.NoError(t, f.Close())
		v, err := loadEnv(env, f.Name())
		require.NoError(t, err)
		require.Equal(t, "123", v)
	})
	t.Run("env file with an empty line at the end", func(t *testing.T) {
		env := "SUPER_SECRET_ENV_NOPE"
		require.NoError(t, os.Unsetenv(env))
		f, err := os.CreateTemp(t.TempDir(), "token")
		require.NoError(t, err)
		fmt.Fprintf(f, "123\n")
		require.NoError(t, f.Close())
		v, err := loadEnv(env, f.Name())
		require.NoError(t, err)
		require.Equal(t, "123", v)
	})
	t.Run("env file is not readable", func(t *testing.T) {
		env := "SUPER_SECRET_ENV_NOPE"
		require.NoError(t, os.Unsetenv(env))
		f, err := os.CreateTemp(t.TempDir(), "token")
		require.NoError(t, err)
		fmt.Fprintf(f, "123")
		require.NoError(t, f.Close())
		err = os.Chmod(f.Name(), 0o377)
		require.NoError(t, err)
		v, err := loadEnv(env, f.Name())
		require.EqualError(t, err, fmt.Sprintf("open %s: permission denied", f.Name()))
		require.Equal(t, "", v)
	})
}
