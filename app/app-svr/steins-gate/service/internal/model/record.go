package model

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	xecode "go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	_choiceSeparator        = ","
	RecWithoutCursor        = 1
	RecWithCursorInProgress = 2
	RecWithCursorPerfect    = 3
)

// NewRecordHandler 是生成新的存档的处理方法
type NewRecordHandler func(gameRec *api.GameRecords, reqID, rootID, cursor int64, portal int32) (inChoices, inCursorChoices string, fromID, currentCursor int64, recState int, err error)

// NewEdgeRecord 主要是，fromID逻辑取currentEdge
func NewEdgeRecord(gameRec *api.GameRecords, edgeID, firstEid, cursor int64, portal int32) (inChoices, inCursorChoices string, fromID, currentCursor int64, recState int, err error) {
	if gameRec != nil {
		fromID = gameRec.CurrentEdge
	}
	inChoices, inCursorChoices, currentCursor, recState, err = newRecord(gameRec, fromID, edgeID, firstEid, cursor, portal)
	return
}

// NewNodeRecord def.
func NewNodeRecord(gameRec *api.GameRecords, nodeID, firstNid, cursor int64, portal int32) (inChoices, inCursorChoices string, fromID, currentCursor int64, recState int, err error) {
	if gameRec != nil {
		fromID = gameRec.CurrentNode
	}
	inChoices, inCursorChoices, currentCursor, recState, err = newRecord(gameRec, fromID, nodeID, firstNid, cursor, portal)
	return
}

// PressHandler 类型即决定record往哪里填值
type PressHandler func(records *api.GameRecords, requestID int64)

// PressEdge 注意，edgeInfo写currentEdge
func PressEdge(rec *api.GameRecords, edgeID int64) {
	rec.CurrentEdge = edgeID
}

// PressNode 注意，nodeInfo写currentNode
func PressNode(rec *api.GameRecords, nodeID int64) {
	rec.CurrentNode = nodeID
}

// PressHandler 类型即决定record往哪里填值
type PullHandler func(records *api.GameRecords) (currentID int64)

// PressEdge 注意，edgeInfo写currentEdge
func PullEdge(rec *api.GameRecords) (currentID int64) {
	if rec != nil {
		currentID = rec.CurrentEdge
	}
	return
}

// PressNode 注意，nodeInfo写currentNode
func PullNode(rec *api.GameRecords) (currentID int64) {
	if rec != nil {
		currentID = rec.CurrentNode
	}
	return
}

// maxCursor从游标选择中获取最后一个游标，即为最大的游标 30,44,45,46 => 46
func maxCursor(cursorChoioces string) (lastCurInt int64) {
	idx := strings.LastIndex(cursorChoioces, _choiceSeparator)
	lastCursor := cursorChoioces[idx+1:]
	lastCurInt, _ = strconv.ParseInt(lastCursor, 10, 64)
	return
}

// fakeCursor用于客户端未传cursor的情况下，选择离尾部最近的cursor进行回溯
func fakeCursor(idList, cursorList []int64, currentID int64) (result int64) {
	for i := len(idList) - 1; i >= 0; i-- {
		if idList[i] == currentID {
			return cursorList[i]
		}
	}
	return
}

// initRec 空存档初始化逻辑
func initRec(requestID, rootID int64) (inChoices, inCursorChoices string, currentCursor int64) {
	if requestID != rootID { // 当前节点非根节点，则在存档开头增加根节点
		inChoices = fmt.Sprintf("%d,%d", rootID, requestID)
		inCursorChoices = fmt.Sprintf("%d,%d", 0, 1)
		currentCursor = 1
		return
	}
	inChoices = strconv.FormatInt(requestID, 10)
	inCursorChoices = strconv.FormatInt(0, 10)
	currentCursor = 0
	return
}

// newRecord 存档核心方法
func newRecord(gameRec *api.GameRecords, fromID, requestID, rootID, requestCursor int64, portal int32) (inChoices, inCursorChoices string, currentCursor int64, recState int, err error) { // 无存档第一次玩 或者用户进入详情页后再登陆请求详情页
	if gameRec == nil || portal == 2 { // 如果无存档/中途登陆
		inChoices, inCursorChoices, currentCursor = initRec(requestID, rootID)
		return
	}
	var (
		tmpChoices                   string
		nextCursor                   int64 // nextCursor用于获取fromCursor的下一个游标，用于判断是不是走过的回溯
		fromIndex                    int
		choiceList, cursorChoiceList []int64 // 对应的choice和cursor choice
	)
	if choiceList, err = xstr.SplitInts(gameRec.Choices); err != nil {
		err = xecode.GraphLoopRecordErr
		log.Error("Record %+v, Request RequestID %d Illegal!", gameRec, requestID)
		return
	}
	if gameRec.CursorChoice == "" {
		recState = RecWithoutCursor
	} else { // 有游标
		if cursorChoiceList, err = xstr.SplitInts(gameRec.CursorChoice); err != nil {
			err = xecode.GraphLoopRecordErr
			log.Error("Record %+v, Request RequestID %d Illegal!", gameRec, requestID)
			return
		}
		if len(cursorChoiceList) == len(choiceList) { // 如果长度一致，信任游标存档
			recState = RecWithCursorPerfect
		} else {
			recState = RecWithCursorInProgress
			log.Warn("NewRecordLenNotMatch %+v, Request RequestID %d, Len CursorChoice %d, Len Choice %d  Illegal!", gameRec, requestID, len(cursorChoiceList), len(choiceList))
		}
	}
	if recState == RecWithoutCursor || recState == RecWithCursorInProgress { // 无游标老存档或者有游标但是长度不一致 洗成一致
		tmpChoices, gameRec.CursorChoice, cursorChoiceList, gameRec.CurrentCursor, nextCursor, fromIndex = choiceWithoutCursor(choiceList, fromID)
	}
	if recState == RecWithCursorPerfect { // 游标存档完全体
		tmpChoices, nextCursor, fromIndex = choiceWithCursor(choiceList, cursorChoiceList, gameRec.CurrentCursor)
	}
	tmpChoices = fmt.Sprintf(",%s,", tmpChoices) // 前后加上逗号进行格式化
	if portal == 1 {                             // 纯回溯
		if requestCursor == IllegalCursor { // 前台没传游标和回溯场景下，需要赋值cursor和currentCursor
			requestCursor = fakeCursor(choiceList, cursorChoiceList, requestID)
		}
		if !strings.Contains(tmpChoices, fmt.Sprintf(",%d_%d,", requestID, requestCursor)) { // 如果是回溯，并且当前回溯的点从未到达过
			err = ecode.RequestErr
			log.Error("Record %+v, Request NodeID %d Illegal!", gameRec, requestID)
			return
		}
		return gameRec.Choices, gameRec.CursorChoice, requestCursor, recState, nil // 回溯仅改变当前游标
	}
	if portal == 0 && fromID != requestID { // 不是回溯操作且不是重播，说明是往前走
		if curIndex := strings.Index(tmpChoices, fmt.Sprintf(",%d_%d,%d_%d,", fromID, gameRec.CurrentCursor, requestID, nextCursor)); !isEnd(cursorChoiceList, gameRec.CurrentCursor) && curIndex != -1 { // 不是end或者曾经出现过，说明不改变存档，只影响cursor
			return gameRec.Choices, gameRec.CursorChoice, nextCursor, recState, nil // 如果是走过路径，仅改变当前游标
		}
		index := strings.Index(tmpChoices, fmt.Sprintf(",%d_%d,", fromID, gameRec.CurrentCursor)) // 说明是在往前走
		if index < 0 {
			log.Error("Record %+v, Last CurrentId %d, Illegal!", gameRec, fromID)
			err = ecode.RequestErr
			return
		}
		// todo 需要传参改为gameRec
		inChoices, inCursorChoices, err = filterChoices( // 进行最大200个id和游标计算
			gameRec.Choices, gameRec.CursorChoice,
			isEnd(cursorChoiceList, gameRec.CurrentCursor),
			fromIndex+1, requestID, maxCursor(gameRec.CursorChoice)+1)
		return inChoices, inCursorChoices, maxCursor(gameRec.CursorChoice) + 1, recState, err // 游标递增
	}
	log.Error("NewRecordLenNotMatch gameRec %+v, "+ // portal == 0 且 fromID和requestID相同的兜底逻辑 日志报警
		"portal %d, requestID %d, fromID %d",
		gameRec, portal, requestID, fromID)
	return gameRec.Choices, gameRec.CursorChoice, gameRec.CurrentCursor, recState, nil
}

// ClientCompatibility 客户端兼容逻辑，把node里的id替换成edgeID
func ClientCompatibility(node *api.GraphNode, edgeID int64) {
	node.Id = edgeID
}

func isEnd(inCursorChoices []int64, fromCursor int64) bool {
	return fromCursor == inCursorChoices[len(inCursorChoices)-1]
}

func choiceWithoutCursor(choiceList []int64, fromID int64) (tmpChoices, cursorChoice string, cursorChoiceList []int64, fromCursor, nextCursor int64, fromIndex int) {
	for i := 0; i < len(choiceList); i++ {
		tmpChoices += strconv.Itoa(int(choiceList[i])) + "_" + strconv.Itoa(i) + "," // 拼接出id_cursor用于判断是不是回溯
		cursorChoice += strconv.Itoa(i) + ","
		cursorChoiceList = append(cursorChoiceList, int64(i))
		if choiceList[i] == fromID { // 判断fromID的游标位置
			fromCursor = int64(i)
			fromIndex = i
			if i < len(choiceList)-1 { // 如果不是最后一个就赋值nextCursor
				nextCursor = int64(i + 1)
			}
		}
	}
	tmpChoices = trimEnd(tmpChoices)     // 去掉最后一个逗号
	cursorChoice = trimEnd(cursorChoice) // 去掉最后一个逗号
	return
}

func choiceWithCursor(choiceList, cursorChoiceList []int64, fromCursor int64) (tmpChoices string, nextCursor int64, fromIndex int) {
	for i := 0; i < len(choiceList); i++ {
		tmpChoices += strconv.Itoa(int(choiceList[i])) + "_" + strconv.Itoa(int(cursorChoiceList[i])) + "," // 拼接出id_cursor用于判断是不是回溯 		// 拿到当前游标值
		if fromCursor == cursorChoiceList[i] {                                                              // 如果当前游标为fromCursor，判断要么是最后一个，要么就赋值nextCursor
			fromIndex = i
			if i < len(choiceList)-1 {
				nextCursor = cursorChoiceList[i+1]
			}
		}
	}
	tmpChoices = trimEnd(tmpChoices) // 去掉最后一个逗号
	return
}

func filterChoices(inChoices, inCursorChoices string, isEnd bool, fromIndex int, requestID, currentCursor int64) (outChoices, outCursorChoices string, err error) {
	var choiceList, cursorChoiceList []int64
	if inChoices == "" || inCursorChoices == "" {
		err = ecode.RequestErr
		log.Error("filterChoices inChoices %s, inCursorChoices %s, Illegal!", inChoices, inCursorChoices)
		return
	}
	if choiceList, err = xstr.SplitInts(inChoices); err != nil {
		err = ecode.RequestErr
		log.Error("filterChoices inChoices %s, inCursorChoices %s, Illegal!", inChoices, inCursorChoices)
		return
	}
	if cursorChoiceList, err = xstr.SplitInts(inCursorChoices); err != nil {
		err = ecode.RequestErr
		log.Error("filterChoices inChoices %s, inCursorChoices %s, Illegal!", inChoices, inCursorChoices)
		return
	}
	if len(choiceList) != len(cursorChoiceList) { //  需要保证两个列表长度一致
		err = ecode.RequestErr
		log.Error("filterChoices inChoices %s, inCursorChoices %s, Illegal!", inChoices, inCursorChoices)
		return
	}
	if !isEnd {
		choiceList = append(choiceList[:fromIndex], requestID)
		cursorChoiceList = append(cursorChoiceList[:fromIndex], currentCursor)
	} else {
		choiceList = append(choiceList, requestID)
		cursorChoiceList = append(cursorChoiceList, currentCursor)
	}
	if len(choiceList) > MaxChoiceLen { //
		choiceList = append(choiceList[:1], choiceList[len(choiceList)-SecondBiggestChoice:]...)
		cursorChoiceList = append(cursorChoiceList[:1], cursorChoiceList[len(cursorChoiceList)-SecondBiggestChoice:]...)
	}
	outChoices = xstr.JoinInts(choiceList)
	outCursorChoices = xstr.JoinInts(cursorChoiceList)
	return
}

func trimEnd(in string) (out string) {
	out = strings.TrimRight(in, ",")
	return

}
