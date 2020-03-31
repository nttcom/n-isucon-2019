const fs = require('fs');
const axios = require('axios');

const users_raw = fs.readFileSync('../../data/merged_user_pass.json', { encoding: 'utf-8' });
const users = JSON.parse(users_raw);

const apiUrl = 'http://localhost:5000';

const sleep = msec => new Promise(resolve => setTimeout(resolve, msec));

(async () => {
    const initRes = await axios.get(`${apiUrl}/initialize`);
    console.log(initRes.status);

    let i = 1;
    for (let user of users) {
        const res = await axios.post(`${apiUrl}/users`, user);
        console.log(i++, res.status);
    }
})();
