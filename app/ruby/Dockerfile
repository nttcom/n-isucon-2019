FROM ruby:2.6.3

ARG project_dir=/app/
WORKDIR $project_dir

COPY . $project_dir

RUN gem install bundler -v 2.0.1
RUN bundle install --force --path .bundle

EXPOSE 9292

CMD ["bundle", "exec", "rackup", "-o", "0.0.0.0"]
