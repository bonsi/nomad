package structs

import "time"

type SecureVariable struct {
	Namespace   string            `json:"Namespace,omitempty"`
	Path        string            `json:"Path,omitempty"`
	CreateIndex int64             `json:"CreateIndex,omitempty"`
	CreateTime  time.Time         `json:"CreateTime,omitempty"`
	ModifyIndex int64             `json:"ModifyIndex,omitempty"`
	ModifyTime  time.Time         `json:"ModifyTime,omitempty"`
	Items       map[string]string `json:"Items,omitempty"`
	Meta        map[string]string `json:"Meta,omitempty"`
	Version     int64             `json:"version,omitempty"`
}

// SecureVariablesListRequest is used to request a list of namespaces
type SecureVariablesListRequest struct {
	QueryOptions
}

// SecureVariablesListResponse is used for a list request
type SecureVariablesListResponse struct {
	SecureVariables []*SecureVariable
	QueryMeta
}

// SecureVariableSpecificRequest is used to fetch a specific
// secure variable bag by path
type SecureVariableSpecificRequest struct {
	Path string
	QueryOptions
}

// SingleNamespaceResponse is used to return a single variable bag
type SingleSecureVariableResponse struct {
	SecureVariable *SecureVariable
	QueryMeta
}

// SingleSecureVariableRequest is used to put a single variable bag
// at a path
type SecureVariableGetRequest struct {
	Path string
	QueryOptions
}

// SecureVariableDeleteRequest is used to delete a set of secure variable
// bags at a given set of paths.
type SecureVariableDeleteRequest struct {
	Path string
	WriteRequest
}

// NamespaceUpsertRequest is used to upsert a set of namespaces
type SecureVariableUpsertRequest struct {
	SecureVariable *SecureVariable
	WriteRequest
}

func (sv SecureVariable) Copy() SecureVariable {
	out := SecureVariable{
		Namespace:   sv.Namespace,
		Path:        sv.Path,
		CreateIndex: sv.CreateIndex,
		CreateTime:  sv.CreateTime,
		ModifyIndex: sv.ModifyIndex,
		ModifyTime:  sv.ModifyTime,
		Items:       make(map[string]string, len(sv.Items)),
		Meta:        make(map[string]string, len(sv.Meta)),
		Version:     sv.Version,
	}
	for k, v := range sv.Items {
		out.Items[k] = v
	}
	for k, v := range sv.Meta {
		out.Meta[k] = v
	}
	return out
}
