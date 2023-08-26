package v1

import "fmt"

func (m *SeriesPointMatchConfig) ManualValidate() error {
	if !m.UseTeamGroup {
		return nil
	}
	if len(m.Teams) == 0 {
		return fmt.Errorf("use_team_group=true but team_group=nil")
	}
	for idx, group := range m.Teams {
		if group.Group == "" {
			return fmt.Errorf("team_group[%v] group is empty", idx)
		}
	}
	return nil
}

func (m *SeriesKnockoutContestInfoItem) travel(fn func(m *SeriesKnockoutContestInfoItem) error) error {
	queue := []*SeriesKnockoutContestInfoItem{m}
	level := 0
	for len(queue) != 0 {
		checkpoint := len(queue)
		for i := 0; i < checkpoint; i++ {
			if queue[i] != nil {
				if err := fn(queue[i]); err != nil {
					return err
				}
				for _, n := range queue[i].Children {
					queue = append(queue, n)
				}
			}
		}
		queue = queue[checkpoint:]
		level++
	}
	return nil
}

func (m *SeriesKnockoutMatchInfo) Travel(fn func(m *SeriesKnockoutContestInfoItem) error) error {
	for _, group := range m.Groups {
		if err := group.travel(fn); err != nil {
			return err
		}
	}
	return nil
}
