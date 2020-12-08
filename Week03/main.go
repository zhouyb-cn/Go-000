// 基于errgroup 实现一个http server的启动和关闭
// 以及linux singal信号的注册和处理，保证一个退出，全部注销退出
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const PORT = ":9001"

func main() {
	done := make(chan bool, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, _ := errgroup.WithContext(ctx)

	srv := http.Server{Addr: "127.0.0.1" + PORT}

	// server
	g.Go(func() error {
		err := srv.ListenAndServe()
		return err
	})

	// 信号注册处理
	g.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-quit:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			srv.SetKeepAlivesEnabled(false)
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Could not gracefully shutdown the server: %v\n", err)
			}
			close(done)
		case <-ctx.Done():
			log.Println(ctx.Err())
		}
		return nil
	})

	// 模拟2秒主动退出
	//go func() {
	//	time.Sleep(time.Second * 2)
	//	cancel()
	//}()

	if err := g.Wait(); err != nil {
		cancel()
		log.Printf("err: %v", err)
	}

	<-done
	fmt.Println("all server shutdown")
}
