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

// ringi の domain/application/status.go をそのまま渡す。
// 第1回では日本語の散文を渡したが、今回は型と遷移表というGoのコードを渡す。
const statusGo = "```go\n" + `// domain/application/status.go
package application

import "fmt"

// Status は申請の状態を表す値オブジェクト
type Status string

const (
	StatusDraft     Status = "draft"     // 下書き
	StatusSubmitted Status = "submitted" // 申請中
	StatusApproved  Status = "approved"  // 承認済み
	StatusRejected  Status = "rejected"  // 差戻し
)

// transitions は許可される状態遷移の一覧。
// 「どの状態からどこへ動けるか」というドメインルールはここに集約する
var transitions = map[Status][]Status{
	StatusDraft:     {StatusSubmitted},                // 下書き → 提出
	StatusSubmitted: {StatusApproved, StatusRejected}, // 申請中 → 承認 or 差戻し
	StatusRejected:  {StatusSubmitted},                // 差戻し → 再提出
	StatusApproved:  {},                               // 承認済みは終端状態
}

// NewStatus は外部入力(DBの値、JSONなど)から Status を復元する。
// 定義外の値はエラーにする(値オブジェクトの自己検証)
func NewStatus(value string) (Status, error) {
	s := Status(value)
	if _, ok := transitions[s]; !ok {
		return "", fmt.Errorf("不正な状態です: %q", value)
	}
	return s, nil
}

// CanTransitionTo は next への遷移が許可されているかを返す
func (s Status) CanTransitionTo(next Status) bool {
	for _, allowed := range transitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}
` + "```\n"

// 承認ステップの存在だけを最小限伝える。
// この説明文が結果に影響することは承知のうえで、リセットの扱いを観察するために置く。
const stepDoc = `
申請は承認ルートを持ちます。承認ルートは承認者の並びで、「課長 → 部長」のような多段承認を表します。
承認は先頭の段から順に行われ、順番を飛ばすことはできません。
すべての段が承認されたとき、申請は承認済みになります。
`

const instruction = `
上記の Status 型と遷移表を使って、申請(Application)を実装してください。
提出、承認、差戻し、再提出のふるまいを定義してください。
遷移表に従わない遷移はエラーにしてください。
コードだけを返してください。
`

type condition struct {
	name   string
	prompt string
}

var conditions = []condition{
	// 遷移表のみ。ステップの説明を足さない対照。
	{
		name:   "table-only",
		prompt: statusGo + instruction,
	},
	// 遷移表 + 承認ステップの説明。リセットの扱いを見る本命。
	{
		name:   "table-steps",
		prompt: statusGo + stepDoc + instruction,
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
		"max_tokens": 3000,
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

		// 実際に何を渡したかを残しておく。
		if err := os.WriteFile(
			filepath.Join(dir, "prompt.txt"), []byte(c.prompt), 0o644,
		); err != nil {
			panic(err)
		}

		for i := 1; i <= runsPerCondition; i++ {
			text, err := ask(c.prompt)
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
