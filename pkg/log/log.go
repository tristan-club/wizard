package log

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/pkg/cluster/otgrpc"
	"google.golang.org/grpc/metadata"
	"os"
	"runtime"
	"strconv"
	"time"
)

var logger *zerolog.ConsoleWriter

const (
	traceId = "traceid"
)

func init() {
	if config.EnvIsDev() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if config.UseConsoleWrite() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

}

func Error() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Error()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}
	return e
}

func ErrorCtx(ctx context.Context) *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Error()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}

	if ctx != nil {
		var md metadata.MD
		if v, ok := ctx.Value(otgrpc.MdContextKey).(metadata.MD); ok {
			md = v
		} else if v, ok := metadata.FromIncomingContext(ctx); ok {
			md = v
		}

		var tid string

		if md != nil {
			if vs := md.Get(traceId); len(vs) > 0 {
				tid = vs[0]
			}
		}

		if len(tid) > 0 {
			e.Str(traceId, tid)
		}
	} else {
		Error().Msgf("ctx nil")
	}

	return e
}

func Debug() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Debug()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}
	return e
}

func Warn() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Warn()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}
	return e
}

func Info() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Info()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}
	return e
}

func InfoCtx(ctx context.Context) *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Info()
	if ok {
		e = e.Str("line", file+":"+strconv.Itoa(line))
	}
	if ctx != nil {
		var md metadata.MD
		if v, ok := ctx.Value(otgrpc.MdContextKey).(metadata.MD); ok {
			md = v
		} else if v, ok := metadata.FromIncomingContext(ctx); ok {
			md = v
		}

		var tid string

		if md != nil {
			if vs := md.Get(traceId); len(vs) > 0 {
				tid = vs[0]
			}
		}

		if len(tid) > 0 {
			e.Str(traceId, tid)
		}
	} else {
		Error().Msgf("ctx nil")
	}

	return e
}
