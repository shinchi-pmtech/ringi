# ontology/experiment2

連載第2回で使った実験コードです。第1回では日本語の散文を渡しましたが、今回は
`domain/application/status.go` の型と遷移表をそのまま渡し、返ってきたGoの実装を観察します。

## 実験の設計

`status.go` に条件ごとの説明を足して、それぞれ5回ずつ流します。

| 条件 | 渡したもの |
|---|---|
| table-only | `status.go` のみ（対照） |
| table-steps | `status.go` + 承認ルートの説明3行 |

table-only を置いたのは、承認ステップやそのリセットが、足した説明文によるものか
モデルの推測によるものかを切り分けるためです。

## 実行

```sh
export ANTHROPIC_API_KEY=...   # PowerShell は $env:ANTHROPIC_API_KEY = "..."
go run .
```

`out/<条件名>/` に、渡したプロンプト（`prompt.txt`）と各回の出力
（`run-01.go.txt` 〜 `run-05.go.txt`）が保存されます。

## 記事執筆時の結果

`out/` に置いてあるのは、記事を書いた時点の出力です。要点は3つです。

- 10回とも遷移表に従った。`CanTransitionTo` を指示なしで呼び、表にない遷移は足さなかった
- 名前が `StatusRejected` でも、10回とも差戻しから申請中へ戻せる実装になった
- table-only では承認ステップが登場しない。table-steps では5回とも実装され、
  差戻し時の承認リセットの位置が `Reject` で4回、`Submit` で1回に割れた

生成結果なので、同じプロンプトでも実行時期やモデルによって変わります。
再現しない場合があることを前提に読んでください。