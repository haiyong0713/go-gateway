package like

// LiveActBusiness .
const LiveActBusiness = "live.activity_appointment"

// LiveMsg live follow room databus msg
type LiveMsg struct {
	Seid       int64  `json:"seid"`
	Businessid string `json:"businessid"`
	UID        int64  `json:"uid"`
	Status     int64  `json:"status"`
}
