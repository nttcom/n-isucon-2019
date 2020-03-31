FROM python:3.7.1

ARG project_dir=/app/
WORKDIR $project_dir

ADD requirements.txt $project_dir
RUN pip install -r requirements.txt

ADD app.py $project_dir
ADD endpoint.sh $project_dir

COPY ./utils/ /app/utils/
COPY ../common/ /app/common/

CMD ["./endpoint.sh"]

EXPOSE 5000
