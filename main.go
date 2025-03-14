package main

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// RSSフィード構造体
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel構造体
type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

// Item構造体
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}


// メッセージを受信したときの処理
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Bot 自身のメッセージは無視
    if m.Author.ID == s.State.User.ID {
        return
    }

    content := strings.ToLower(m.Content)
    if strings.HasPrefix(content, "!zenn ") {
        topic := strings.TrimPrefix(content, "!zenn ")
        rss := get(topic)
        if len(rss.Channel.Items) == 0 {
            s.ChannelMessageSend(m.ChannelID, "記事が見つからなかったよ😢")
            return
        }

        articles := []string{}
        for _, item := range rss.Channel.Items {
            articles = append(articles, fmt.Sprintf("今回のピックアップ記事は...\n%s\n   %s\n", item.Title, item.Link))
        }

        // ランダムなインデックスを生成
        n, err := rand.Int(rand.Reader, big.NewInt(int64(len(articles))))
        if err != nil {
            panic(err)
        }

        // ランダムな記事を送信
        s.ChannelMessageSend(m.ChannelID, articles[n.Int64()])
    } else if content == "!zenn" {
        s.ChannelMessageSend(m.ChannelID, "トピックを選択してください")
    }
}

func get(topicsName string) RSS{
	var rss RSS
	// URLを直接コードに組み込む
	url := fmt.Sprintf("https://zenn.dev/topics/%s/feed", topicsName)// このURLを目的のRSSフィードURLに変更してください
		
	// RSSフィードを取得
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "フィードの取得に失敗しました: %v\n", err)

	}
	defer resp.Body.Close()

	// レスポンスのステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "サーバーからエラー応答: %s\n", resp.Status)

	}

	// RSSフィードをパース
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "レスポンスの読み込みに失敗しました: %v\n", err)

	}


	err = xml.Unmarshal(body, &rss)
	if err != nil {
		fmt.Fprintf(os.Stderr, "XMLのパースに失敗しました: %v\n", err)
	}

	// チャンネルのリンクを表示
	fmt.Printf("チャンネルリンク: %s\n", rss.Channel.Link)

	// 各アイテムのリンクを表示
	fmt.Println("\n記事のリンク:")
	for i, item := range rss.Channel.Items {
		fmt.Printf("%d. %s\n   %s\n", i+1, item.Title, item.Link)
	}
	return rss
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
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		messageCreate(s, m)
	})

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