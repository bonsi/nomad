package agent

import (
	"testing"

	"github.com/hashicorp/nomad/nomad/mock"
	"github.com/stretchr/testify/require"
)

func TestMockVariables(t *testing.T) {
	sv1 := mock.SecureVariable()
	mvs.Add(sv1.Path, *sv1)
	out := *mvs.List()
	require.NotNil(t, out)
	require.Len(t, out, 1)
}

func TestDeleteMockVariables(t *testing.T) {
	sv1 := mock.SecureVariable()
	mvs.Add(sv1.Path, *sv1)
	out := *mvs.List()
	require.NotNil(t, out)
	require.Len(t, out, 1)
	mvs.Delete(sv1.Path)
	require.Nil(t, mvs.List())
}
