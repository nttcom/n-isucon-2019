// Check the result file is specified or not.
let resultfile;
if (!process.argv[2]) {
    console.error('第1引数に解析対象のファイルを指定してください');
    process.exit(1);
} else {
    resultfile = process.argv[2];
}

// Check the target URL is specified or not. This value is used for notification.
let targetUrl;
if (!process.argv[3]) {
    console.error('第2引数にテスト対象URLを指定してください');
    process.exit(1);
} else {
    targetUrl = process.argv[3];
}

// This array is used for to read RAW result from k6.
const rawResults = [];

// Main process triggered after finished parsing RAW result.
const parseResults = () => {
    const httpReqDurationResults = rawResults.filter(result => {
        return result.metric === 'checks' && result.type === 'Point'
    });

    const getResults = {};
    const postResults = {};
    const patchResults = {};
    const delResults = {};

    createResultAggregator('GET')(httpReqDurationResults, getResults);
    createResultAggregator('POST')(httpReqDurationResults, postResults);
    createResultAggregator('PATCH')(httpReqDurationResults, patchResults);
    createResultAggregator('DEL')(httpReqDurationResults, delResults);

    preProcessBeforeCalcScore([getResults, postResults, patchResults, delResults]);
    const score = calcResults([getResults], [postResults, patchResults, delResults]);

    totalSuccess = getResults['1'] || 0 + postResults['1'] || 0  + patchResults['1'] || 0  + delResults['1'] || 0;
    totalFailure = getResults['0'] || 0 + postResults['0'] || 0  + patchResults['0'] || 0  + delResults['0'] || 0;

    console.log(`${score} ${totalSuccess} ${totalFailure}`);
    console.log('GET   ->', 'FAIL:', getResults['0'] || 0, 'SUCCESS:', getResults['1'] || 0);
    console.log('POST  ->', 'FAIL:', postResults['0'] || 0, 'SUCCESS:', postResults['1'] || 0);
    console.log('PATCH ->', 'FAIL:', patchResults['0'] || 0, 'SUCCESS:', patchResults['1'] || 0);
    console.log('DEL   ->', 'FAIL:', delResults['0'] || 0, 'SUCCESS:', delResults['1'] || 0);

    // const firebaseBody = constructBodyForFirebase(getResults, listResults, postResults, putResults, deleteResults, score);
    // postDataForDashboard(firebaseBody);
};

const constructBodyForFirebase = (getResults, listResults, postResults, putResults, deleteResults, score) => {
    const intLevel = parseInt(testLevel.slice(-1), 10);
    const unixTimestampMsec = new Date() / 1000;
    const body = {
        score: score,
        level: intLevel,
        results: {
            get: [getResults['1'], getResults['0']],
            list: [listResults['1'], listResults['0']],
            post: [postResults['1'], postResults['0']],
            put: [putResults['1'], putResults['0']],
            delete: [deleteResults['1'], deleteResults['0']],
        },
        timestamp: unixTimestampMsec
    };
    if (score > THRESHOLD) {
        body.pass = true;
    } else {
        body.pass = false;
    }
    return body;
};

// todo: store data for dashboard
const postDataForDashboard = (body) => {
    const teamName = require('url').parse(targetUrl).host.split('.')[0];
    const strBody = JSON.stringify(body);
    const options = {
        hostname: 'iaas-ho-01-05726456.firebaseio.com',
        port: 443,
        path: `/iaas/teams/${teamName}.json`,
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(strBody)
        }
    };
    const req = require('https').request(options, res => {
        if (res.statusCode === 200) {
            console.log('Succeeded in posting data to Firebase.');
        } else {
            console.error('Failed to post data to Firebase. Status code is', res.statusCode);
        }
    });
    req.on('error', e => {
        console.error('Failed to post data to Firebase.')
        console.error(e);
    });
    req.write(strBody);
    req.end();

};

// Utility function to create function to aggregate raw result easily
const createResultAggregator = httpMethod => {
    return (array, typeResults) => {
        array.filter(result => {
            /** 
             * result.data.tags.groupの中に、
             *   ::既存の記事::コメント::POST /items/:id/comments
             * という文字列があるので、この中からhttp method名を取得する。上記の例だと POST
             */
            const httpMethodName = result.data.tags.group.split('::').slice(-1)[0].split(' ')[0];
            return httpMethodName === httpMethod
        }).forEach(result => {
            typeResults[result.data.value] = (typeResults[result.data.value] || 0) + 1;
        });
    };
};

// This value is used for calculate score, when the HTTP request fails
const FAIL_WEIGHT = 20;
const WRIGHT_WEIGHT = 10;
// Judge the test is passed or failed.
const THRESHOLD = 5000;

const calcResults = (readResults, writeResults) => {
    let score = 0;
    readResults.forEach(result => {
        score += (result['1'] - result['0'] * FAIL_WEIGHT)
    });
    writeResults.forEach(result => {
        score += (result['1'] * WRIGHT_WEIGHT - result['0'] * FAIL_WEIGHT)
    });
    return score;
}

const preProcessBeforeCalcScore = resultsList => {
    resultsList.forEach(results => {
        initwithZero(results);
    });
}

const initwithZero = (results) => {
    if (!results['0']) {
        results['0'] = 0;
    }
    if (!results['1']) {
        results['1'] = 0;
    }
}

// Read results from files
const lineReader = require('readline').createInterface({
    input: require('fs').createReadStream(resultfile)
})

lineReader.on('line', line => {
    const result = JSON.parse(line);
    rawResults.push(result);
});

lineReader.on('close', () => {
    parseResults();
});