# openai_cost_exporter

OpenAI PlatformとChatGPT EnterpriseのコストをPrometheusへ公開する、Go製exporterの開発用初期構成です。

**現在はHTTPサーバーと開発基盤のみです。両サービスの実コスト取得・認証・コストメトリクスの公開は未実装です。** 起動にAPIキーは不要で、OpenAIへのリクエストも行いません。

## 現在の機能

- `/metrics`: `openai_cost_exporter_build_info{version="..."}`、Go runtime、プロセスメトリクス
- `/healthz`: HTTPサーバーの生存確認（`200 ok`）。課金APIへの接続可否やデータの鮮度は保証しません。
- `--web.listen-address`（既定: `127.0.0.1:9464`）、`--version`、`--help`
- SIGINT / SIGTERMによるgraceful shutdown
- PlatformとChatGPT Enterpriseを分離する`internal/provider.Provider`インターフェース

未取得のコストを`0`として公開することや、ダミーの請求額を公開することはありません。

## 開発環境

- Go 1.26以上
- GNU Make
- race detector用のCコンパイラー（Linuxではgccなど）
- Docker（コンテナー利用時のみ）

```sh
make setup
make check   # gofmt検査、go vet、race付きテスト、ビルド
make run
```

別のターミナルから確認します。

```sh
curl --fail http://127.0.0.1:9464/healthz
curl --fail http://127.0.0.1:9464/metrics
```

ビルドしたバイナリの使用例:

```sh
make build VERSION=local
./bin/openai_cost_exporter --version
./bin/openai_cost_exporter --web.listen-address=127.0.0.1:9464
```

### Prometheus設定例

exporterとPrometheusが同じホストで動く場合:

```yaml
scrape_configs:
  - job_name: openai_cost_exporter
    scrape_interval: 60s
    static_configs:
      - targets: ['127.0.0.1:9464']
```

別ホストや別コンテナーからscrapeする場合は、接続先アドレスとlistenアドレスを変更してください。

### Docker

```sh
make docker VERSION=local
docker run --rm --name openai-cost-exporter \
  -p 127.0.0.1:9464:9464 openai-cost-exporter:local
```

イメージは非rootユーザーで起動します。Docker内では`:9464`でlistenし、上の例ではホストのloopbackにのみ公開します。認証・TLSは未実装なので、インターネットへ直接公開しないでください。外部公開が必要な場合は認証付きリバースプロキシ等を使用してください。

## 今後の実装方針

### OpenAI Platform

- [公式Costs API](https://developers.openai.com/api/reference/resources/organization/subresources/costs/methods/get)の`GET /v1/organization/costs`を使用する想定です。
- 通常の推論用APIキーではなく、適切な権限の**Admin API key**が必要です。
- 日次バケット、ページネーション、project / line item別集計、通貨を扱うアダプターを追加します。

### ChatGPT Enterprise

- [公式説明](https://help.openai.com/en/articles/9039756-managing-billing-for-chatgpt-and-the-api-platform)のとおり、ChatGPTとAPI Platformの請求システムは別です。
- Enterprise契約で取得できる課金・請求・利用データの仕様とアクセス権を確認し、専用アダプターを実装します。取得経路はまだ確定していません。
- Platform Costs APIからEnterpriseの請求額を取得できるとは仮定しません。利用量や監査ログも請求額そのものとは扱いません。

### 共通の境界

`internal/provider`に`platform` / `chatgpt_enterprise`のサービス識別子、期間・通貨・金額等を持つ`Cost`型、`Provider`インターフェースを定義しています。期間はUTC基準の`[start, end)`として扱う想定です。各サービスの認証・API仕様をここから分離します。実装済みのproviderはまだありません。

今後のコストメトリクスでは、訂正・返金で減少し得る期間集計値をCounterや`_total`にせずGaugeとして扱い、集計期間・取得成否・データ鮮度を明示する方針です。具体的なメトリクス名とラベルはデータソースの仕様確認後に確定します。APIキーや個人情報をラベルに含めません。

## 構成

```text
cmd/openai_cost_exporter/  CLIとHTTPサーバーのライフサイクル
internal/exporter/        PrometheusレジストリとHTTPハンドラー
internal/provider/        サービスごとのコスト取得インターフェース
.github/workflows/        Go検査・テスト・ビルドとDockerビルド
```

`make fmt`で整形し、`make check`を通してから変更を提出してください。APIキーや請求データをコミットしないでください。
