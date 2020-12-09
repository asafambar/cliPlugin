package commands

import (
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_extractAllArgsAndFlags(t *testing.T) {
	tcs := []struct {
		name             string
		arguments        []string
		expectedError    bool
		ExpectedResponse *ReleaseNotesConfiguration
	}{
		{
			name:      "product and version - valid",
			arguments: []string{"artifactory", "7.11.2"},
			ExpectedResponse: &ReleaseNotesConfiguration{
				Product: "artifactory",
				Version: "7.11.2",
				Current: false,
				Date:    false,
			},
		},
		{
			name:             "too many args - expect error",
			arguments:        []string{"artifactory", "7.11.2", "xray"},
			ExpectedResponse: nil,
			expectedError:    true,
		},
		{
			name:             "product only - expect error",
			arguments:        []string{"artifactory"},
			ExpectedResponse: nil,
			expectedError:    true,
		},
	}
	for _, tc := range tcs {
		cmpctx := &components.Context{
			Arguments: tc.arguments,
		}
		rc, err := extractAllArgsAndFlags(cmpctx)
		if tc.expectedError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		assert.Equal(t, tc.ExpectedResponse, rc)
	}

}
