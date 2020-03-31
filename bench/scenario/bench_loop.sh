#!/bin/bash

SLEEP_TIME=3
INITIALIZE_TIMEOUT_SEC=20
REDIS_HOST='10.0.0.4'
DB_HOST='10.74.80.3'

# DB格納用
STATUS_SUCCESS=0
STATUS_FAIL=1

# ベンチマーク用
VUS=16
BENCH_DURATION_SEC=60

# アイコン準備
if [ -e 'icons.tgz' ]; then
    echo "Start to decompress icon files..."
    tar zxf icons.tgz
    echo "Done: Decompress icon files..."
fi

while true
do
    message=`redis-cli -h ${REDIS_HOST} rpop benchqueue`

    if [ -n "$message" ]; then
        # $messageはCSV形式で"1,0.0.0.0"を持つ、
        # 1つ目のFieldがチームID、2つ目がベンチ対象IP
        team_id=`echo $message | awk '{split($1, a, ","); print a[1]}'`
        target_host=`echo $message | awk '{split($1, a, ","); print a[2]}'`

        echo "Going to benchmark to ${target_host} tuned by team_id:  ${team_id} ..."

        ### STEP1 initialize
        message="STEP1: Start to Initialize DB..."
        command_log="${message}\n"
        echo "${message}"
        curl --fail --max-time ${INITIALIZE_TIMEOUT_SEC} "http://${target_host}/initialize"

        if [ $? -ne 0 ]; then
            message="STEP1: Failed to initialize DB for ${target_host}\nEnd of benchmark..."
            echo ${message}
            command_log+="${message}\n"

            mysql -uisucon -pisucon dashboard -h ${DB_HOST} \
              -e "INSERT INTO results (team_id, score, status, command_result, finished_date) VALUES (${team_id}, 0, ${STATUS_FAIL}, '${command_log}', now());"

            sleep 5
            continue
        else
            message="STEP1: Succeeded in initialization of DB for ${target_host}"
            echo ${message}
            command_log+="${message}\n"
        fi

        ### STEP2 アプリケーション互換性チェック

        message="STEP2: Start to Check App Compatibility..."
        echo ${message}
        command_log+="\n${message}\n"

        newman run postman.json --color off --global-var "url=http://${target_host}" --color off  --reporters json --reporter-json-export outputfile.json > /dev/null

        if [ $? -ne 0 ]; then
            message="STEP2: Failed to check compatibility for ${target_host}"
            echo ${message}
            command_log+="${message}\n"

            newman_log=`cat outputfile.json | jq -c -r '.run.failures[] | [.source.name, .error.message]' | sed -e "s/'/ /g"`
            command_log+="${newman_log}\n"

            mysql -uisucon -pisucon dashboard -h ${DB_HOST} \
              -e "INSERT INTO results (team_id, score, status, command_result, finished_date) VALUES (${team_id}, 0, ${STATUS_FAIL}, '${command_log}', now());"

            sleep 5
            continue
        else
            message="STEP2: Succeeded in checking compatibility for ${target_host}"
            echo ${message}
            command_log+="${message}\n"
        fi

        ### STEP3 ベンチマーク60秒
        message="STEP3: Start to perform a benchmark..."
        echo ${message}
        command_log+="\n${message}"

        filename=`date "+%Y%m%d_%H%M%S"`
        k6 run --out json=/var/tmp/${filename}.json --duration ${BENCH_DURATION_SEC}s --vus ${VUS}  bench_scenario.js

        if [ $? -ne 0 ]; then
            message="STEP3: Failed to perform a benchmark for ${target_host}"
            echo ${message}
            command_log+="${message}\n"

            mysql -uisucon -pisucon dashboard -h ${DB_HOST} \
              -e "INSERT INTO results (team_id, score, status, command_result, finished_date) VALUES (${team_id}, 0, ${STATUS_FAIL}, '${command_log}', now());"

            sleep 5
            continue
        else
            message="STEP3: Succeeded in benchmark for ${target_host}. Caluculating score..."
            echo ${message}
            command_log+="\n${message}\n"

            # resultsには配列が入る 0 => スコア, 1=> SUCCESS数 =>, 2=> FAILURE数
            results=( `node result_parser.js /var/tmp/${filename}.json "http://${target_host}" | head -n 1` )
            score=${results[0]}
            num_success=${results[1]}
            num_failure=${results[2]}

            message=`cowsay -W 200 "Congratulations! Score: ${score}, Success requests: ${num_success}, Failed requests: ${num_failure}"`
            command_log+="\n${message}"

            mysql -uisucon -pisucon dashboard -h ${DB_HOST} \
              -e "INSERT INTO results (team_id, score, status, command_result, finished_date) VALUES (${team_id}, ${results[0]}, ${STATUS_SUCCESS}, '${command_log}', now());"

            # todo: firebaseへ投稿
        fi
    else
        echo "Seems there's no benchmark request in the queue. Let's just wait for ${SLEEP_TIME} secs..."
        sleep ${SLEEP_TIME}
    fi
done

