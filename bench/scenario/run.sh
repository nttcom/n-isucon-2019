#!/bin/bash

# ベンチマーク用
VUS=1
BENCH_DURATION_SEC=60
INITIALIZE_TIMEOUT_SEC=15

# アイコン準備
if [ -e 'icons.tgz' ]; then
    echo "Start to decompress icon files..."
    tar zxf icons.tgz
    echo "Done: Decompress icon files..."
fi

# $messageはCSV形式で"1,0.0.0.0"を持つ、
# 1つ目のFieldがチームID、2つ目がベンチ対象IP
target_host="192.168.33.11"

echo "Going to benchmark to ${target_host}"

### STEP1 initialize
message="STEP1: Start to Initialize DB..."
echo "${message}"
curl --fail --max-time ${INITIALIZE_TIMEOUT_SEC} "http://${target_host}/initialize"

if [ $? -ne 0 ]; then
    message="STEP1: Failed to initialize DB for ${target_host}\nEnd of benchmark..."
    echo ${message}

    exit 1
else
    message="STEP1: Succeeded in initialization of DB for ${target_host}"
    echo ${message}
fi

### STEP2 アプリケーション互換性チェック

message="STEP2: Start to Check App Compatibility..."
echo ${message}

newman run postman.json --global-var "url=http://${target_host}"

if [ $? -ne 0 ]; then
    message="STEP2: Failed to check compatibility for ${target_host}"
    echo ${message}

    exit 1
    continue
else
    message="STEP2: Succeeded in checking compatibility for ${target_host}"
    echo ${message}
fi

### STEP3 ベンチマーク60秒
message="STEP3: Start to perform a benchmark..."
echo ${message}

filename=`date "+%Y%m%d_%H%M%S"`
k6 run --out json=/var/tmp/${filename}.json --duration ${BENCH_DURATION_SEC}s --vus ${VUS}  bench_scenario.js

if [ $? -ne 0 ]; then
    message="STEP3: Failed to perform a benchmark for ${target_host}"
    echo ${message}

    exit 1
else
    message="STEP3: Succeeded in benchmark for ${target_host}. Caluculating score..."
    echo ${message}

    # resultsには配列が入る 0 => スコア, 1=> SUCCESS数 =>, 2=> FAILURE数
    results=( `node result_parser.js /var/tmp/${filename}.json "http://${target_host}" | head -n 1` )
    score=${results[0]}
    num_success=${results[1]}
    num_failure=${results[2]}
    cowsay -W 200 "Congratulations! Score: ${score}, Success requests: ${num_success}, Failed requests: ${num_failure}"
fi
