const faker = require('faker');
const DateGenerator = require('random-date-generator');
const moment = require('moment')
const _ = require('lodash')

const fs = require('fs');
const users_raw = fs.readFileSync('../user_pass/merged_user_pass.json', { encoding: 'utf-8' });
const users = JSON.parse(users_raw);
const usernames = users.map(user => user.username);

const getRand = (min, max) => {
    return Math.floor(Math.random() * (max - min + 1) + min);
};

const startDate = new Date(2000, 1, 1);
const endDate = new Date(2019, 8, 30);
const d = DateGenerator.getRandomDateInRange(startDate, endDate);

[...Array(200000)].forEach((unused_num, i) => {
    const user_id = getRand(1, 10000)
    const title = faker.lorem.words();
    const body = faker.lorem.paragraphs();
    const d = DateGenerator.getRandomDateInRange(startDate, endDate);
    const date = moment(d).format('YYYY-MM-DD HH:mm:ss');

    let like = "NULL";
    if (Math.random() <= 0.5) {
        like = `'${_.sampleSize(usernames, getRand(0, 100)).join(',')}'`;
    } 

    console.log("INSERT INTO `items` VALUES ");
    console.log(`(${i+1}, '${title}', '${body}',${user_id},${like},'${date}','${date}');`)
});
