package generator

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/heimdalr/dag"
	"strconv"
)

type FlowNode struct {
	NodeName     string
	TaskNodeList []string
	IsMainFlow   bool
}

type FlowCreator struct {
	Dag             *dag.DAG
	VertexTaskMap   map[string]string
	Queue           *list.List
	FlowNodeList    []*FlowNode
	FlowNodeCount   int
	TaskChildrenNum map[string]int
	TaskParentNum   map[string]int
}

func newFlowCreator(taskList []*TaskInfo) (f *FlowCreator, err error) {
	vertexTaskMap := make(map[string]string)
	d := dag.NewDAG()
	taskChildrenNum := make(map[string]int)
	taskParentNum := make(map[string]int)

	for _, task := range taskList {
		vertexTaskMap[task.Name], err = d.AddVertex(task.Name)
		if err != nil {
			return
		}
	}

	for _, task := range taskList {
		if task.InputList != nil {
			for _, taskInputInfo := range task.InputList {
				if taskInputInfo.PreTask == "" {
					continue
				}
				err = d.AddEdge(vertexTaskMap[taskInputInfo.PreTask], vertexTaskMap[task.Name])
				if err != nil {
					return
				}
				taskChildrenNum[taskInputInfo.PreTask] += 1
				taskParentNum[task.Name] += 1
			}
		}
	}

	fmt.Println(d)
	f = &FlowCreator{}
	f.Dag = d
	f.VertexTaskMap = vertexTaskMap
	f.Queue = list.New()
	f.TaskChildrenNum = taskChildrenNum
	f.TaskParentNum = taskParentNum
	return
}

func (p *FlowCreator) run() (err error) {
	err = p.createTaskTopologicalSort()
	if err != nil {
		return
	}

	return
}

func (p *FlowCreator) createTaskTopologicalSort() (err error) {
	queue := p.Queue

	var topologicalSort []*FlowNode
	err = p.pushRootNode()
	if err != nil {
		return
	}
	for {
		if queue.Len() == 0 {
			break
		}
		for {
			if queue.Len() == 0 {
				break
			}
			element := queue.Front()
			queue.Remove(element)
			flowNode := element.Value.(*FlowNode)

			topologicalSort = append(topologicalSort, flowNode)
		}

		err = p.pushRootNode()
		if err != nil {
			return
		}
	}
	p.FlowNodeList = topologicalSort

	jsonData, _ := json.Marshal(topologicalSort)
	fmt.Println("FlowNodeList: " + string(jsonData))
	return
}

func (p *FlowCreator) pushRootNode() (err error) {
	d := p.Dag
	queue := p.Queue
	vertexTaskMap := p.VertexTaskMap

	vertexIdMap := p.mapInterface2String(d.GetRoots())
	for _, taskName := range vertexIdMap {
		var serialTask []string
		serialTask, err = p.findSerialTask(taskName)
		if err != nil {
			return
		}

		//fmt.Println(serialTask)
		p.FlowNodeCount++
		flowNode := &FlowNode{NodeName: strconv.Itoa(p.FlowNodeCount)}
		for _, taskName = range serialTask {
			err = d.DeleteVertex(vertexTaskMap[taskName])
			if err != nil {
				return
			}

			flowNode.TaskNodeList = append(flowNode.TaskNodeList, taskName)
		}

		if len(vertexIdMap) == 1 {
			flowNode.IsMainFlow = true
		}
		queue.PushBack(flowNode)
	}
	return
}

func (p *FlowCreator) findSerialTask(taskName string) (serialTask []string, err error) {
	d := p.Dag
	vertexTaskMap := p.VertexTaskMap
	taskParentNum := p.TaskParentNum
	taskChildrenNum := p.TaskChildrenNum
	vertexId := vertexTaskMap[taskName]
	serialTask = append(serialTask, taskName)
	for {

		if taskChildrenNum[taskName] != 1 {
			return
		}

		var childrenMap map[string]interface{}
		childrenMap, err = d.GetChildren(vertexId)
		if err != nil {
			return
		}

		for childrenVertexId, taskNameInterface := range childrenMap {
			taskName = taskNameInterface.(string)
			if taskParentNum[taskName] != 1 {
				return
			}

			vertexId = childrenVertexId
			serialTask = append(serialTask, taskName)
		}
	}

}

func (p *FlowCreator) mapInterface2String(inputData map[string]interface{}) map[string]string {
	outputData := map[string]string{}
	for key, value := range inputData {
		outputData[key] = value.(string)
	}
	return outputData
}

func (p *FlowCreator) getFlowNodeList() (flowNodeList []*FlowNode, err error) {
	return p.FlowNodeList, nil
}
