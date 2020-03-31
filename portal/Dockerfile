FROM ruby:2.6.3

ARG project_dir=/app/
WORKDIR $project_dir

ADD Gemfile $project_dir
ADD Gemfile.lock $project_dir

RUN gem install bundler -v 2.0.1
RUN bundle install --path .bundle

COPY . $project_dir

CMD ["bundle", "exec", "rackup", "-o", "0.0.0.0"]

EXPOSE 9292
