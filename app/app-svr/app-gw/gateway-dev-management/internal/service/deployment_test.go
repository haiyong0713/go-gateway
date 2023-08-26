package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

func Test_Resource(t *testing.T) {
	var resource *model.CalcResource
	str := "{\"mem_req\":2048,\"cpu_req\":100,\"estorage_req\":0,\"mem_limit\":4096,\"cpu_limit\":2000,\"estorage_limit\":0,\"gpu_req\":0,\"gpu_limit\":0}"
	err := json.Unmarshal([]byte(str), &resource)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resource)
}
