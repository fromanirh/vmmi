package convsched

import (
	"encoding/json"
	"fmt"
	"github.com/fromanirh/vmmi/pkg/vmmi/progress"
	"github.com/fromanirh/vmmi/pkg/xstrings"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	ActionAbort          = "abort"
	ActionEnablePostCopy = "postcopy"
	ActionSetDowntime    = "setDowntime"
)

type ConvergenceAction struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

func (action ConvergenceAction) String() string {
	return fmt.Sprintf("%s(%s)", action.Name, strings.Join(action.Params, ", "))
}

type VMMigrator interface {
	SetDowntime(value int) error
	StartPostCopy() error
	Abort() error
	Progress() *progress.Progress
}

func (action *ConvergenceAction) Exec(mig VMMigrator) error {
	var err error
	switch action.Name {
	case ActionSetDowntime:
		downtime, err := strconv.Atoi(action.Params[0])
		if err == nil {
			return err
		}
		err = mig.SetDowntime(downtime)
	case ActionEnablePostCopy:
		err = mig.StartPostCopy()
	case ActionAbort:
		err = mig.Abort()
	}
	return err
}

type ConvergenceItem struct {
	Action ConvergenceAction `json:"action"`
	Limit  int64             `json:"limit"`
}

func (item ConvergenceItem) String() string {
	return fmt.Sprintf("%s@%d", item.Action, item.Limit)
}

type ConvergenceSchedule struct {
	Init     []ConvergenceAction `json:"init"`
	Stalling []ConvergenceItem   `json:"stalling"`
}

func (sched ConvergenceSchedule) String() string {
	return fmt.Sprintf("{ init=%v stalling=%v }", xstrings.Join(sched.Init), xstrings.Join(sched.Stalling))
}

func (cs *ConvergenceSchedule) HasPostcopy() bool {
	for _, item := range cs.Stalling {
		if item.Action.Name == ActionEnablePostCopy {
			return true
		}
	}
	return false
}

func (cs *ConvergenceSchedule) PopAction(iteration int64) *ConvergenceAction {
	var ret *ConvergenceAction
	if cs.Stalling[0].Limit < iteration {
		ret = &cs.Stalling[0].Action
		cs.Stalling = cs.Stalling[1:]
	}
	return ret

}

func Load(r io.Reader) (*ConvergenceSchedule, error) {
	dec := json.NewDecoder(r)
	var cs ConvergenceSchedule
	err := dec.Decode(&cs)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

type ConvergenceScheduleConfiguration struct {
	Schedule        ConvergenceSchedule `json:"schedule"`
	MonitorInterval time.Duration       `json:"monitorInterval"`
}

type ConfigurationMessage struct {
	Configuration ConvergenceScheduleConfiguration `json:"configuration"`
}

func LoadConfiguration(r io.Reader) (*ConvergenceScheduleConfiguration, error) {
	dec := json.NewDecoder(r)
	var conf ConfigurationMessage
	err := dec.Decode(&conf)
	if err != nil {
		return nil, err
	}
	return &conf.Configuration, nil
}
