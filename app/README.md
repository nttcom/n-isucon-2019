# 各言語のappの動かし方

## python appの動かし方

```
// 開発向けにVolumeマウントして使う
$ docker-compose -f docker-compose-dev.yaml stop && docker-compose -f docker-compose-dev.yaml rm && docker-compose -f docker-compose-dev.yaml up

//// 上記でマウントしたものを実行する
$ docker-compose -f docker-compose-dev.yaml exec app ./endpoint.sh

// Volumeマウントしない場合
$ docker-compose stop && docker-compose rm && docker-compose up --build
```