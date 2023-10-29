package xlog

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tomhjx/xlog/internal/severity"
)

func createTestingUniqueID() string {
	return uuid.NewString()
}

func writeTestingGlobalLogs(vl int, sl severity.Severity, depth int, a any) map[severity.Severity][]string {
	createLog := func(m string, args ...any) string {
		return fmt.Sprint(m, "#", createTestingUniqueID(), args)
	}

	info := Info
	infoDepth := InfoDepth
	infoSDepth := InfoSDepth
	errorS := ErrorS

	ignoreWritelogs := map[severity.Severity][]string{}

	if vl > 0 {
		verb := V(Level(vl))
		info = func(args ...any) {
			verb.Info(args)
		}
		infoDepth = func(depth int, args ...any) {
			verb.InfoDepth(depth, args)
		}
		infoSDepth = func(depth int, msg string, keysAndValues ...any) {
			verb.InfoSDepth(depth, msg, keysAndValues...)
		}
		errorS = func(err error, msg string, keysAndValues ...any) {
			verb.ErrorS(err, msg, keysAndValues...)
		}
		ignoreWritelogs[severity.WarningLog] = []string{"Warning", "WarningDepth"}
		ignoreWritelogs[severity.ErrorLog] = []string{"Error", "ErrorDepth"}
	}

	writelogs := map[severity.Severity]map[string]func(depth int, a any, kvs []any){}

	for _, v := range []severity.Severity{severity.InfoLog, severity.WarningLog, severity.ErrorLog} {
		writelogs[v] = map[string]func(depth int, a any, kvs []any){}
	}

	writelogs[severity.InfoLog]["Info"] = func(depth int, a any, kvs []any) { info(a) }
	writelogs[severity.InfoLog]["InfoDepth"] = func(depth int, a any, kvs []any) { infoDepth(depth, a) }
	writelogs[severity.InfoLog]["InfosDepth"] = func(depth int, a any, kvs []any) { infoSDepth(depth, fmt.Sprint(a), kvs...) }
	writelogs[severity.WarningLog]["Warning"] = func(depth int, a any, kvs []any) { Warning(a) }
	writelogs[severity.WarningLog]["WarningDepth"] = func(depth int, a any, kvs []any) { WarningDepth(depth, a) }
	writelogs[severity.ErrorLog]["Error"] = func(depth int, a any, kvs []any) { Error(a) }
	writelogs[severity.ErrorLog]["Errors"] = func(depth int, a any, kvs []any) { errorS(fmt.Errorf("%s", a), fmt.Sprint(a), kvs...) }
	writelogs[severity.ErrorLog]["ErrorDepth"] = func(depth int, a any, kvs []any) { ErrorDepth(depth, a) }

	for sl, svs := range ignoreWritelogs {
		for _, sv := range svs {
			delete(writelogs[sl], sv)
		}
	}

	ret := map[severity.Severity][]string{}

	for m, w := range writelogs[sl] {
		kvs := []any{fmt.Sprint(m, "key", time.Now()), fmt.Sprint(m, "val", time.Now())}
		c := createLog(m, vl, sl, depth, a, kvs)
		w(depth, c, kvs)
		ret[sl] = append(ret[sl], c)
	}

	return ret
}

func TestGlobalLogger(t *testing.T) {
	logFile := fmt.Sprintf("%sxlog-testing/%s.log", os.TempDir(), time.Now().Format("20060102/150405"))
	defer os.Remove(logFile)
	t.Log("create file:", logFile)
	SetFile(logFile)

	type option struct {
		name         string
		sv           int
		ss           severity.Severity
		v            int
		s            severity.Severity
		depth        int
		wantContains bool
	}

	tests := []option{}
	for i := 0; i < 2; i++ {
		o := option{
			name:         "log general",
			wantContains: true,
			depth:        i,
		}
		tests = append(tests, o)
	}
	o := option{
		name:         "log with verbose",
		sv:           2,
		wantContains: false,
	}

	for i := 0; i < 4; i++ {
		o.v = i
		o.wantContains = false
		if o.sv >= o.v {
			o.wantContains = true
		}
		pwc := o.wantContains
		tests = append(tests, o)
		o.ss = severity.WarningLog
		for s, sn := range severity.Name {
			o.name = fmt.Sprintf("log with verbose,severity %s", sn)
			o.s = severity.Severity(s)
			o.wantContains = pwc
			if o.ss > o.s && o.wantContains {
				o.wantContains = false
			}
			tests = append(tests, o)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testOption := map[string]any{}
			testOption["ss"] = tt.ss
			testOption["s"] = tt.s
			testOption["sv"] = tt.sv
			testOption["v"] = tt.v

			SetVerbosity(tt.sv)
			SetSeverity(tt.ss)
			logs := writeTestingGlobalLogs(tt.v, tt.s, tt.depth, tt.name)
			r, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatal(err)
			}
			rlog := string(r)
			check := func(logKeys ...string) {
				for _, lk := range logKeys {
					assert.Contains(t, rlog, lk, testOption)
				}
			}
			if !tt.wantContains {
				check = func(logKeys ...string) {
					for _, lk := range logKeys {
						assert.NotContains(t, rlog, lk, testOption)
					}
				}
			}
			for _, log := range logs {
				check(log...)
			}
		})
	}

}
