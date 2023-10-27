package xlog

import (
	"testing"
)

func TestInfoSDepth(t *testing.T) {
	type args struct {
		depth         int
		msg           string
		keysAndValues []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"t1", args{
			1, "msg1,depth=1,has kv.", []interface{}{"tk1", "tv1", "tk2", "tv2"},
		}},
		// {"t2", args{
		// 	2, "msg2,depth=2", []interface{}{},
		// }},
	}
	logging.setVState(3)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			V(2).InfoSDepth(tt.args.depth, tt.args.msg, tt.args.keysAndValues...)
			V(3).InfoSDepth(tt.args.depth, tt.args.msg, tt.args.keysAndValues...)
			V(4).InfoSDepth(tt.args.depth, tt.args.msg, tt.args.keysAndValues...)
			InfoSDepth(tt.args.depth, tt.args.msg, tt.args.keysAndValues...)
			WarningDepth(tt.args.depth, tt.args.msg)
		})
	}
}
