FROM golang:1.12.5

ARG project_dir=/app/
WORKDIR $project_dir

RUN go get github.com/zenazn/goji
RUN go get github.com/zenazn/goji/web
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/sessions

ADD app.go $project_dir
ADD endpoint.sh $project_dir

COPY ./utils/ ${project_dir}utils/


CMD ["./endpoint.sh"]

EXPOSE 5000