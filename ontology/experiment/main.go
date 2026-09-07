package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// チームで合意していたドメインの説明。第1回でそのまま渡したもの。
const domainDoc = `
申請には次の状態がある。
- 下書き
- 申請中
- 承認済み
- 差戻し

申請者は下書きの申請を提出できる。
承認者は申請中の申請を承認、または差戻しできる。
承認者は申請者と同一人物であってはならない。
`

const instruction = `
上記のドメインをGoで実装してください。
状態を表す型と、状態遷移を行う関数を定義してください。
不正な遷移はエラーにしてください。
コードだけを返してください。
`

// 条件ごとに domainDoc へ足す一文。
type condition struct {
	name string
	add  string
}

var conditions = []condition{
	{
		name: "baseline",
		add:  "",
	},
	{
		name: "b-word",
		add:  "差戻された申請は、修正して再提出できる。\n",
	},
	{
		name: "c-transition",
		add:  "差戻しの状態から、申請中の状態に戻ることがある。\n",
	},
}

const runsPerCondition = 5

// 環境変数はコピー時に改行や空白が混ざることがあるため、前後を落としておく。
func apiKey() string {
	return strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
}

func ask(prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      "claude-sonnet-4-6",
		"max_tokens": 2000,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey())
	req.Header.Set("anthropic-version", "2023-06-01")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d: %s", res.StatusCode, raw)
	}

	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}

	var out bytes.Buffer
	for _, c := range parsed.Content {
		out.WriteString(c.Text)
	}
	return out.String(), nil
}

func main() {
	if apiKey() == "" {
		fmt.Fprintln(os.Stderr, "ANTHROPIC_API_KEY を設定してください")
		os.Exit(1)
	}

	for _, c := range conditions {
		dir := filepath.Join("out", c.name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}

		prompt := domainDoc + c.add + instruction

		// 実際に何を渡したかを残しておく。
		promptPath := filepath.Join(dir, "prompt.txt")
		if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
			panic(err)
		}

		for i := 1; i <= runsPerCondition; i++ {
			text, err := ask(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s run %d: %v\n", c.name, i, err)
				continue
			}

			path := filepath.Join(dir, fmt.Sprintf("run-%02d.go.txt", i))
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				panic(err)
			}
			fmt.Printf("%s run %d: %s\n", c.name, i, path)
		}
	}
}
