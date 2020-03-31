# n-isucon-2019
## これはなに
2019年9月に開催した、N-ISUCON 2019で実際に使用したプログラムたちです。競技当日の雰囲気や、それを支えた裏側のお話はこちらをご覧ください。

* [社内ISUCON『N-ISUCON』を開催しました | NTT Communications Developer Portal](https://developer.ntt.com/ja/blog/31e9fb86-7309-4145-a72b-0ea5f17e3f5a)
* [N-ISUCONを支えた技術 - Qiita](https://qiita.com/naosuke2dx/items/d395cd2a52e1dd5fde62)
* [社内ISUCON(N-ISUCON)におけるベンチマーカにおいて考慮・実装で工夫したこと - Qiita](https://qiita.com/iwashi86/items/126b6a7b0ae5ccf6f43e)


## VM構成
本戦では、競技VM 3台 (2Core 2GB) で開催しました。
また、ベンチマーカとしてのVM (16Core) を複数台用意し、実際に負荷をかけました。

このリポジトリのコードを用いて、競技VMをデフォルト1台（最大3台）、ベンチマーカVMを最大1台 (2Core 2GB) 起動して、実際の競技を体験することができます。

## 競技VMの作成
### 動作に必要な環境
- コンピュータ（いずれか一つが必要）
  - Windows 8以降が動作するコンピュータ
  - macOS 10.13以降が動作するMac
- ソフトウェア（すべて必須）
  - VirtualBox (https://www.virtualbox.org/)
  - Vagrant (https://www.vagrantup.com/)

動作確認はWindows 10 1909およびmacOS 10.15が動作するコンピュータで行っています。
VirtualBoxが動作するLinuxディストリビューションが入ったコンピュータでも動作すると考えられますが、確認はしていません。

Windows環境の場合、Hyper-Vが現に動作している環境（Docker for Windowsをご利用の方は該当すると考えられます）では
VirtualBoxの仮想マシンが起動しないため、起動に失敗します。
Hyper-V機能を削除するか、一時的に無効にしてご利用ください。

### seedデータのダウンロード
このリポジトリのルートにある `seed.sql.7z` ファイルを7-Zip等で展開してください。
展開後のファイルサイズは300MB以上ありますのでご注意ください。

生成された `seed.sql` ファイルを `infra/production/roles/deploy_app/files` に配置してください。

### VM起動
競技VMはVagrantコマンドを使って起動できます。

```
vagrant up
```

デフォルトでは、競技VMが1台、ベンチマーカが1台起動します。
また、初回起動時にAnsibleが走り、自動で環境が作られます。相当の時間がかかるのでお待ちください。

IPアドレスの衝突が発生した場合は、Vagrantfileを修正してください。
デフォルトでは、192.168.33.11-13, 101 にVMが起動する設定になっています

## 競技方法
### レギュレーション
本戦のレギュレーションのうち、技術に関連する部分は下記のとおりです。
- 競技時間は6時間
- 参加者は与えられたソフトウェア、もしくは自ら競技時間中に実装したソフトウェアを用います
- 運営から、高速化対象のソフトウェアとして複数言語で実装したWebアプリケーションが与えられます
  - 各言語ごとの実装において、性能が同一であることは保証されません
- 次のことは行っても差し支えありません
  - 各サーバのソフトウェアの入れ替え
  - 設定の変更
  - ソフトウェアのソースコードの変更および入れ替え
- 次のことは禁止します
  - 与えられたサーバ以外の外部リソースを利用する行為（例：自身で持ち込んだ計算機への処理の委譲など）
    - （注：当日は各チームに対してVMが3台割り当てられており、その内側で処理を完結させてください、というニュアンスです）
  - 高速化対象のソフトウェアのうち、次の点を変更すること
    - アクセス先のURI
    - レスポンス (HTML) のDOM構造、JavaScript/CSSファイルの内容（ただし、表示に影響しない範囲での修正は許可される）
    - 画像および動画等のメディアファイルの内容
- 次の点に留意してください
  - サーバ再起動後に全てのアプリケーションコードが正常に動作する状態にすること
  - ベンチマーク実行時にアプリケーションに書き込まれたデータがサーバ再起動後も取得できること

## 競技のやり方
### app
VMへは、 `vagrant ssh` を利用してログインしてください。
また、競技用のコードは `isucon` ユーザで管理されているので、ユーザの切り替えも実施してください。

```
vagrant ssh app1
sudo su - isucon
```

home dirに、 `app/` ディレクトリがあり、その中に言語別の実装が入っています。
初期状態では、競技アプリは起動していないので下記コマンドで起動をしてください。

```
sudo systemctl niita_${言語}
```

また、 `app1` の 80/tcp はVirtualBoxホストの8001/tcpにフォワードされます。
ホストのブラウザから `http://localhost:8001` にアクセスすることで、ブラウザでアプリケーションの動作を確認できます。
詳しくは `Vagrantfile` を確認してください。

### bench
appと同様に、 `vagrant ssh` を利用して接続してください。
ベンチマーカは `isucon_admin` ユーザで管理されているので、ユーザの切り替えも実施してください。

```
vagrant ssh bench
sudo su - isucon_admin
```

`~/bench/scenario` に移動すると、ベンチマークのスクリプト群が配置されています。
ベンチマークは、下記コマンドで実行できます。

```
./run.sh
```

run.shは、192.168.33.11 (app1) に対してベンチマークを実施します。
3段階のフローがあり、実際のスコアの測定は第3段階になります。

1. DB初期化
2. API整合性チェック
3. ベンチマーク

3段階目が終了後、標準出力にスコアが出力されます。

```
STEP3: Succeeded in benchmark for 192.168.33.11. Caluculating score...
 ________________________________________________________________________
< Congratulations! Score: 239, Success requests: 119, Failed requests: 0 >
 ------------------------------------------------------------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||

```

# その他
## 免責事項
このリポジトリ上のファイルはすべて現状有姿で提供され、NTTコミュニケーションズ株式会社は明示的・暗黙的に関わらずあらゆる保証を提供いたしません。
詳しくは [LICENSE](LICENSE) をご覧ください。

## Pull requestについて
本戦に利用した環境をそのまま提供することを目的としているため、このリポジトリに対するpull requestには応答いたしません。ご了承ください。

This repository is under NO maintenance.

## トラブルの報告
動作に必要な環境を準備の上、手順通りに操作しても環境が起動できないといったトラブルに関しては、
上記にかかわらずIssueでの報告を受け付けます。なお回答に時間を要する場合がある点をご了承ください。

ただしVagrantやVirtualBoxのインストール方法、Hyper-Vを削除・無効化する方法など必要な環境を準備する方法にはお答えできません。

## 著作権表示
Copyright (c) 2020 NTT Communications Corporation
