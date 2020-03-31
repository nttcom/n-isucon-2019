import http from "k6/http";
import { check, group } from "k6";

//////// 変数・定数定義 ////////

// Dev or Prodのアクセス先定義
let targetUrl = 'http://192.168.33.11';


// 1-20000 のユーザ名・PWを格納しており、既存ユーザへのアクセスに使う
const users = JSON.parse(open("./user_pass.json"));

// const binFile = open("./icons/100.png", "b");
const filenames = [...Array(10).keys()].map(i => `${i}.png`);
let iconImages = [];
filenames.forEach(filename => {
    iconImages.push(open(`./icons/${filename}`, 'b'));
})

// headerはこれぐらいしか使わないので共通で定義
const params = { headers: { "Content-Type": "application/json" } };

// POST用のアイコンを読み込む


// 初回の記事数
// const num_initial_item = 20000;
// const item_per_page = 10;
const num_initial_item = 10;
const item_per_page = 10;

//////// 関数定義 ////////
const getRand = (min, max) => {
    return Math.floor(Math.random() * (max - min + 1) + min);
};

const makeRandomString = (length = 8) => {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
};

//////// ここからテストシナリオ ////////

export default function () {

    group('静的ファイル', () => {
        group('GET index.html/app.js/favicon.ico/logo.png', () => {
            const responses = http.batch([
                ["GET", `${targetUrl}/index.html`],
                ["GET", `${targetUrl}/app.js`],
                ["GET", `${targetUrl}/favicon.ico`],
                ["GET", `${targetUrl}/img/logo.png`],
            ]);

            check(responses[0], {
                "index html status was 200": res => res.status === 200,
            });
            check(responses[1], {
                "app.js status was 200": res => res.status === 200,
            });
            check(responses[2], {
                "favicon.ico status was 200": res => res.status === 200,
            });
            check(responses[3], {
                "logo.png status was 200": res => res.status === 200,
            });
        });
    });

    group('未ログイン時の / へアクセス時の処理', () => {
        let items;

        // 新着順
        group('GET /items', () => {
            const resItems = http.get(`${targetUrl}/items?page=0`);
            check(resItems, {
                "items status was 200": r => r.status === 200
            });
            items = resItems.json('items');
        });

        group('GET /items:id', () => {
            items.forEach(item => {
                const resItem = http.get(`${targetUrl}/items/${item.id}`);
                check(resItem, {
                    "item status was 200": r => r.status === 200
                });
            });
        });

        group('GET /users/:username/icon', () => {
            items.forEach(item => {
                const resItem = http.get(`${targetUrl}/users/${item.username}/icon`);
                check(resItem, {
                    "icon status was 200": r => r.status === 200
                });
            });
        });

        // ここから良いね順
        group('GET /items like order', () => {
            const resItems = http.get(`${targetUrl}/items?sort=like&page=0`);
            check(resItems, {
                "items status was 200": r => r.status === 200
            });
            items = resItems.json('items');
        });

        group('GET /items:id', () => {
            items.forEach(item => {
                const resItem = http.get(`${targetUrl}/items/${item.id}`);
                check(resItem, {
                    "item status was 200": r => r.status === 200
                });
            });
        });

        group('GET /users/:username/icon', () => {
            items.forEach(item => {
                const resItem = http.get(`${targetUrl}/users/${item.username}/icon`);
                check(resItem, {
                    "icon status was 200": r => r.status === 200
                });
            });
        });
    });

    // 新規ユーザ作成 or 既存ユーザを抽出
    let username, password;
    group('新規ユーザ作成 or 既存ユーザのユーザ名・PWを抽出', () => {
        // if (Math.random() >= 0.5) {
        if (true) { // DeBuG: DBシードを突っ込むまでは、trueにしとく
            group('POST /users', () => {
                // 新規ユーザ作成
                username = makeRandomString(); //奇跡的な確率で既存ユーザと衝突するが、それは誤差みたいなもん
                password = makeRandomString();
                const body = {
                    username: username,
                    password: password
                };
                const resPostUser = http.post(`${targetUrl}/users`, JSON.stringify(body), params);
                check(resPostUser, {
                    "status is 201": r => r.status === 201
                });
            });
        } else {
            const randomIndex = Math.floor(Math.random() * users.length);
            username = users[randomIndex].username;
            password = users[randomIndex].password;
        }
    })

    group('ログイン', () => {
        // 1/2の確率でログイン失敗
        if (Math.random() >= 0.5) {
            group('POST /signin with wrong param', () => {
                const body = {
                    username: makeRandomString(),
                    password: makeRandomString(),
                };
                const resSignin = http.post(`${targetUrl}/signin`, JSON.stringify(body), params);
                check(resSignin, {
                    "login w/ wrong param should be 401": r => r.status === 401
                });
            });
        }

        // ログイン成功
        group('POST /signin with right param', () => {
            const body = {
                username: username,
                password: password
            };
            const resSignin = http.post(`${targetUrl}/signin`, JSON.stringify(body), params);
            check(resSignin, {
                "login w/ right param should be 200": r => r.status === 200
            });
        });
    });

    group('新規記事', () => {
        let item_id;
        group("投稿", () => {
            let item_title, item_body;
            group('POST /items', () => {
                item_title = makeRandomString(20);
                item_body = makeRandomString(400);
                const body = {
                    title: item_title,
                    body: item_body,
                };
                const resPostItem = http.post(`${targetUrl}/items`, JSON.stringify(body), params);
                item_id = resPostItem.json('id');

                check(resPostItem, {
                    "post items w/ right param should be 201": r => r.status === 201
                });
            });

            // 投稿された記事があるか確認する。ブラウザ同様、付随していいねとコメントを取得する。
            group('GET /item/:id', () => {
                // 記事を取得
                const resPostItem = http.get(`${targetUrl}/items/${item_id}`);
                check(resPostItem, {
                    "Get created item should be 200": r => r.status === 200,
                    "Title should contain the body previously posted": r => r.json().title === item_title,
                    "Body should contain the body previously posted": r => r.json().body === item_body,
                });

                // いいねを取得
                const resLikes = http.get(`${targetUrl}/items/${item_id}/likes`);
                check(resLikes, {
                    "Get likes should be 200": r => r.status === 200,
                    "Like count should be 0 for newly posted item": r => r.json().like_count === 0,
                });

                // コメントを取得
                const resComments = http.get(`${targetUrl}/items/${item_id}/comments`);
                check(resComments, {
                    "Get comments should be 200": r => r.status === 200,
                });
            });

        });

        let patched_item_title, patched_item_body;
        group('投稿記事の編集', () => {
            // 投稿した記事を編集する
            group('PATCH /items', () => {
                patched_item_title = makeRandomString(20);
                patched_item_body = makeRandomString(400);
                const body = {
                    title: patched_item_title,
                    body: patched_item_body,
                };
                const resPatchItem = http.patch(`${targetUrl}/items/${item_id}`, JSON.stringify(body), params);
                check(resPatchItem, {
                    "patch item w/ right param should be 200": r => r.status === 200
                });
            });

            // 編集された記事があるか確認する。ブラウザ同様、付随していいねとコメントを取得する。
            group('GET /item/:id', () => {
                // 記事を取得
                const resGetItem = http.get(`${targetUrl}/items/${item_id}`);
                check(resGetItem, {
                    "Get patched item should be 200": r => r.status === 200,
                    "Patched title should contain the body previously posted": r => r.json().title === patched_item_title,
                    "Patched body should contain the body previously posted": r => r.json().body === patched_item_body,
                });

                // いいねを取得
                const resLikes = http.get(`${targetUrl}/items/${item_id}/likes`);
                check(resLikes, {
                    "Get likes should be 200": r => r.status === 200,
                    "Like count should be 0 for newly posted item": r => r.json().like_count === 0,
                });

                // コメントを取得
                const resComments = http.get(`${targetUrl}/items/${item_id}/comments`);
                check(resComments, {
                    "Get comments should be 200": r => r.status === 200,
                });
            });
        });
    });

    // ランダムで 1-20000の間の記事を取得して、既存記事にいいね作成＆削除・コメント作成＆削除する。
    // これにより、既存レコードを抜いていないかチェックする。
    group('既存の記事', () => {
        const seedItemId = getRand(1, num_initial_item);
        let likes_count;

        // 記事を取得
        group('GET /item/:id', () => {
            const resGetItem = http.get(`${targetUrl}/items/${seedItemId}`);
            check(resGetItem, {
                "Get patched item should be 200": r => r.status === 200,
            });
        });

        group('いいね', () => {
            // いいねを取得
            group('GET /item/:id/likes', () => {
                const resLikes = http.get(`${targetUrl}/items/${seedItemId}/likes`);
                check(resLikes, {
                    "Get likes should be 200": r => r.status === 200,
                });

                likes_count = resLikes.json().like_count;
            });

            // いいねを追加
            group('POST /item/:id/likes', () => {
                const resLikes = http.post(`${targetUrl}/items/${seedItemId}/likes`);
                check(resLikes, {
                    "POST likes should be 204": r => r.status === 204,
                });
            })

            // いいねが増えていることを確認
            group('GET /item/:id/likes', () => {
                const resLikes = http.get(`${targetUrl}/items/${seedItemId}/likes`);
                check(resLikes, {
                    "Get likes body should contain logged in username":
                        r => r.json().likes.split(',').includes(username),
                    "Get likes body should increment likes_count":
                        r => r.json().like_count === likes_count + 1,
                });
            });

            // いいねを消す
            group('DEL /item/:id/likes', () => {
                const resLikes = http.del(`${targetUrl}/items/${seedItemId}/likes`);
                check(resLikes, {
                    "POST likes should be 204": r => r.status === 204,
                });
            })

            // いいねが減っていることを確認
            group('GET /item/:id/likes', () => {
                const resLikes = http.get(`${targetUrl}/items/${seedItemId}/likes`);
                check(resLikes, {
                    "Get likes body should contain logged in username":
                        r => !(r.json().likes.split(',').includes(username)),
                    "Get likes body should increment likes_count":
                        r => r.json().like_count === likes_count,
                });
            });
        });

        group('コメント', () => {
            // コメントを取得
            let comments;
            group('GET /items/:id/comments', () => {
                const resComments = http.get(`${targetUrl}/items/${seedItemId}/comments`);
                if (resComments.status !== 404) {
                    comments = resComments.json()
                }
                check(resComments, {
                    "Get comments should be 200 or 404": r => r.status === 200 || r.status === 404,
                });
            });

            // コメント投稿
            let comment, comment_id;
            group('POST /items/:id/comments', () => {
                comment = makeRandomString(100);
                const body = {
                    comment: comment,
                };
                const resComments = http.post(`${targetUrl}/items/${seedItemId}/comments`, JSON.stringify(body), params);
                comment_id = resComments.json().comment_id;
                check(resComments, {
                    "Post comments should be 201": r => r.status === 201,
                    "Posted comment should have my comment": r => r.json().comment === comment,
                });
            });

            // コメントを取得し、投稿したコメントが含まれているか確認
            group('GET /items/:id/comments', () => {
                const resComments = http.get(`${targetUrl}/items/${seedItemId}/comments`);
                const comments_array = resComments.json().comments.map(c => c.comment);
                check(resComments, {
                    "Comments should be 200": r => r.status === 200,
                    "Comments should include my comment": r => comments_array.includes(comment),
                });
            });

            // コメントを消す
            group('DEL /items/:id/comments/:comment_id', () => {
                const resComments = http.del(`${targetUrl}/items/${seedItemId}/comments/${comment_id}`);
                check(resComments, {
                    "Deleted Comment should be 204": r => r.status === 204,
                });
            });

            // 削除したコメントが含まれていないか確認
            group('GET /items/:id/comments', () => {
                const resComments = http.get(`${targetUrl}/items/${seedItemId}/comments`);
                check(resComments, {
                    "Get comments should be 200 or 404": r => r.status === 200 || r.status === 404,
                });
                if (resComments.status !== 404) {
                    const comments_array = resComments.json().comments.map(c => {
                        if (c.comment) {
                            return c.comment
                        }
                    });
                    check(resComments, {
                        "Comments should include my comment": r => !(comments_array.includes(comment)),
                    });
                }
            });
        });
    });

    // ここからユーザ情報を変更する
    group('ユーザ情報', () => {
        // 変更前のユーザ情報を取得
        group('GET /users/:username', () => {
            const resUser = http.get(`${targetUrl}/users/${username}`);
            check(resUser, {
                "Get user should be 200": r => r.status === 200,
            });
        });

        // アイコンを変更する
        let iconLength;
        group('POST /users/:username/icon', () => {
            const data = {
                iconimage: http.file(iconImages[0], '0.png', 'image/png'),
            };
            iconLength = data.length;
            const resIcon = http.post(`${targetUrl}/users/${username}/icon`, data);
            check(resIcon, {
                "POST icon should be 201": r => r.status === 201,
            });
        });

        // アイコンを取得する
        group('GET /users/:username/icon', () => {
            const resIcon = http.get(`${targetUrl}/users/${username}/icon`);
            check(resIcon, {
                "GET icon should be 200": r => r.status === 200,
            });
        });

        // ユーザ情報(パスワード)を変更
        let newPassword;
        group('PATCH /users/:username', () => {
            newPassword = makeRandomString();
            const body = {
                password: newPassword,
            }
            const resUser = http.patch(`${targetUrl}/users/${username}`, JSON.stringify(body), params);
            check(resUser, {
                "Get user should be 200": r => r.status === 200,
            });
        });

        // 変更後のユーザ情報を取得
        group('GET /users/:username', () => {
            const resUser = http.get(`${targetUrl}/users/${username}`);
            check(resUser, {
                "Get user should be 200": r => r.status === 200,
            });
        });
    });

    // ログアウトして処理おしまい
    group('ログアウト', () => {
        group('GET /signout', () => {
            const resUser = http.get(`${targetUrl}/signout`);
            check(resUser, {
                "Get user should be 204": r => r.status === 204,
            });
        });
    });
};
