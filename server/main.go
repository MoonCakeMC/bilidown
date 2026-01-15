package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"bilidown/router"
	"bilidown/util"

	_ "modernc.org/sqlite"
)

const (
	HTTP_PORT = 25008      // 限定 HTTP 服务器端口
	HTTP_HOST = ""        // 限定 HTTP 服务器主机
	VERSION   = "v2.0.15-modified" // 软件版本号，将影响托盘标题显示
)

func main() {
	checkFFmpeg()
	
	// 初始化数据表
	mustInitTables()
	// 配置和启动 HTTP 服务器
	mustRunServer()
	// 保持运行
	select {}
}

// checkFFmpeg 检测 ffmpeg 的安装情况，如果未安装则打印提示信息。
func checkFFmpeg() {
	if _, err := util.GetFFmpegPath(); err != nil {
		fmt.Println("🚨 FFmpeg is missing. Install it from https://www.ffmpeg.org/download.html or place it in ./bin, then restart the application.")
		select {}
	}
}

// 配置和启动 HTTP 服务器
func mustRunServer() {
	// 前端打包文件
	http.Handle("/", http.FileServer(http.Dir("static")))
	// 后端接口服务
	http.Handle("/api/", http.StripPrefix("/api", router.API()))
	// 启动 HTTP 服务器
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", HTTP_HOST, HTTP_PORT), nil)
		if err != nil {
			log.Fatal("http.ListenAndServe:", err)
		}
	}()
}

// mustReadFile 返回文件字节内容
func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("os.ReadFile:", err)
	}
	return data
}

// mustInitTables 初始化数据表
func mustInitTables() {
	db := util.MustGetDB()
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "field" (
		"name" TEXT PRIMARY KEY NOT NULL,
		"value" TEXT
	)`); err != nil {
		log.Fatalln("create table field:", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "log" (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"content" TEXT NOT NULL,
		"create_at" text NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		log.Fatalln("create table log:", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS "task" (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"bvid" text NOT NULL,
		"cid" integer NOT NULL,
		"format" integer NOT NULL,
		"title" text NOT NULL,
		"owner" text NOT NULL,
		"cover" text NOT NULL,
		"status" text NOT NULL,
		"folder" text NOT NULL,
		"duration" integer NOT NULL,
		"create_at" text NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		log.Fatalln("create table task:", err)
	}

	if _, err := util.GetCurrentFolder(db); err != nil {
		log.Fatalln("util.GetCurrentFolder:", err)
	}

	if err := initHistoryTask(db); err != nil {
		log.Fatalln("initHistoryTask:", err)
	}
}

// initHistoryTask 将上一次程序运行时未完成的任务进度全部变为 error
func initHistoryTask(db *sql.DB) error {
	util.SqliteLock.Lock()
	_, err := db.Exec(`UPDATE "task" SET "status" = 'error' WHERE "status" IN ('waiting', 'running')`)
	util.SqliteLock.Unlock()
	return err
}
