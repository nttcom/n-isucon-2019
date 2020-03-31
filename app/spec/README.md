## 仕様の標記について

仕様の記載には、swaggerを用いている。

## 編集方法

Swagger Editorを用いて編集すると、プレビューも同時確認できる。
Dockerから起動して、ブラウザでアクセスすればいい。

```
docker run -d -p 80:8080 swaggerapi/swagger-editor
```

## HTML出力の準備

yamlのままでもAPI仕様は確認可能だが、視認性が悪い。
Swagger Editor自体には、HTMLを出力する機能がない。
そこで、YAMLを解釈してhtmlを出力して視認性を上げる別のツールを用いる。

ツールには、 https://github.com/Redocly/redoc を用いる。
インストール方法は公式を参照のこと。

## HTMLの出力方法

```
// swagger.yamlを読み込んでspec.htmlを吐き出す
$ redoc-cli bundle swagger.yaml -o spec.html
```

