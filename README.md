# ringi — Go × DDD 学習用サンプル

Go でドメイン駆動設計(DDD)を実践するための学習用サンプルコードです。
「申請・承認」という小さなドメインを題材に、**パッケージ構成でレイヤードアーキテクチャを表現する**ことを目指しています。

Qiita 連載のコード置き場です。設計判断の背景や解説は記事をご覧ください。

## 各回の記事とコード

記事執筆時点のコードは、回ごとのタグで参照できます。main は常に最新回の状態です。

### 連載1: GoでDDD

| 回 | 記事 | コード |
| --- | --- | --- |
| 第1回 | [GoでDDDを始めたら最初にぶつかった「クラスがない」問題と、パッケージ構成という答え](https://qiita.com/shinchi-pmtech/items/f6748431b93969f5b526) | [article-01](https://github.com/shinchi-pmtech/ringi/tree/article-01) |
| 第2回 | [差戻しされた申請は再提出できるのか。Goの値オブジェクトで状態遷移を型に落とす](https://qiita.com/shinchi-pmtech/items/66e0ee47136eb3a9d720) | [article-02](https://github.com/shinchi-pmtech/ringi/tree/article-02) |
| 第3回 | [インメモリをSQLiteに差し替えても、usecaseは無傷でいられるのか。Goのリポジトリパターンで確かめる依存性逆転](https://qiita.com/shinchi-pmtech/items/d8937ab75348fbf63904) | [article-03](https://github.com/shinchi-pmtech/ringi/tree/article-03) |
| 第4回 | [承認ルートは申請の中に置くべきか、外に出すべきか。Goで多段承認を作りながら集約の境界を引く](https://qiita.com/shinchi-pmtech/items/59a665b68909ca55b87c) | [article-04](https://github.com/shinchi-pmtech/ringi/tree/article-04) |
| 第5回 | [「10万円以上は部長承認」をどこに書くか。Goのドメインサービスと、使いすぎない線引き](https://qiita.com/shinchi-pmtech/items/9046380853c316b68bf7) | [article-05](https://github.com/shinchi-pmtech/ringi/tree/article-05) |

### 連載2: ユビキタス言語をAIに渡す

連載1で整理した語彙を、AIに渡せるかを試す連載です。実験コードは [`ontology/`](./ontology) 以下にあります。

| 回 | 記事 | コード |
| --- | --- | --- |
| 第1回 | [差戻された申請は、もう一度出せますか？　AIの答えは「出せない」](https://qiita.com/shinchi-pmtech/items/fea33ad995978944a432) | [ai-01](https://github.com/shinchi-pmtech/ringi/tree/ai-01) |

## 動かし方

```bash
git clone https://github.com/shinchi-pmtech/ringi.git
cd ringi
go run .
```

申請金額によって承認ルートが変わる様子が動きます。実行するとカレントディレクトリに `ringi.db`(SQLiteのデータファイル)が作られます。

```
マウス購入         5,000円      draft      ○kacho
開発端末の購入     150,000円    draft      ○kacho → ○bucho

マウス購入         5,000円      approved   ●kacho

組織図にない申請者: 承認者が見つかりません: 申請者 unknown の組織情報がありません
```

`●` は承認済み、`○` は未承認です。10万円未満は課長のみ、10万円以上は課長と部長の2段承認になります。

テストも用意しています。

```bash
go test ./...
```

動作確認済み環境:Go 1.22

## パッケージ構成

```
.
├── domain
│   └── application          # 「申請」集約(エンティティ・値オブジェクト・承認ステップ・ドメインサービス・interface)
├── usecase                  # 「申請する」「承認する」などのユースケース
├── infrastructure
│   ├── organization         # 組織図の参照(ApproverResolver の実装)
│   └── persistence          # リポジトリの実装(SQLite / インメモリ)
├── presentation
│   └── handler              # HTTPハンドラ
└── ontology
    └── experiment           # 連載2の実験コード(AIに語彙を渡す試み)
```

依存の向きはすべて `domain` に向かいます。`domain` は標準ライブラリ以外の何にも依存しません。

## 設計のポイント

- **ビジネスルールは domain に置く**。「承認者は申請者と同一人物であってはならない」は承認ルートの検証で、「順番を飛ばせない」は `Application.Approve()` で守られます
- **interface は使う側(domain)に定義する**。Go の慣習に従うことで、依存性逆転が自然に実現します
- **フィールドは非公開**。状態変更の唯一の手段をメソッドに限定し、不変条件をコンパイラに守らせます
- **状態遷移のルールは遷移表に集約する**。「どの状態からどこへ動けるか」は `status.go` の遷移表 1 箇所にあり、テストは表の写しで書けます
- **生成と復元は別物**。DBから読み戻した値は `NewStatus` の検証を通してから `Reconstruct` で組み立てます。不正なデータはドメインに入る前に弾かれます
- **集約は丸ごと扱う**。承認ステップは申請集約の一部なので、専用のリポジトリを持たず、申請と同じトランザクションで保存されます。外へはスライスの複製を返し、書き換えを防ぎます
- **どのモデルにも属さないルールはドメインサービスへ**。「10万円以上は部長承認」の判定は金額・組織のルール・組織図をまたぐため、`DecideApprovalRoute` として独立させています。状態を持たないので関数です

## 連載の状況

連載1は全5回で、エンティティ・値オブジェクト・リポジトリ・集約・ドメインサービスという戦術的DDDの主要な部品が一通り揃い、一区切りとしました。

連載2では、そこで整理したユビキタス言語をAIに渡す話(オントロジー、MCP)を扱っています。コードは引き続きこのリポジトリに追加していきます。

## ライセンス

MIT License([LICENSE](./LICENSE) を参照)

## 免責事項

- 本リポジトリは学習用のサンプルです。エラーハンドリングや並行性の考慮は説明のため最小限にしています
- 本コードを利用したことによるいかなる損害についても、作者は責任を負いません。利用は自己責任でお願いします
- 記載内容は個人の学習に基づくものであり、所属組織の見解や品質基準を表すものではありません