package itest

import (
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-plugin-template/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestXrayWithVersion(t *testing.T) {
	conf := &commands.ReleaseNotesConfiguration{
		Product: "xray",
		Version: "3.11.2",
		Current: false,
		Date:    false,
	}
	rnResp, err := commands.DoGetReleaseNotes(&components.Context{}, conf)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(response3112Xray), strings.TrimSpace(rnResp))
}

func TestXrayWithVersionDateOnly(t *testing.T) {
	conf := &commands.ReleaseNotesConfiguration{
		Product: "xray",
		Version: "3.11.2",
		Current: false,
		Date:    true,
	}
	rnResp, err := commands.DoGetReleaseNotes(&components.Context{}, conf)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(response3112XrayDate), strings.TrimSpace(rnResp))
}

func TestArtiWithVersion(t *testing.T) {
	conf := &commands.ReleaseNotesConfiguration{
		Product: "artifactory",
		Version: "7.9.2",
		Current: false,
		Date:    false,
	}
	rnResp, err := commands.DoGetReleaseNotes(&components.Context{}, conf)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(response7106Arti), strings.TrimSpace(rnResp))
}

func TestArtiWithVersionDateOnly(t *testing.T) {
	conf := &commands.ReleaseNotesConfiguration{
		Product: "artifactory",
		Version: "7.9.2",
		Current: false,
		Date:    true,
	}
	rnResp, err := commands.DoGetReleaseNotes(&components.Context{}, conf)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(response7106ArtiDate), strings.TrimSpace(rnResp))
}

var response3112Xray = `### Xray 3.11.2

Released: November 11, 2020
#### Resolved Issues
1. Fixed an issue whereby, when a call to an Xray endpoint that requires authentication is done with bad credentials, consecutive API calls, even with good credentials, might fail as well.
2. Fixed an issue whereby, duplicate update Metadata server events were created causing redundant load on internal systems like RabbitMQ, PostgreSQL and MDS.
3. Fixed an issue whereby, lack of data sanitation sometimes led to SQL injection.
`
var response3112XrayDate = `Released: November 11, 2020`

var response7106Arti = `### Artifactory 7.9.2

Released: 20 October, 2020

#### Resolved Issues
1. Fixed an issue occurring in Artifactory version 7.9, whereby when installing or upgrading a JFrog Artifactory HA environment, the HA nodes sometimes failed to start due to a bad hex format for the join key.
2. Fixed an issue, whereby missing dependencies caused RPM installs to fail on certain operating systems.
`
var response7106ArtiDate = `Released: 20 October, 2020`
