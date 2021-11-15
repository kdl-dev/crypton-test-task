package logs

import (
	"fmt"
	"server/pkg/model"
)

type LogFormat struct {
	Uid model.Uid
	FileName string
	Size int64
	Mode string
}

func (l LogFormat) String() string {
	return fmt.Sprintf("uid: %d, name: %s, size: %d, %s.",
		l.Uid,
		l.FileName,
		l.Size,
		l.Mode)
}
