# frontend

## Project setup

npmで必要なパッケージ周りを入れます。npmがない場合は、そちらから入れてください。（手順外）

```
npm install
```

(開発環境の内容ですが...)
vueのデフォルトWebサーバは localhost:8080 で動作します。
一方で、バックエンドのアプリは localhost:5000 で動作するため、CORSを突破する必要があります。
そこで、フロントエンド開発時はflask-corsのインストール＋修正いれてください（Pythonの場合）

```
pip install -U flask-cors
```

```
from flask import Flask
from flask_cors import CORS

app = Flask(__name__)
CORS(app, supports_credentials=True, origins='http://localhost:8080')
```

### Compiles and hot-reloads for development
```
npm run serve
```

で確認できます。 http://localhost:8080/ にて。

なお、APIサーバをローカルで動作時は
main.js の `Vue.prototype.$apiUrl = '';` を
`Vue.prototype.$apiUrl = 'http://localhost:5000';` などに変更する。

### Compiles and minifies for production
```
# minify版 (たぶん使わない)
npm run build

# 非minify版（使う）
npm run build-dev
```

### N-ISUCON用デプロイ

```
deploy_to_backend.sh
```

で、dist配下をbackendのappが参照するpublicにコピーする。

### Run your tests
```
npm run test
```

未実装。

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).
