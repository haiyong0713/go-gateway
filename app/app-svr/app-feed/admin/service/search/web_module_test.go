package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

func TestService_WebModuleList(t *testing.T) {
	param := &show.SearchWebModuleLP{
		Pn:    1,
		Ps:    100,
		Query: "a",
	}
	values, err := s.SearchWebModuleList(context.Background(), param)
	if err != nil {
		panic(err)
	}
	bs, _ := json.Marshal(values)
	fmt.Println(string(bs))
}

func TestService_OpenSearchWebModule(t *testing.T) {
	param := &show.SearchWebModuleLP{
		Pn: 1,
		Ps: 100,
	}
	values, err := s.OpenSearchWebModule(context.Background(), param)
	if err != nil {
		panic(err)
	}
	bs, _ := json.Marshal(values)
	fmt.Println(string(bs))
}

func TestService_WebModuleAdd(t *testing.T) {
	param := &show.SearchWebModuleAP{
		Reason: "test",
		Module: "[{\"value\":\"5\"},{\"value\":\"1\"}]",
		Query:  "[{\"value\":\"test1\"},{\"value\":\"test2\"}]",
	}
	err := s.AddSearchWebModule(context.Background(), param)
	if err != nil {
		panic(err)
	}
}

func TestService_WebModuleUpdate(t *testing.T) {
	param := &show.SearchWebModuleUP{
		ID:     6,
		Reason: "test",
		Module: "[{\"id\":7,\"order\":2,\"value\":\"1\"},{\"value\":\"3\"}]",
		Query:  "[{\"id\":11,\"value\":\"test1\"},{\"value\":\"test2\"}]",
	}
	err := s.UpdateSearchWebModule(context.Background(), param)
	if err != nil {
		panic(err)
	}
}
