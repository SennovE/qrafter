package expr

import "github.com/SennovE/qrafter/internal/core"

type CheckExperssion interface {
	core.Renderer
	expression()
}
