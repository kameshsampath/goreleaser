package teams

import (
	"testing"

	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestStringer(t *testing.T) {
	require.Equal(t, Pipe{}.String(), "teams")
}

func TestDefault(t *testing.T) {
	ctx := context.New(config.Project{})
	require.NoError(t, Pipe{}.Default(ctx))
	require.Equal(t, ctx.Config.Announce.Teams.MessageTemplate, defaultMessageTemplate)
}

func TestAnnounceInvalidTemplate(t *testing.T) {
	ctx := context.New(config.Project{
		Announce: config.Announce{
			Teams: config.Teams{
				Enabled:         true,
				MessageTemplate: "{{ .Foo }",
			},
		},
	})
	require.EqualError(t, Pipe{}.Announce(ctx), `teams: template: tmpl:1: unexpected "}" in operand`)
}

func TestAnnounceMissingEnv(t *testing.T) {
	ctx := context.New(config.Project{
		Announce: config.Announce{
			Teams: config.Teams{
				Enabled: true,
			},
		},
	})
	require.NoError(t, Pipe{}.Default(ctx))
	require.EqualError(t, Pipe{}.Announce(ctx), `teams: env: environment variable "TEAMS_WEBHOOK" should not be empty`)
}

func TestSkip(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		require.True(t, Pipe{}.Skip(context.New(config.Project{})))
	})

	t.Run("dont skip", func(t *testing.T) {
		ctx := context.New(config.Project{
			Announce: config.Announce{
				Teams: config.Teams{
					Enabled: true,
				},
			},
		})
		require.False(t, Pipe{}.Skip(ctx))
	})
}
