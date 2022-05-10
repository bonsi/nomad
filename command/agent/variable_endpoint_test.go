package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/nomad/ci"
	"github.com/hashicorp/nomad/nomad/mock"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/stretchr/testify/require"
)

func TestHTTP_SecureVariableList(t *testing.T) {
	ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		sv1 := mock.SecureVariable()
		sv2 := mock.SecureVariable()
		sv3 := mock.SecureVariable()
		for _, sv := range []*structs.SecureVariable{sv1, sv2, sv3} {
			args := structs.SecureVariableUpsertRequest{
				SecureVariable: sv,
				WriteRequest:   structs.WriteRequest{Region: "global"},
			}
			var resp structs.GenericResponse
			SV_Upsert(&args, &resp)
			// require.Nil(s.Agent.RPC("SecureVariable.UpsertSecureVariables", &args, &resp))
		}
		// Make the HTTP request
		req, err := http.NewRequest("GET", "/v1/vars", nil)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariablesRequest(respW, req)
		require.NoError(t, err)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))
		require.Equal(t, "true", respW.HeaderMap.Get("X-Nomad-KnownLeader"))
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-LastContact"))

		// Check the output (the 3 we register )
		require.Len(t, obj.([]*structs.SecureVariable), 3)
	})
}

func TestHTTP_SecureVariableQuery(t *testing.T) {
	ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		sv1 := mock.SecureVariable()
		args := structs.SecureVariableUpsertRequest{
			SecureVariable: sv1,
			WriteRequest:   structs.WriteRequest{Region: "global"},
		}
		var resp structs.GenericResponse
		SV_Upsert(&args, &resp)
		//require.Nil(s.Agent.RPC("SecureVariable.UpsertSecureVariables", &args, &resp))

		// Make the HTTP request
		req, err := http.NewRequest("GET", "/v1/var/"+sv1.Path, nil)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariableSpecificRequest(respW, req)
		require.NoError(t, err)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))
		require.Equal(t, "true", respW.HeaderMap.Get("X-Nomad-KnownLeader"))
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-LastContact"))

		// Check the output
		require.Equal(t, sv1.Path, obj.(*structs.SecureVariable).Path)
	})
}

func TestHTTP_SecureVariableCreate(t *testing.T) {
	ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		// Make the HTTP request
		sv1 := *mock.SecureVariable()
		buf := encodeReq(sv1)
		req, err := http.NewRequest("PUT", "/v1/var/"+sv1.Path, buf)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariableSpecificRequest(respW, req)
		require.NoError(t, err)
		require.Nil(t, obj)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))

		// Check policy was created
		out := *mvs.Get(sv1.Path)
		require.NotNil(t, out)

		sv1.CreateIndex, sv1.ModifyIndex = out.CreateIndex, out.ModifyIndex
		require.Equal(t, sv1.Path, out.Path)
		require.Equal(t, sv1, out)
	})
}

func TestHTTP_SecureVariableUpdate(t *testing.T) {
	ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		// Make the HTTP request
		sv1 := *mock.SecureVariable()
		buf := encodeReq(sv1)
		req, err := http.NewRequest("PUT", "/v1/var/"+sv1.Path, buf)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariableSpecificRequest(respW, req)
		require.NoError(t, err)
		require.Nil(t, obj)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))

		// Check policy was created
		out := *mvs.Get(sv1.Path)
		require.NotNil(t, out)

		sv1.CreateIndex, sv1.ModifyIndex = out.CreateIndex, out.ModifyIndex
		require.Equal(t, sv1.Path, out.Path)
		require.Equal(t, sv1, out)
	})
}

func TestHTTP_SecureVariableDelete(t *testing.T) {
	ci.Parallel(t)
	defer mvs.Reset()
	httpTest(t, nil, func(s *TestAgent) {
		sv1 := mock.SecureVariable()
		args := structs.SecureVariableUpsertRequest{
			SecureVariable: sv1,
			WriteRequest:   structs.WriteRequest{Region: "global"},
		}
		var resp structs.GenericResponse
		SV_Upsert(&args, &resp)
		// require.Nil(s.Agent.RPC("SecureVariable.UpsertSecureVariables", &args, &resp))

		// Make the HTTP request
		req, err := http.NewRequest("DELETE", "/v1/var/"+sv1.Path, nil)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariableSpecificRequest(respW, req)
		require.NoError(t, err)
		require.Nil(t, obj)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))

		// Check variable bag was deleted
		out := mvs.Get(sv1.Path)
		require.Nil(t, out)
	})
}
