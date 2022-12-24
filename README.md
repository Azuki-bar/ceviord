# Ceviord
[![Go Report Card](https://goreportcard.com/badge/github.com/azuki-bar/ceviord)](https://goreportcard.com/report/github.com/azuki-bar/ceviord)
[![Go Reference](https://pkg.go.dev/badge/github.com/azuki-bar/ceviord.svg)](https://pkg.go.dev/github.com/azuki-bar/ceviord)

Discord 上のテキストチャンネルに投稿したメッセージを読みあげる bot です。[CeVIGO](https://github.com/gotti/cevigo) を用いて CeVIO AI の API を利用しています。

## How to use

Ceviord はコンテナイメージを提供しています。もし Docker や Kubernetes が使える環境であればコンテナイメージを用いた運用をおすすめします。

### コンテナイメージを用いる方法

CeVIGO が動作しているサーバに対して gRPC にて通信を行います。別途 CeVIO AI をインストールしているマシン上で CeVIGO を動作させることが必要です。

[環境変数ドキュメント](./docs/config.md)に従って環境変数を設定してください。

[docker-compose.yml](./docker-compose.yml)も参考にしてください。

MySQL を変換辞書の保存先として使用します。用意してください。

### go のバイナリを用いる方法

`make build && ./ceviord`で実行できます。実行時のディレクトリにある `auth.yaml`を読み込むことができます。また環境変数で値を入れることもできます。詳しくは[設定ドキュメント](./docs/config.md)を参照してください。

#### Requirement

- [Opus development library](https://github.com/layeh/gopus#requirements)
- [CeVIO.Talk.RemoteService2.Talker2](https://cevio.jp/)
- mySQL

## コマンドリファレンス

[ceviord commands reference](./docs/cmd.md)
