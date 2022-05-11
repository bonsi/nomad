package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/nomad/nomad/mock"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/stretchr/testify/require"
)

func TestHTTP_SecureVariableList(t *testing.T) {
	//ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		// Test the empty list case
		req, err := http.NewRequest("GET", "/v1/vars", nil)
		require.NoError(t, err)
		respW := httptest.NewRecorder()

		// Make the request
		obj, err := s.Server.SecureVariablesRequest(respW, req)
		require.NoError(t, err)

		// add vars and test a populated backend
		sv1 := mock.SecureVariable()
		sv2 := mock.SecureVariable()
		sv3 := mock.SecureVariable()
		sv4 := mock.SecureVariable()
		sv4.Path = sv1.Path + "/child"
		for _, sv := range []*structs.SecureVariable{sv1, sv2, sv3, sv4} {
			args := structs.SecureVariableUpsertRequest{
				SecureVariable: sv,
				WriteRequest:   structs.WriteRequest{Region: "global"},
			}
			var resp structs.GenericResponse
			SV_Upsert(&args, &resp)
			// require.Nil(s.Agent.RPC("SecureVariable.UpsertSecureVariables", &args, &resp))
		}
		// Make the HTTP request
		req, err = http.NewRequest("GET", "/v1/vars", nil)
		require.NoError(t, err)
		respW = httptest.NewRecorder()

		// Make the request
		obj, err = s.Server.SecureVariablesRequest(respW, req)
		require.NoError(t, err)

		// Check for the index
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-Index"))
		require.Equal(t, "true", respW.HeaderMap.Get("X-Nomad-KnownLeader"))
		require.NotZero(t, respW.HeaderMap.Get("X-Nomad-LastContact"))

		// Check the output (the 3 we register )
		require.Len(t, obj.([]*structs.SecureVariableStub), 4)

		// test prefix query
		req, err = http.NewRequest("GET", "/v1/vars?prefix="+sv1.Path, nil)
		require.NoError(t, err)
		respW = httptest.NewRecorder()

		// Make the request
		obj, err = s.Server.SecureVariablesRequest(respW, req)
		require.NoError(t, err)
		require.Len(t, obj.([]*structs.SecureVariableStub), 2)

	})
}

func TestHTTP_SecureVariableQuery(t *testing.T) {
	//ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		// Make a request for a non-existent variable
		req, err := http.NewRequest("GET", "/v1/var/does/not/exist", nil)
		require.NoError(t, err)
		respW := httptest.NewRecorder()
		obj, err := s.Server.SecureVariableSpecificRequest(respW, req)
		require.EqualError(t, err, "Secure variable not found")

		// Don't pass a path
		req, err = http.NewRequest("GET", "/v1/var/", nil)
		require.NoError(t, err)
		respW = httptest.NewRecorder()
		obj, err = s.Server.SecureVariableSpecificRequest(respW, req)
		require.EqualError(t, err, "Missing secure variable path")

		// Use an incorrect verb
		req, err = http.NewRequest("LOLWUT", "/v1/var/foo", nil)
		require.NoError(t, err)
		respW = httptest.NewRecorder()
		obj, err = s.Server.SecureVariableSpecificRequest(respW, req)
		require.EqualError(t, err, ErrInvalidMethod)

		// Query a variable
		sv1 := mock.SecureVariable()
		args := structs.SecureVariableUpsertRequest{
			SecureVariable: sv1,
			WriteRequest:   structs.WriteRequest{Region: "global"},
		}
		var resp structs.GenericResponse
		SV_Upsert(&args, &resp)
		//require.Nil(s.Agent.RPC("SecureVariable.UpsertSecureVariables", &args, &resp))

		// Make the HTTP request
		req, err = http.NewRequest("GET", "/v1/var/"+sv1.Path, nil)
		require.NoError(t, err)
		respW = httptest.NewRecorder()

		// Make the request
		obj, err = s.Server.SecureVariableSpecificRequest(respW, req)
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
	//ci.Parallel(t)
	httpTest(t, nil, func(s *TestAgent) {
		sv1 := mock.SecureVariable()
		args := structs.SecureVariableUpsertRequest{
			SecureVariable: sv1,
			WriteRequest:   structs.WriteRequest{Region: "global"},
		}
		var resp structs.GenericResponse
		SV_Upsert(&args, &resp)

		// Make a change for update
		sv1U := sv1.Copy()
		sv1U.Items["newness"] = "awwyeah"

		// Make the HTTP request
		buf := encodeReq(&sv1U)
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
		require.NotEqual(t, sv1, out)
		require.Contains(t, out.Items, "newness")

		// break the request body
		badBuf := encodeBrokenReq(&sv1U)

		req, err = http.NewRequest("PUT", "/v1/var/"+sv1.Path, badBuf)
		require.NoError(t, err)
		respW = httptest.NewRecorder()

		// Make the request
		obj, err = s.Server.SecureVariableSpecificRequest(respW, req)
		require.EqualError(t, err, "unexpected EOF")
		var cErr HTTPCodedError
		require.ErrorAs(t, err, &cErr)
		require.Equal(t, http.StatusBadRequest, cErr.Code())
	})
}

func TestHTTP_SecureVariableUpdate(t *testing.T) {
	//ci.Parallel(t)
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
	//ci.Parallel(t)
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

func encodeBrokenReq(obj interface{}) io.ReadCloser {
	// var buf *bytes.Buffer
	// enc := json.NewEncoder(buf)
	// enc.Encode(obj)
	b, _ := json.Marshal(obj)
	b = b[0 : len(b)-5] // strip newline and final }
	return ioutil.NopCloser(bytes.NewReader(b))
}
