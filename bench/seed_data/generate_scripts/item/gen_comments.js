const faker = require('faker');
const _ = require('lodash')

const fs = require('fs');
const users_raw = fs.readFileSync('../user_pass/merged_user_pass.json', { encoding: 'utf-8' });
const users = JSON.parse(users_raw);
const usernames = users.map(user => user.username);

const getRand = (min, max) => {
    return Math.floor(Math.random() * (max - min + 1) + min);
};

[...Array(200000)].forEach((unused_num, i) => {

    if (Math.random() <= 0.3) {
        return;
    } 

    const comment_users = _.sampleSize(usernames, getRand(1, 50));


    comment_users.forEach((comment_user, j) => {
        comment_obj = {}
        comment_obj.comment = faker.lorem.words();
        comment_obj.username = comment_user;
        comment_obj.comment_id = j+1;

        if (j === 0) {
            console.log("INSERT INTO comments (id, comment_001) VALUES (" + (i+1) + ", '" + JSON.stringify(comment_obj) + "');");
        } else {
            console.log("UPDATE comments SET comment_" + String(j+1).padStart(3,0) + "='" + JSON.stringify(comment_obj) + "'");
            console.log("WHERE id=" + (i+1) + ";");
        }
    });

});
