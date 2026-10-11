package jsruntime

import (
	"context"
	"fmt"
	"testing"
)

// 使用真实 QuickJS polyfill 和可控时钟，覆盖 #514 的句柄复用及回调交错。
func TestTimerHandleSafety(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "expired timeout cleanup cannot cancel a new listener",
			code: `
				var old = setTimeout(function() { clearTimeout(old); }, 0);
				__processExpiredTimers();
				var fired = [];
				setTimeout(function() { fired.push('listener'); }, 0);
				clearTimeout(old);
				clearInterval(old);
				__processExpiredTimers();
				return { fired: fired, pending: __getPendingOneShotTimerCount() };
			`,
			want: `{"fired":["listener"],"pending":0}`,
		},
		{
			name: "duplicate and unknown cancellations cannot overwrite live timers",
			code: `
				clearTimeout(undefined);
				clearInterval(null);
				clearTimeout(-1);
				var old = setTimeout(function() {}, 1000);
				clearTimeout(old);
				clearTimeout(old);
				clearInterval(old);
				var fired = [];
				setTimeout(function() { fired.push('first'); }, 0);
				setTimeout(function() { fired.push('second'); }, 0);
				__processExpiredTimers();
				return fired;
			`,
			want: `["first","second"]`,
		},
		{
			name: "expiration snapshot cannot execute a replacement early",
			code: `
				var fired = [];
				var canceled;
				setTimeout(function() {
					clearTimeout(canceled);
					setTimeout(function() { fired.push('replacement'); }, 1000);
				}, 0);
				canceled = setTimeout(function() { fired.push('canceled'); }, 0);
				__processExpiredTimers();
				var early = fired.slice();
				var deadline = __getNextTimerDeadline();
				now += 1000;
				__processExpiredTimers();
				return { early: early, deadline: deadline, fired: fired };
			`,
			want: `{"early":[],"deadline":1100,"fired":["replacement"]}`,
		},
		{
			name: "interval self cancellation and stale cancellation preserve other timers",
			code: `
				var ticks = 0;
				var listenerFired = false;
				var interval = setInterval(function() {
					if (++ticks === 2) {
						clearInterval(interval);
						setTimeout(function() { listenerFired = true; }, 0);
						clearInterval(interval);
					}
				}, 10);
				var pendingWithInterval = __getPendingOneShotTimerCount();
				now += 10;
				__processExpiredTimers();
				now += 10;
				__processExpiredTimers();
				__processExpiredTimers();
				now += 10;
				__processExpiredTimers();
				return { ticks: ticks, listenerFired: listenerFired, pendingWithInterval: pendingWithInterval, deadline: __getNextTimerDeadline() };
			`,
			want: `{"ticks":2,"listenerFired":true,"pendingWithInterval":0,"deadline":0}`,
		},
		{
			name: "timeout and interval share cancellation semantics",
			code: `
				var fired = [];
				clearInterval(setTimeout(function() { fired.push('timeout'); }, 0));
				clearTimeout(setInterval(function() { fired.push('interval'); }, 10));
				now += 10;
				__processExpiredTimers();
				return { fired: fired, pending: __getPendingOneShotTimerCount(), deadline: __getNextTimerDeadline() };
			`,
			want: `{"fired":[],"pending":0,"deadline":0}`,
		},
	}
	for _, timerType := range []string{"setTimeout", "setInterval"} {
		tests = append(tests, struct {
			name string
			code string
			want string
		}{
			name: timerType + " fails at the safe integer limit without reusing a handle",
			code: fmt.Sprintf(`
				__timerIdCounter = Number.MAX_SAFE_INTEGER - 1;
				var fired = false;
				var last = %s(function() { fired = true; }, 10);
				var errors = [];
				for (var i = 0; i < 2; i++) {
					try { %s(function() {}, 10); } catch(e) { errors.push(e.name); }
				}
				now += 10;
				__processExpiredTimers();
				clearTimeout(last);
				return { errors: errors, fired: fired, deadline: __getNextTimerDeadline() };
			`, timerType, timerType),
			want: `{"errors":["RangeError","RangeError"],"fired":true,"deadline":0}`,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJSEnvManager()
			defer manager.SignalShutdown()
			const envID = "test-timer-handle-safety"
			if err := manager.CreateEnv(envID, polyfillJS, 1); err != nil {
				t.Fatalf("CreateEnv: %v", err)
			}
			defer manager.DestroyEnv(envID)
			code := `JSON.stringify((function() {
				var now = 100;
				var originalNow = __go_now_ms;
				__go_now_ms = function() { return now; };
				try {` + tt.code + `} finally { __go_now_ms = originalNow; }
			})())`
			result, err := manager.ExecuteJS(context.Background(), envID, code, 1000)
			if err != nil {
				t.Fatalf("ExecuteJS: %v", err)
			}
			if result.Result != tt.want {
				t.Fatalf("timer outcome = %s, want %s", result.Result, tt.want)
			}
		})
	}
}
