FROM node:10

RUN npm install -g newman

ARG project_dir=/test/
WORKDIR $project_dir

ADD apitest_collection.json $project_dir

CMD ["/bin/sh", "-c", "sleep 36000"] ' #TODO: 現状コンテナ内に入って`newman run hogefuga.json`を叩いているのであとで修正する
