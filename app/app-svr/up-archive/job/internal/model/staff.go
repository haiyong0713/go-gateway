package model

const StaffStateNormal = 1

type ArchiveStaffCanalMsg struct {
	Action string        `json:"action"`
	Table  string        `json:"table"`
	New    *ArchiveStaff `json:"new"`
	Old    *ArchiveStaff `json:"old"`
}

type ArchiveStaff struct {
	ID           int64  `json:"id"`
	Aid          int64  `json:"aid"`
	Mid          int64  `json:"mid"`
	StaffMid     int64  `json:"staff_mid"`
	StaffTitle   string `json:"staff_title"`
	StaffTitleId int64  `json:"staff_title_id"`
	State        int64  `json:"state"`
}
