package http

import (
	"bytes"
	"encoding/csv"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/ticket"
	"time"
)

func ticketCreate(c *bm.Context) {
	req := new(ticket.ReqTicketCreate)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(actSrv.TicketCreate(c, req))
}

func ticketExport(c *bm.Context) {
	tickets, err := actSrv.TicketExport(c)
	if err != nil {
		c.JSON(nil, err)
	} else {
		b := &bytes.Buffer{}
		b.WriteString("\xEF\xBB\xBF")
		wr := csv.NewWriter(b)
		wr.Write([]string{"id", "门票码", "状态(1:已签到,0:未签到)", "创建时间", "最后更新时间/签到时间"})
		for _, ti := range tickets {
			wr.Write([]string{
				fmt.Sprint(ti.ID),
				ti.Ticket,
				fmt.Sprint(ti.State),
				ti.Ctime.Time().Format("2006-01-02 15:04:05"),
				ti.Mtime.Time().Format("2006-01-02 15:04:05"),
			})
		}
		wr.Flush()
		c.Writer.Header().Set("Content-Type", "text/csv")
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=electronic_tickets.%s.csv", time.Now().Format("20060102")))
		tet := b.String()
		c.String(200, tet)
	}
}
