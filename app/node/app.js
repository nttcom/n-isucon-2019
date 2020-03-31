const fastify = require('fastify')({
    logger: true
});

const fastifyStatic = require("fastify-static");
const fastifyMysql = require("fastify-mysql");
const fastifyCookie = require("fastify-cookie");
const fastifyMultipart = require('fastify-multipart');

const path = require("path");
const child_process = require("child_process");
const util = require("util");
const fs = require('fs');
const execFile = util.promisify(child_process.execFile);
const execCommand = util.promisify(child_process.exec);

const myUtil = require('./utils/utility');

//// fastify registration

fastify.register(fastifyStatic, {
    root: path.join(__dirname, "public"),
});

fastify.register(fastifyMysql, {
    host: process.env.MYSQL_HOST,
    port: 3306,
    user: process.env.MYSQL_USER,
    password: process.env.MYSQL_PASSWORD,
    database: process.env.MYSQL_DATABASE,
    promise: true,
    connectionLimit: 1024,
});

fastify.register(fastifyCookie);
fastify.register(fastifyMultipart);

//// Helper functions

async function getDbConn() {
    return fastify.mysql.getConnection();
};

async function getPasswordhash(salt, password) {
    let str = password + salt;
    for (const _ of Array(1000)) {
        const cmd = 'echo -n ' + str + ' | openssl sha256';
        const res = await execCommand(cmd);
        str = res.stdout.split(' ')[1].trim();
    }
    return str;
};

function isEmpty(property) {
    return !(property || property === '');
};

function loginRequired(request, reply, done) {
    if (!request.cookies.username) {
        return reply.type("application/json").code(401).send("");
    }
    done();
};

//// Routes

fastify.post('/users', async (request, reply) => {
    if (isEmpty(request.body.username) || isEmpty(request.body.password)) {
        return reply.type("application/json").code(400).send("");
    }

    const username = request.body.username;
    const password = request.body.password;

    const conn = await getDbConn();

    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (rows.length > 0) {
        return reply.type("application/json").code(409).send("");
    }

    const salt = myUtil.getSalt();
    const hash = await getPasswordhash(salt, password);
    const currentTime = new Date();

    let newUser;
    try {
        const insertSql = fastify.mysql.format('INSERT INTO users (username, password_hash, salt, created_at, updated_at) VALUES (?, ?, ?, ?, ?);',
            [username, hash, salt, currentTime, currentTime]);
        const [insertedRows] = await conn.query(insertSql);
        const lastUserId = insertedRows.insertId;
        fastify.log.info(insertSql);

        await conn.commit();

        newUser = {
            id: lastUserId,
            username,
            created_at: currentTime,
            updated_at: currentTime,
        };
        return reply.code(201).send(newUser);
    } catch (e) {
        return reply.type("application/json").code(500).send("");
        fastify.log.info(e);
    } finally {
        conn.release();
    }
});

fastify.get('/users/:username', async (request, reply) => {
    const conn = await getDbConn();

    const sql = fastify.mysql.format('SELECT id, username, created_at, updated_at FROM users WHERE username = ?',
        [request.params.username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);
    conn.release();

    if (rows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    return reply.code(200).send({
        id: rows[0].id,
        username: rows[0].username,
        created_at: rows[0].created_at,
        updated_at: rows[0].updated_at,
    });
});

fastify.patch('/users/:username', { preHandler: loginRequired }, async (request, reply) => {
    if (isEmpty(request.body.username) && isEmpty(request.body.password)) {
        return reply.type("application/json").code(400).send("");

    }
    const requestedUsername = request.body.username;
    const password = request.body.password;

    const conn = await getDbConn();

    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [request.params.username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (rows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (rows[0].id !== parseInt(request.cookies.user_id)) {
        return reply.type("application/json").code(403).send("");
    }

    let newUsername;
    if (isEmpty(requestedUsername)) {
        newUsername = rows[0].username;
    } else {
        const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [requestedUsername]);
        const [newUserRows] = await conn.query(sql);
        fastify.log.info(sql);

        if (newUserRows.length > 0) {
            return reply.type("application/json").code(409).send("");
        } else {
            newUsername = requestedUsername;
        }
    }

    let newSalt, newPasswordHash;
    if (password === '') {
        newSalt = rows[0].salt;
        newPasswordHash = rows[0].password_hash;
    } else {
        newSalt = myUtil.getSalt();
        newPasswordHash = await getPasswordhash(newSalt, password);
    }

    try {
        const updateSql = fastify.mysql.format('UPDATE users SET username=?, password_hash=?, salt=?, updated_at=? WHERE id=?',
            [newUsername, newPasswordHash, newSalt, new Date(), request.cookies.user_id]);
        await conn.query(updateSql);
        fastify.log.info(updateSql);
        await conn.commit();

        const selectSql = fastify.mysql.format('SELECT id, username, created_at, updated_at FROM users WHERE id= ?', [request.cookies.user_id]);
        const [updatedUserRows] = await conn.query(selectSql);
        fastify.log.info(selectSql);
        return reply.send(updatedUserRows[0]);
    } catch (e) {
        fastify.log.error(e);
    } finally {
        conn.release();
    }
});

fastify.delete('/users/:username', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();

    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [request.params.username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (rows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (parseInt(request.cookies.user_id) !== rows[0].id) {
        return reply.type("application/json").code(403).send("");
    }

    try {
        const delSql = fastify.mysql.format('DELETE FROM users WHERE id=?', [rows[0].id]);
        await conn.query(delSql);
        fastify.log.info(delSql);
        await conn.commit();
        conn.release();
    } catch(e) {
        fastify.log.error(e);
    }
    conn.release();

    return reply.setCookie("user_id", "", {
        path: "/",
        expires: new Date(0),
    }).setCookie("username", "", {
        path: "/",
        expires: new Date(0),
    }).code(204).send("");
});

fastify.post('/items', { preHandler: loginRequired }, async (request, reply) => {
    if (isEmpty(request.body.title) || isEmpty(request.body.body)) {
        return reply.type("application/json").code(400).send("");
    }

    const conn = await getDbConn();
    const currentTime = new Date();

    try {
        const insertSql = fastify.mysql.format('INSERT INTO items (user_id, title, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?);',
            [request.cookies.user_id, request.body.title, request.body.body,
                currentTime, currentTime]);
        const [insertedRows] = await conn.query(insertSql);
        const lastItemId = insertedRows.insertId;
        fastify.log.info(insertSql);
        await conn.commit();

        const selectSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [lastItemId]);
        const [selectedRows] = await conn.query(selectSql);
        fastify.log.info(selectSql);

        if (isEmpty(selectedRows[0]['likes'])) {
            selectedRows[0]['likes'] = '';
        }

        selectedRows[0]['username'] = request.cookies.username;
        delete selectedRows[0]['user_id'];

        return reply.code(201).send(selectedRows[0]);
    } catch (e) {
        fastify.log.error(e);
    } finally {
        conn.release();
    }
});

fastify.get('/items/:item_id', async (request, reply) => {
    const conn = await getDbConn();
    try {
        const selectSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
        const [selectedRows] = await conn.query(selectSql);
        fastify.log.info(selectSql);

        if (selectedRows.length === 0) {
            return reply.type("application/json").code(404).send("");
        }

        if (isEmpty(selectedRows[0]['likes'])) {
            selectedRows[0]['likes'] = '';
        }

        const selectUserSql = fastify.mysql.format('SELECT * FROM users WHERE id=?', [selectedRows[0]['user_id']]);
        const [selectedUserRows] = await conn.query(selectUserSql);
        fastify.log.info(selectUserSql);

        selectedRows[0]['username']  = selectedUserRows[0]['username'];
        delete selectedRows[0]['user_id'];

        return reply.send(selectedRows[0]);
    } catch (e) {
        fastify.log.error(e);
    } finally {
        conn.release();
    }
});

fastify.patch('/items/:item_id', { preHandler: loginRequired }, async (request, reply) => {
    if (isEmpty(request.body.title) && isEmpty(request.body.body)) {
        return reply.type("application/json").code(400).send("");
    }

    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (parseInt(request.cookies.user_id) !== selectedItemRows[0]['user_id']) {
        return reply.type("application/json").code(403).send("");
    }

    let title, body;
    if (isEmpty(request.body.title)) {
        title = selectedItemRows[0]['title'];
    } else {
        title = request.body.title;
    }

    if (isEmpty(request.body.body)) {
        body = selectedItemRows[0]['body'];
    } else {
        body = request.body.body;
    }

    const updateItemSql = fastify.mysql.format('UPDATE items SET title=?, body=?, updated_at=? WHERE id = ?',
        [title, body, new Date(), request.params.item_id]);
    await conn.query(updateItemSql);
    fastify.log.info(updateItemSql);
    await conn.commit();

    const selectItemAfterUpdateSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [updatedItemRows] = await conn.query(selectItemAfterUpdateSql);
    fastify.log.info(selectItemAfterUpdateSql);

    conn.release();

    if (isEmpty(updatedItemRows['likes'])) {
        updatedItemRows[0]['likes'] = '';
    }

    updatedItemRows[0]['username'] = request.cookies.username;
    delete updatedItemRows[0]['user_id'];

    return reply.send(updatedItemRows[0]);
});

fastify.get('/items', async (request, reply) => {
    const ITEM_LIMIT = 10;

    let page, sorting, offset;
    if (isEmpty(request.query.page)) {
        page = 0;
        offset = 0;
    } else {
        // Should validate here but we just omit this time.
        page = request.query.page;
        offset = parseInt(page) * ITEM_LIMIT;
    }
    if (!isEmpty(request.query.sort)) {
        sorting = request.query.sort;
    }

    const conn = await getDbConn();

    // let's return early if there's no items;
    const selectSql = fastify.mysql.format('SELECT * FROM items');
    const [selectedRows] = await conn.query(selectSql);
    fastify.log.info(selectSql);
    const count = selectedRows.length;

    if (selectedRows.length === 0) {
        conn.release();
        return reply.code(200).send({
            count: 0,
            items: []
        });
    }

    const selectItemsSql = fastify.mysql.format('SELECT * FROM items');
    let [selectedItemsRows] = await conn.query(selectItemsSql);
    fastify.log.info(selectItemsSql);

    if (sorting === 'like') {
        selectedItemsRows.sort((a, b) => {
            let a_likes_count, b_likes_count;
            if (isEmpty(a['likes'])) {
                a_likes_count = 0;
            } else {
                a_likes_count = a['likes'].split(',').length;
            }
            if (isEmpty(b['likes'])) {
                b_likes_count = 0;
            } else {
                b_likes_count = b['likes'].split(',').length;
            }

            if (a_likes_count > b_likes_count) return -1;
            if (a_likes_count < b_likes_count) return 1;
            return 0;
        });
        selectedItemsRows = selectedItemsRows.slice(0, ITEM_LIMIT);
    } else {
        selectedItemsRows.sort((a, b) => {
            const a_date = new Date(a['created_at']);
            const b_date = new Date(b['created_at']);

            if (a_date > b_date) return -1;
            if (a_date < b_date) return 1;
            return 0
        });
    }

    for (let [i, itemRow] of selectedItemsRows.entries()) {
        const selectUserSql = fastify.mysql.format('SELECT * FROM users WHERE id = ?', itemRow['user_id']);
        const [selectedUsersRow] = await conn.query(selectUserSql);
        selectedItemsRows[i]['username'] = selectedUsersRow[0]['username'];

        // remove unnecessary properties
        delete selectedItemsRows[i]['user_id'];
        delete selectedItemsRows[i]['likes'];
        delete selectedItemsRows[i]['body'];
        delete selectedItemsRows[i]['updated_at'];
    }

    conn.release();

    return reply.code(200).send({
        count,
        items: selectedItemsRows.slice(offset, offset + ITEM_LIMIT),
    });
});

fastify.delete('/items/:item_id', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (parseInt(request.cookies.user_id) !== selectedItemRows[0]['user_id']) {
        return reply.type("application/json").code(403).send("");
    }

    const deleteItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    await conn.query(deleteItemSql);
    fastify.log.info(deleteItemSql);
    await conn.commit();
    conn.release();

    return reply.code(204).send("");
});

fastify.post('/items/:item_id/likes', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }
    const username = request.cookies.username;
    let likesStr;
    if (isEmpty(selectedItemRows[0]['likes'])) {
        likesStr = username;
    } else {
        const currentLikeList = selectedItemRows[0]['likes'].split(',');

        if (!currentLikeList.includes(username)) {
            currentLikeList.push(username);
        }

        likesStr = currentLikeList.join(',');
    }

    const updateLikesSql = fastify.mysql.format('UPDATE items SET likes=? WHERE id=?', [likesStr, request.params.item_id]);
    await conn.query(updateLikesSql);
    fastify.log.info(updateLikesSql);
    await conn.commit();
    conn.release();

    return reply.code(204).send();
});

fastify.get('/items/:item_id/likes', async (request, reply) => {
    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);
    conn.release();

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    let likesStr, like_count;
    if (isEmpty(selectedItemRows[0]['likes'])) {
        likesStr = '';
        like_count = 0;
    } else {
        likesStr = selectedItemRows[0]['likes'];
        like_count = likesStr.split(',').length;
    }

    return reply.send({
        likes: likesStr,
        like_count,
    });
});

fastify.delete('/items/:item_id/likes', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (isEmpty(selectedItemRows[0]['likes'])) {
        return reply.type("application/json").code(404).send("");
    }

    const username = request.cookies.username;
    const currentLikesList = selectedItemRows[0]['likes'].split(',');

    if (!currentLikesList.includes(username)) {
        return reply.type("application/json").code(404).send("");
    }

    const filteredList = currentLikesList.filter(u => u !== username);

    const updateLikesSql = fastify.mysql.format('UPDATE items SET likes=? WHERE id=?',
        [filteredList.join(','), request.params.item_id]);
    await conn.query(updateLikesSql);
    fastify.log.info(updateLikesSql);
    await conn.commit();
    conn.release();

    return reply.code(204).send();
});

fastify.post('/items/:item_id/comments', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [request.params.item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    if (isEmpty(request.body.comment)) {
        return reply.type("application/json").code(400).send("");
    }

    const selectCommentSql = fastify.mysql.format('SELECT * FROM comments WHERE id=?', [request.params.item_id]);
    const [selectedCommentRows] = await conn.query(selectCommentSql);
    fastify.log.info(selectCommentSql);

    const username = request.cookies.username;
    const commentData = { comment: request.body.comment, username };

    if (isEmpty(selectedCommentRows[0])) {
        // first comment
        commentData['comment_id'] = 1;
        const insertCommentSql = fastify.mysql.format('INSERT INTO comments (id, comment_001) VALUES (?, ?)',
            [request.params.item_id, JSON.stringify(commentData)]);
        await conn.query(insertCommentSql);
        fastify.log.info(insertCommentSql);
        await conn.commit();
    } else {
        // There are some comments already.
        for (let [key, value] of Object.entries(selectedCommentRows[0])) {
            if (isEmpty((value))) {
                commentData['comment_id'] = parseInt(key.slice(-3));
                const updateCommentSql = fastify.mysql.format('UPDATE comments SET ' + key + '=? WHERE id=?',
                    [JSON.stringify(commentData), request.params.item_id]);
                await conn.query(updateCommentSql);
                fastify.log.info(updateCommentSql);
                await conn.commit();
                break;
            }
        }
    }

    conn.release();
    commentData['item_id'] = request.params.item_id;
    return reply.code(201).send(commentData);
});

fastify.get('/items/:item_id/comments', async (request, reply) => {
    const item_id = parseInt(request.params.item_id);

    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    const selectCommentSql = fastify.mysql.format('SELECT * FROM comments WHERE id=?', [item_id]);
    const [selectedCommentRows] = await conn.query(selectCommentSql);
    fastify.log.info(selectCommentSql);

    const resData = {item_id};
    const itemCommentsExceptNull = [];
    if (isEmpty(selectedCommentRows[0])) {
        resData['comments'] = [];
        return reply.send(resData);
    }

    for (let [key, value] of Object.entries(selectedCommentRows[0])) {
        if (key === 'id' || isEmpty(value)) {
            continue;
        }
        itemCommentsExceptNull.push(value);
    }

    resData['comments'] = itemCommentsExceptNull;
    conn.release();
    return reply.send(resData);
});

fastify.delete('/items/:item_id/comments/:comment_id', { preHandler: loginRequired }, async (request, reply) => {
    const item_id = parseInt(request.params.item_id);
    const comment_id = parseInt(request.params.comment_id);

    const conn = await getDbConn();
    const selectItemSql = fastify.mysql.format('SELECT * FROM items WHERE id=?', [item_id]);
    const [selectedItemRows] = await conn.query(selectItemSql);
    fastify.log.info(selectItemSql);

    if (selectedItemRows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }
    const selectCommentSql = fastify.mysql.format('SELECT * FROM comments WHERE id=?', [item_id]);
    const [selectedCommentRows] = await conn.query(selectCommentSql);
    fastify.log.info(selectCommentSql);

    if (isEmpty(selectedCommentRows[0])) {
        return reply.type("application/json").code(404).send("");
    }

    for (let [key, value] of Object.entries(selectedCommentRows[0])) {
        if (key === 'id' || isEmpty(value)) {
            continue;
        }
        const comment = value;
        const username = request.cookies.username;

        if (parseInt(comment['comment_id']) !== comment_id) {
            continue;
        }

        if (comment['username'] !== username) {
            return reply.type("application/json").code(403).send("");
        }

        const updateCommentSql = fastify.mysql.format('UPDATE comments SET ' + key + '=NULL WHERE id=?',
            [item_id]);
        await conn.query(updateCommentSql);
        fastify.log.info(updateCommentSql);
        await conn.commit();
        conn.release();
        return reply.code(204).send("");
    }

    conn.release();
    return reply.type("application/json").code(404).send("");
});


fastify.setErrorHandler(async (error, request, reply) => {
    if (error.code === 'FST_ERR_CTP_INVALID_MEDIA_TYPE') {
        return reply.type("application/json").code(400).send("");
    }
});

fastify.post('/users/:username/icon', { preHandler: loginRequired }, async (request, reply) => {
    const conn = await getDbConn();
    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [request.params.username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (request.params.username !== request.cookies.username) {
        conn.release();
        return reply.type("application/json").code(403).send("");
    }

    if (rows.length === 0) {
        conn.release();
        return reply.type("application/json").code(404).send("");
    }

    const selectIconSql = fastify.mysql.format('SELECT * FROM icon WHERE user_id = ?', [rows[0]['id']]);
    const [iconRows] = await conn.query(selectIconSql);
    fastify.log.info(selectIconSql);

    if (!isEmpty(iconRows[0])) {
        conn.release();
        return reply.type("application/json").code(409).send("");
    }

    const data = [];
    const mp = request.multipart((field, file, filename, encoding, mimetype) => {
        file.on('data', chunk => {
            data.push(chunk)
        });

        if (isEmpty(field)) {
            return reply.type("application/json").code(400).send("");
        }
    }, async err => {
        const buf = Buffer.concat(data);

        const insertIconSql = fastify.mysql.format('INSERT INTO icon (user_id, icon) VALUES (? ,?)',
            [rows[0]['id'], buf.toString('base64')]);

        const [iconRows] = await conn.query(insertIconSql);
        await conn.commit();
        fastify.log.info(insertIconSql);
        conn.release();

        return reply.code(201).send("");
    });

    mp.on('field', function (key, value) {
        // do nothing currently
    });
});

fastify.get('/users/:username/icon', async (request, reply) => {
    const conn = await getDbConn();
    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [request.params.username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (rows.length === 0) {
        return reply.type("application/json").code(404).send("");
    }

    const selectIconSql = fastify.mysql.format('SELECT * FROM icon WHERE user_id = ?', [rows[0]['id']]);
    const [iconRows] = await conn.query(selectIconSql);

    if (isEmpty(iconRows[0])) {
        // return reply.sendfile('public/img/default_user_icon.png');
        const stream = fs.createReadStream('public/img/default_user_icon.png');
        return reply.type('image/png').send(stream);
    }

    const img = Buffer.from(iconRows[0]['icon'].toString('binary'), 'base64');
    return reply.type("image/png").send(img);
});


fastify.post('/signin', async (request, reply) => {
    if (isEmpty(request.body.username) || isEmpty(request.body.password)) {
        return reply.type("application/json").code(400).send("");
    }

    const username = request.body.username;
    const password = request.body.password;

    const conn = await getDbConn();

    const sql = fastify.mysql.format('SELECT * FROM users WHERE username = ?', [username]);
    const [rows] = await conn.query(sql);
    fastify.log.info(sql);

    if (rows.length === 0) {
        return reply.type("application/json").code(401).send("");
    }

    const salt = rows[0].salt;
    const passwordHash = rows[0].password_hash;

    if (await getPasswordhash(salt, password) !== passwordHash) {
        return reply.type("application/json").code(401).send("");
    }

    reply.setCookie('user_id', rows[0].id, {
        path: '/',
    }).setCookie('username', username, {
        path: '/',
    });

    conn.release();
    return reply.send({ username });
});

fastify.get('/signout',{preHandler: loginRequired}, async (request, reply) => {
    if (isEmpty(request.cookies.user_id) || isEmpty(request.cookies.username)) {
        return reply.code(401).send("");
    }

    return reply.setCookie("user_id", "", {
        path: "/",
        expires: new Date(0),
    }).setCookie("username", "", {
        path: "/",
        expires: new Date(0),
    }).code(204).send();
});

fastify.get("/initialize", async (_request, reply) => {
    await child_process.execFile("../common/db/init.sh");
    return reply.code(200).send("");
});

//// Bootstrap

fastify.listen(5000, '0.0.0.0', err => {
    if (err) {
        fastify.log.error(err);
        process.exit(1)
    };
});
