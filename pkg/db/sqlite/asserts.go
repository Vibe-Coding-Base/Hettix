package sqlite

import (
	"github.com/Vibe-Coding-Base/Hettix/pkg/proj"
	"github.com/Vibe-Coding-Base/Hettix/pkg/reqlog"
	"github.com/Vibe-Coding-Base/Hettix/pkg/sender"
)

var (
	_ reqlog.Repository = (*Database)(nil)
	_ sender.Repository = (*Database)(nil)
	_ proj.Repository   = (*Database)(nil)
)
