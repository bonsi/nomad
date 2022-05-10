package agent

import (
	"fmt"
	"sync"

	"github.com/hashicorp/nomad/nomad/structs"
)

var mvs MockVariableStore

type MockVariableStore struct {
	m          sync.RWMutex
	backingMap map[string]*structs.SecureVariable
}

func (mvs *MockVariableStore) List() *[]structs.SecureVariable {
	fmt.Println("***** List *****")
	mvs.m.Lock()
	if len(mvs.backingMap) == 0 {
		return nil
	}
	vars := make([]structs.SecureVariable, 0, len(mvs.backingMap))
	for _, sVar := range mvs.backingMap {
		outVar := sVar.Copy()
		vars = append(vars, outVar)
	}
	mvs.m.Unlock()
	return &vars
}
func (mvs *MockVariableStore) Add(p string, bag structs.SecureVariable) {
	fmt.Println("***** Add *****")
	mvs.m.Lock()
	nv := bag.Copy()
	mvs.backingMap[p] = &nv
	mvs.m.Unlock()
}

func (mvs *MockVariableStore) Get(p string) *structs.SecureVariable {
	fmt.Println("***** Get *****")
	var out structs.SecureVariable
	mvs.m.Lock()
	defer mvs.m.Unlock()

	if v, ok := mvs.backingMap[p]; ok {
		out = v.Copy()
	} else {
		return nil
	}
	return &out
}

// Delete removes a key from the store. Removing a non-existent key is a no-op
func (mvs *MockVariableStore) Delete(p string) {
	fmt.Println("***** Delete *****")
	mvs.m.Lock()
	delete(mvs.backingMap, p)
	mvs.m.Unlock()
}

// Delete removes a key from the store. Removing a non-existent key is a no-op
func (mvs *MockVariableStore) Reset() {
	fmt.Println("***** Reset *****")
	mvs.m.Lock()
	mvs.backingMap = make(map[string]*structs.SecureVariable)
	mvs.m.Unlock()
}

func init() {
	fmt.Println("***** Initializing mock variables backend *****")
	fmt.Println("***** Initializing mock variables backend *****")
	fmt.Println("***** Initializing mock variables backend *****")
	mvs.m.Lock()
	mvs.backingMap = make(map[string]*structs.SecureVariable)
	mvs.m.Unlock()
}

func SV_List(args *structs.SecureVariablesListRequest, out *structs.SecureVariablesListResponse) {
	var vars []*structs.SecureVariable
	vars = make([]*structs.SecureVariable, 0, len(mvs.backingMap))
	for _, sVar := range mvs.backingMap {
		outVar := sVar.Copy()
		vars = append(vars, &outVar)
	}
	out.SecureVariables = vars
	out.QueryMeta.KnownLeader = true
	out.QueryMeta.Index = 999
	out.QueryMeta.LastContact = 19
}

func SV_Upsert(args *structs.SecureVariableUpsertRequest, out *structs.GenericResponse) {
	nv := args.SecureVariable.Copy()
	mvs.Add(nv.Path, nv)
	out.WriteMeta.Index = 9999
}
func SV_Read(args *structs.SecureVariableGetRequest, out *structs.SingleSecureVariableResponse) {
	out.SecureVariable = mvs.Get(args.Path)
	out.Index = 9999
	out.QueryMeta.KnownLeader = true
	out.QueryMeta.Index = 999
	out.QueryMeta.LastContact = 19
}
func SV_Delete(args *structs.SecureVariableDeleteRequest, out *structs.GenericResponse) {
	mvs.Delete(args.Path)
	out.WriteMeta.Index = 9999
}
