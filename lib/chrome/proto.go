package chrome

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type ProtoReq[T any] interface {
	ProtoReq() string
	Call(proto.Client) (*T, error)
}

func CDPCommand[I ProtoReq[O], O any](ctx context.Context, page *rod.Page, req I) (res O, err error) {
	if ctx == nil {
		ctx = page.GetContext()
	}
	resBuff, err := page.Call(ctx, string(page.SessionID), req.ProtoReq(), req)
	if err != nil {
		err = fmt.Errorf("%s: %w", req.ProtoReq(), err)
		return
	}
	err = sonic.Unmarshal(resBuff, &res)
	if err != nil {
		err = fmt.Errorf("%s: %w", req.ProtoReq(), err)
		return
	}
	return
}

type ProtoProcedure interface {
	ProtoReq() string
	Call(proto.Client) error
}

func CDPProcedure[I ProtoProcedure](ctx context.Context, page *rod.Page, req I) (err error) {
	if ctx == nil {
		ctx = page.GetContext()
	}
	_, err = page.Call(ctx, string(page.SessionID), req.ProtoReq(), req)
	if err != nil {
		err = fmt.Errorf("%s: %w", req.ProtoReq(), err)
		return
	}
	return
}
