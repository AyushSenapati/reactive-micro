package logger

import (
	"context"
	"os"

	kitlog "github.com/go-kit/kit/log"
	kitllvl "github.com/go-kit/kit/log/level"

	svcconf "github.com/AyushSenapati/reactive-micro/inventorysvc/conf"
)

func NewLogger(env string) *CustomLogger {
	var opts []kitllvl.Option
	var logger kitlog.Logger

	if env == "dev" {
		opts = append(opts, kitllvl.AllowDebug())
		logger = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stdout))
	} else {
		logger = kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))

		if env == "test" || env == "testing" {
			opts = append(opts, kitllvl.AllowDebug())
		} else if env == "staging" {
			opts = append(opts, kitllvl.AllowInfo())
		} else {
			opts = append(opts, kitllvl.AllowWarn())
		}
	}

	logger = kitllvl.NewFilter(logger, opts...)

	return &CustomLogger{logger}
}

type CustomLoggerOpt func(*CustomLogger) error

func WithSvcName(name string) CustomLoggerOpt {
	return func(cl *CustomLogger) error {
		cl.l = kitlog.With(cl.l, "svc", name)
		return nil
	}
}

func WithTimeStamp() CustomLoggerOpt {
	return func(cl *CustomLogger) error {
		cl.l = kitlog.With(cl.l, "ts", kitlog.DefaultTimestampUTC)
		return nil
	}
}

type CustomLogger struct {
	l kitlog.Logger
}

func (cl *CustomLogger) Configure(opts ...CustomLoggerOpt) {
	for _, o := range opts {
		o(cl)
	}
}

const msgKey = "msg"

func (cl *CustomLogger) Debug(ctx context.Context, msg interface{}) {
	logWithCtx(kitllvl.Debug(cl.l), ctx, msg)
}

func (cl *CustomLogger) Info(ctx context.Context, msg interface{}) {
	logWithCtx(kitllvl.Info(cl.l), ctx, msg)
}

func (cl *CustomLogger) Warn(ctx context.Context, msg interface{}) {
	logWithCtx(kitllvl.Warn(cl.l), ctx, msg)
}

func (cl *CustomLogger) Error(ctx context.Context, msg interface{}) {
	logWithCtx(kitllvl.Error(cl.l), ctx, msg)
}

// LogIfError is a helper to log only non-nil errors
func (cl *CustomLogger) LogIfError(ctx context.Context, err error) {
	if err != nil {
		cl.Error(ctx, err)
	}
}

func logWithCtx(l kitlog.Logger, ctx context.Context, msg interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	reqID := ctx.Value(svcconf.C.ReqIDKey)
	if reqID == nil {
		reqID = ""
	}
	l.Log("trace-id", reqID, msgKey, msg)
}
