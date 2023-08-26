package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestFlowCreator(t *testing.T) {
	filename := "../test/taskDSL1000.json"
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	var taskDSL TaskDSL
	err = json.Unmarshal(jsonData, &taskDSL)
	if err != nil {
		fmt.Println(err)
		return
	}

	serviceInterfaceInfo := &ServiceInterfaceInfo{
		FuncName:   "SayHello",
		InputType:  "HelloReq",
		OutputType: "HelloResp",
		HttpPath:   "/sayHello",
	}

	ctx := context.Background()
	build := &Builder{
		ServiceInterfaceInfo: serviceInterfaceInfo,
	}
	build.parseTaskDSL(ctx, &taskDSL)

	flowCreator, err := newFlowCreator(taskDSL.TaskList)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = flowCreator.run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
