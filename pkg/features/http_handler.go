package features

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type flagType struct {
	Name      string
	ID        int64
	State     bool
	Immutable bool
}

type patchResponse struct {
	Applied bool
	Errors  []error
}

func (f *featuresImpl) FlagPatchByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ = json.Marshal(err)
		w.Write(res)
		return
	}

	var input []ChangeRequestByID
	err = json.Unmarshal(body, &input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ = json.Marshal(err)
		w.Write(res)
		return
	}

	var resp patchResponse
	for _, request := range input {
		if request.ID < 0 {
			resp.Errors = append(resp.Errors, fmt.Errorf("flag with id=%v is immutable (id < 0)", request.ID))
			continue
		}
		f.SetFlagByID(request.ID, request.Value)
	}
	if len(resp.Errors) < len(input) {
		resp.Applied = true
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	res, _ = json.Marshal(resp)
	w.Write(res)
	w.WriteHeader(http.StatusOK)
}

func (f *featuresImpl) FlagPatchByNameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var res []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ = json.Marshal(err)
		w.Write(res)
		return
	}

	var input []ChangeRequestByName
	err = json.Unmarshal(body, &input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ = json.Marshal(err)
		w.Write(res)
		return
	}

	var resp patchResponse
	for _, request := range input {
		ok := f.SetFlagByName(request.Name, request.Value)
		if !ok {
			resp.Errors = append(resp.Errors, fmt.Errorf("flag with name=%v is immutable", request.Name))
			continue
		}
	}
	if len(resp.Errors) < len(input) {
		resp.Applied = true
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	res, _ = json.Marshal(resp)
	w.Write(res)
	w.WriteHeader(http.StatusOK)
}

func (f *featuresImpl) FlagListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	res := make([]flagType, 0)
	f.RLock()
	for flag, id := range f.nameToID {
		res = append(res, flagType{
			Name:      flag,
			ID:        id,
			State:     f.state[id],
			Immutable: id < 0,
		})
	}
	f.RUnlock()

	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ = json.Marshal(err)
		w.Write(data)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}
