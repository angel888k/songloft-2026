package jsplugin

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// 阻塞真实缓存任务的最后一步，强制它在停止/卸载/关闭期间仍未退出。
// 不替换编译器或写盘实现，也不依赖调度速度碰中临时目录清理竞态。
type cacheCompletionBarrier struct {
	slog.Handler
	path    string
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (h *cacheCompletionBarrier) Handle(ctx context.Context, record slog.Record) error {
	if record.Message == "bytecode cached" {
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "path" && attr.Value.String() == h.path {
				h.once.Do(func() {
					close(h.entered)
					<-h.release
				})
			}
			return true
		})
	}
	return h.Handler.Handle(ctx, record)
}

func TestBytecodeCache_ShutdownWaitsForCompletion(t *testing.T) {
	for _, action := range []string{"stop", "unload", "close"} {
		t.Run(action, func(t *testing.T) {
			pluginsDir, dataDir, repo, _ := setupTestEnv(t)
			ctx := context.Background()
			manifest := testManifest("test-cache-shutdown")
			zipData := createTestPluginZip(t, manifest, simpleJSCode)
			zipFileName := manifest.EntryPath + ".jsplugin.zip"
			if err := os.WriteFile(filepath.Join(pluginsDir, zipFileName), zipData, 0o644); err != nil {
				t.Fatal(err)
			}
			plugin := &JSPlugin{
				Name: manifest.Name, Version: manifest.Version, EntryPath: manifest.EntryPath,
				Main: manifest.Main, Permissions: manifest.Permissions,
				Status: JSPluginStatusActive, FilePath: zipFileName,
			}
			if err := repo.Create(ctx, plugin); err != nil {
				t.Fatal(err)
			}

			logger := slog.Default()
			logWriter, logFlags := log.Writer(), log.Flags()
			barrier := &cacheCompletionBarrier{
				Handler: slog.NewTextHandler(os.Stderr, nil),
				path:    filepath.Join(dataDir, manifest.EntryPath, "cache", "main.jsc"),
				entered: make(chan struct{}), release: make(chan struct{}),
			}
			slog.SetDefault(slog.New(barrier))
			t.Cleanup(func() {
				slog.SetDefault(logger)
				log.SetOutput(logWriter)
				log.SetFlags(logFlags)
			})
			manager := NewManager(repo, pluginsDir, dataDir, "", nil, nil)
			t.Cleanup(func() { _ = manager.Close() })
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(barrier.release) }) }
			defer unblock()

			if err := manager.LoadPlugin(ctx, plugin); err != nil {
				t.Fatal(err)
			}
			select {
			case <-barrier.entered:
			case <-time.After(5 * time.Second):
				t.Fatal("cache task did not reach completion barrier")
			}
			done := make(chan error, 1)
			go func() {
				switch action {
				case "stop":
					service, _ := manager.GetService(manifest.EntryPath)
					done <- service.Stop()
				case "unload":
					done <- manager.UnloadPlugin(ctx, manifest.EntryPath)
				case "close":
					done <- manager.Close()
				}
			}()
			select {
			case err := <-done:
				t.Fatalf("%s returned before cache task completed: %v", action, err)
			case <-time.After(100 * time.Millisecond):
			}
			unblock()
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("%s did not finish after cache task completed", action)
			}
			if _, ok := loadBytecodeCache(barrier.path, barrier.path+".sha256", manifest.EntryHash); !ok {
				t.Fatal("shutdown did not preserve a valid bytecode cache")
			}
		})
	}
}
