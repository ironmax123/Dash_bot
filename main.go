package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// メッセージを受信したときの処理
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot 自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// 「hello」と送信されたら「やっほー٩( ᐛ )و」と返信
	if strings.ToLower(m.Content) == "hello" {
		s.ChannelMessageSend(m.ChannelID, "やっほー٩( ᐛ )و ダッシュです！")
	}
}

func main() {
	// 環境変数から Discord Bot のトークンを取得
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Println("トークンが設定されていません。環境変数 DISCORD_BOT_TOKEN を設定してください。")
		return
	}

	// 必要な Intent を設定して Discord セッションを作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("エラー: ", err)
		return
	}

	// すべてのメッセージを取得するための Intents を設定
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// メッセージ受信時のイベントを登録
	dg.AddHandler(messageCreate)

	// WebSocket 接続を開く
	err = dg.Open()
	if err != nil {
		fmt.Println("Bot を起動できませんでした:", err)
		return
	}

	fmt.Println("Bot が起動しました！ Ctrl+C で終了できます。")

	// 終了を待つ
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Bot を停止
	dg.Close()
}
