package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/expr-lang/expr"

	"github.com/gptscript-ai/gptscript/pkg/counter"
	"github.com/gptscript-ai/gptscript/pkg/prompt"
	"github.com/gptscript-ai/gptscript/pkg/types"
)

func (e *Engine) runInterpExpr(ctx Context, input string) (*Return, error) {
	// get code to eval
	_, script, _ := strings.Cut(ctx.Tool.Instructions, "\n")
	if script == "" {
		return nil, nil
	}

	// setup environment for the interpreter
	gptsCtx, err := json.Marshal(
		map[string]any{
			"program": ctx.Program,
			"call":    ctx.GetCallContext(),
		},
	)
	if err != nil {
		return nil, err
	}
	envExpr := map[string]any{
		"GPTSCRIPT_INPUT":   input,
		"GPTSCRIPT_CONTEXT": string(gptsCtx),
		// useful stuff from stdlib
		"fmtSprintf": fmt.Sprintf,
		// gptscript prompt
		"sysPrompt": func(in string) (string, error) {
			return prompt.SysPrompt(ctx.WrappedContext(e), e.Env, in, nil)
		},
	}

	// eval code
	outAny, err := expr.Eval(script, envExpr)
	if err != nil {
		return nil, err
	}
	out := fmt.Sprint(outAny)

	e.Progress <- types.CompletionStatus{
		CompletionID: counter.Next(),
		Response: map[string]any{
			"output": outAny,
			"err":    nil,
		},
	}

	return &Return{
		Result: &out,
	}, nil
}
