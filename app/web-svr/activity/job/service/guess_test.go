package service

import (
	"fmt"
	"testing"
)

// go test -v action.go archive.go article.go articleday.go award.go bnj.go bws.go coin.go college.go college_archive.go college_bonus.go college_mid.go college_score.go college_socre_add.go contribution.go data.go dubbing.go dubbing_data.go ent.go faction.go gameholiday.go guess.go guess_test.go handwrite.go  image.go like.go like_stick_top.go live.go lottery.go mail.go match.go  native.go newstar.go pre.go question.go rank.go remix.go remix_data.go  reply.go reserve.go s10_canal.go s10_free_flow.go s10_parse.go service.go share_url.go stein.go subject.go task.go thumbup.go useraction.go wx_lottery.go
func TestGuessBiz(t *testing.T) {
	//eg := errgroup.WithContext(context.Background())
	//
	//eg.Go(func(ctx context.Context) error {
	//    fmt.Println("1.1。1")
	//    return errors.New("1.1。2")
	//})
	//eg.Go(func(ctx context.Context) error {
	//    fmt.Println("2.1")
	//    time.Sleep(time.Second * 5)
	//    fmt.Println("2.2")
	//    return nil
	//})
	//
	//err := eg.Wait()
	//fmt.Println("done", err)
	t.Run("test rebuild mid map", testRebuildMidMapBiz)
}

func testRebuildMidMapBiz(t *testing.T) {
	midList := []int64{1, 2, 3, 4, 12, 111}
	m := rebuildMidMap(midList)
	for k, v := range m {
		fmt.Println(k, v)
	}
}
