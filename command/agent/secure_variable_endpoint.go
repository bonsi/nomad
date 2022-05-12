package agent

import (
	"net/http"
	"strings"

	"github.com/hashicorp/nomad/nomad/structs"
)

func (s *HTTPServer) SecureVariablesListRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "GET" {
		return nil, CodedError(405, ErrInvalidMethod)
	}

	args := structs.SecureVariablesListRequest{}
	if s.parse(resp, req, &args.Region, &args.QueryOptions) {
		return nil, nil
	}

	var out structs.SecureVariablesListResponse
	if err := s.agent.RPC("SecureVariables.List", &args, &out); err != nil {
		return nil, err
	}

	setMeta(resp, &out.QueryMeta)

	if out.Data == nil {
		out.Data = make([]*structs.SecureVariableStub, 0)
	}
	return out.Data, nil
}

func (s *HTTPServer) SecureVariableSpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	name := strings.TrimPrefix(req.URL.Path, "/v1/var/")
	if len(name) == 0 {
		return nil, CodedError(400, "Missing secure variable path")
	}
	switch req.Method {
	case "GET":
		return s.SecureVariableQuery(resp, req, name)
	case "PUT", "POST":
		return s.SecureVariableUpsert(resp, req, name)
	case "DELETE":
		return s.SecureVariableDelete(resp, req, name)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (s *HTTPServer) SecureVariableQuery(resp http.ResponseWriter, req *http.Request,
	SecureVariablePath string) (interface{}, error) {
	args := structs.SecureVariablesReadRequest{
		Path: SecureVariablePath,
	}
	if s.parse(resp, req, &args.Region, &args.QueryOptions) {
		return nil, nil
	}

	var out structs.SecureVariablesReadResponse
	if err := s.agent.RPC("SecureVariables.Read", &args, &out); err != nil {
		return nil, err
	}

	setMeta(resp, &out.QueryMeta)

	if out.Data == nil {
		return nil, CodedError(404, "Secure variable not found")
	}
	return out.Data, nil
}

func (s *HTTPServer) SecureVariableUpsert(resp http.ResponseWriter, req *http.Request,
	SecureVariablePath string) (interface{}, error) {
	// Parse the SecureVariable
	var SecureVariable structs.SecureVariable
	if err := decodeBody(req, &SecureVariable); err != nil {
		return nil, CodedError(http.StatusBadRequest, err.Error())
	}
	SecureVariable.Path = SecureVariablePath
	// Format the request
	args := structs.SecureVariablesUpsertRequest{
		Data: &SecureVariable,
	}
	s.parseWriteRequest(req, &args.WriteRequest)

	var out structs.SecureVariablesUpsertResponse
	if err := s.agent.RPC("SecureVariables.Update", &args, &out); err != nil {
		return nil, err
	}
	setIndex(resp, out.WriteMeta.Index)

	return nil, nil
}

func (s *HTTPServer) SecureVariableDelete(resp http.ResponseWriter, req *http.Request,
	SecureVariablePath string) (interface{}, error) {

	args := structs.SecureVariablesDeleteRequest{
		Path: SecureVariablePath,
	}
	s.parseWriteRequest(req, &args.WriteRequest)

	var out structs.SecureVariablesDeleteResponse
	if err := s.agent.RPC("SecureVariables.Delete", &args, &out); err != nil {
		return nil, err
	}
	setIndex(resp, out.WriteMeta.Index)
	return nil, nil
}
